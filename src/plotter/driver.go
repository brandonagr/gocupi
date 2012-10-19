package plotter

import (
	"fmt"
	serial "github.com/tarm/goserial"
	"math"
	"strings"
	"time"
)

// Output the coordinates to the screen
func OutputCoords(plotCoords <-chan Coordinate) {

	for coord := range plotCoords {

		fmt.Println(coord)
	}

	fmt.Println("Done plotting")
}

// Takes in coordinates and outputs stepData
func GenerateStepsLinear(plotCoords <-chan Coordinate, stepData chan<- byte) {

	defer close(stepData)

	polarSystem := PolarSystemFromSettings()
	previousPolarPos := PolarCoordinate{LeftDist: Settings.StartingLeftDist_MM, RightDist: Settings.StartingRightDist_MM}
	startingLocation := previousPolarPos.ToCoord(polarSystem)

	fmt.Println("Start Location", startingLocation, "Initial Polar", previousPolarPos)
	fmt.Println("MinY", Settings.MinVertical_MM, "MaxY", Settings.MaxVertical_MM)

	// setup 0,0 as the initial location of the plot head
	polarSystem.XOffset = startingLocation.X
	polarSystem.YOffset = startingLocation.Y

	origin := Coordinate{}
	previousActualPos := origin

	for {
		curTarget, chanOpen := <-plotCoords
		if !chanOpen {
			break
		}

		targetVector := origin.Minus(curTarget)

		actualDistance := targetVector.Len()
		idealTime := actualDistance / Settings.MaxSpeed_MM_S
		numberOfSlices := math.Ceil(idealTime / (Settings.TimeSlice_US / 1000000))

		for slice := 1.0; slice <= numberOfSlices; slice++ {

			percentageAlongLine := slice / numberOfSlices
			sliceTarget := origin.Add(targetVector.Scaled(percentageAlongLine))
			polarSliceTarget := sliceTarget.ToPolar(polarSystem)

			// calc integer number of steps that will be made this time slice
			sliceSteps := previousPolarPos.Minus(polarSliceTarget).Scaled(1 / Settings.StepSize_MM)
			sliceSteps = sliceSteps.Ceil().Clamp(127, -127)

			previousPolarPos = previousPolarPos.Add(sliceSteps.Scaled(Settings.StepSize_MM))
			previousActualPos = previousPolarPos.ToCoord(polarSystem)

			//fmt.Println("Coord target", sliceTarget, "actual", previousActualPos, "Abs Actual", previousActualPos.Add(startingLocation))
			//fmt.Println("Polar target", polarSliceTarget, "actual", previousPolarPos)
			//fmt.Println("Steps", sliceSteps, "Actual", previousActualPos)

			var encodedSteps byte
			if sliceSteps.LeftDist < 0 {
				encodedSteps = byte(-sliceSteps.LeftDist)
			} else {
				encodedSteps = byte(sliceSteps.LeftDist) | 0x80
			}
			stepData <- encodedSteps
			if sliceSteps.RightDist < 0 {
				encodedSteps = byte(-sliceSteps.RightDist) | 0x80
			} else {
				encodedSteps = byte(sliceSteps.RightDist)
			}
			stepData <- encodedSteps
		}
		origin = previousActualPos
	}
	fmt.Println("Done generating steps")
}

// Generates steps from plotCoords and sends those steps to the serial port
func CountSteps(stepData <-chan byte) {

	stepCount := 0
	for _ = range stepData {

		stepCount++
	}
	fmt.Println("Steps ", stepCount/2, "Time", time.Duration(float64(stepCount)*0.5*Settings.TimeSlice_US)*time.Microsecond)
}

// Sends the given stepData to the stepper driver
func WriteStepsToSerial(stepData <-chan byte) {
	fmt.Println("Opening com port")
	c := &serial.Config{Name: "/dev/ttyAMA0", Baud: 57600}
	s, err := serial.OpenPort(c)
	if err != nil {
		panic(err)
	}
	defer s.Close()

	// buffers to use during serial communication
	writeData := make([]byte, 128)
	readData := make([]byte, 1)

	// convert plotCoords to 

	previousSend := time.Now()
	var byteData byte = 0
	for stepDataOpen := true; stepDataOpen; {
		n, err := s.Read(readData)
		if err != nil {
			panic(err)
		}
		if n != 1 {
			panic("Expected only 1 byte on s.Read")
		}

		dataToWrite := int(readData[0])
		for i := 0; i < dataToWrite; i += 2 {

			byteData, stepDataOpen = <-stepData
			writeData[i] = byteData

			byteData, stepDataOpen = <-stepData
			writeData[i+1] = byteData
		}

		curTime := time.Now()
		fmt.Println("Sending data after", curTime.Sub(previousSend))
		previousSend = curTime

		sendTime := time.Now()
		s.Write(writeData)
		fmt.Println("Send took ", time.Now().Sub(sendTime))
	}
}

// Used to manually adjust length of each step
func PerformManualAlignment() {
	alignStepData := make(chan byte, 1024)
	defer close(alignStepData)

	//go WriteStepsToSerial(alignStepData)
	go CountSteps(alignStepData)
	// go func() {
	// 	for step := range alignStepData {
	// 		fmt.Println("Step", step)
	// 	}
	// }()

	for {

		var side string
		var distance float64
		fmt.Print("Input L/R DIST:")
		if _, err := fmt.Scanln(&side, &distance); err != nil {
			return
		}

		side = strings.ToLower(side)

		idealTime := math.Abs(distance) / (Settings.MaxSpeed_MM_S * 0.5)
		numberOfSlices := math.Ceil(idealTime / (Settings.TimeSlice_US / 1000000))
		position := 0.0

		fmt.Println("Moving ", side, distance, "Slices", numberOfSlices)

		for slice := 0.0; slice <= numberOfSlices; slice++ {

			percentageAlongLine := slice / numberOfSlices
			sliceTarget := distance * percentageAlongLine

			// calc integer number of steps that will be made this time slice
			sliceSteps := (sliceTarget - position) * (1 / Settings.StepSize_MM)
			sliceSteps = math.Ceil(sliceSteps)
			position = position + sliceSteps*Settings.StepSize_MM

			if side == "l" {
				if sliceSteps < 0 {
					alignStepData <- byte(-sliceSteps)
				} else {
					alignStepData <- byte(sliceSteps) | 0x80
				}
				alignStepData <- 0
			} else {
				alignStepData <- 0
				if sliceSteps < 0 {
					alignStepData <- byte(-sliceSteps) | 0x80
				} else {
					alignStepData <- byte(sliceSteps)
				}
			}
		}
	}
}
