package plotter

import (
	"fmt"
	serial "github.com/tarm/goserial"
	"math"
	"time"
)

// Output the coordinates to the screen
func OutputCoords(plotCoords chan Coordinate) {

	for {
		coord, chanOpen := <-plotCoords
		if !chanOpen {
			break
		}

		fmt.Println(coord)
	}

	fmt.Println("Done plotting")
}

// Takes in coordinates and outputs stepData
func GenerateStepsLinear(plotCoords chan Coordinate, stepData chan byte) {

	defer close(stepData)

	polarSystem := PolarSystemFromSettings()
	previousPolarPos := PolarCoordinate{LeftDist: Settings.StartingLeftDist_MM, RightDist: Settings.StartingRightDist_MM}
	startingLocation := previousPolarPos.ToCoord(polarSystem)

	//fmt.Println("Start Location", startingLocation, "Initial Polar", previousPolarPos)

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

		//fmt.Println("Reading", curTarget)
		targetVector := origin.Minus(curTarget)

		actualDistance := targetVector.Len()
		idealTime := actualDistance / Settings.MaxSpeed_MM_S
		numberOfSlices := math.Ceil(idealTime / (Settings.TimeSlice_US / 1000000))

		for slice := 1.0; slice <= numberOfSlices; slice++ {

			percentageAlongLine := slice / numberOfSlices
			sliceTarget := origin.Add(targetVector.Scaled(percentageAlongLine))

			polarSliceTarget := sliceTarget.ToPolar(polarSystem)

			sliceSteps := previousPolarPos.Minus(polarSliceTarget).Scaled(1 / Settings.StepSize_MM)
			sliceSteps = sliceSteps.Ceil() //.Clamp(32, -32)

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
			//fmt.Println("Len of stepData", len(stepData), encodedSteps)
		}
		origin = previousActualPos
		//fmt.Println("NEXT STEP --------------------------------------------")
	}
	fmt.Println("Done generating steps")
}

// Generates steps from plotCoords and sends those steps to the serial port
func RenderCoords(plotCoords chan Coordinate, generateSteps func(chan Coordinate, chan byte)) {

	// convert coordinates
	stepData := make(chan byte, 1024)
	go generateSteps(plotCoords, stepData)

	stepCount := 0
	for {
		_, chanOpen := <-stepData
		if !chanOpen {
			break
		}

		stepCount++
	}
	fmt.Println("Steps ", stepCount/2, "Time", time.Duration(float64(stepCount)*0.5*Settings.TimeSlice_US)*time.Microsecond)
}

// Sends the given stepData to the stepper driver
func RunSteps(stepData []byte, loop bool) {
	fmt.Println("Opening com port")
	c := &serial.Config{Name: "/dev/ttyAMA0", Baud: 57600}
	s, err := serial.OpenPort(c)
	if err != nil {
		panic(err)
	}
	defer s.Close()

	writeData := make([]byte, 128)
	readData := make([]byte, 1)

	stepPosition := 0
	previousSend := time.Now()

	for {
		n, err := s.Read(readData)
		if err != nil {
			panic(err)
		}
		if n != 1 {
			panic("Expected only 1 byte on s.Read")
		}

		dataToWrite := int(readData[0])
		for i := 0; i < dataToWrite; i += 2 {

			if stepPosition == len(stepData) {
				if loop {
					stepPosition = 0
				} else {
					writeData[i] = 0
					writeData[i+1] = 0
					continue
				}
			}

			writeData[i] = stepData[stepPosition]
			stepPosition++
			writeData[i+1] = stepData[stepPosition]
			stepPosition++
		}

		curTime := time.Now()
		fmt.Println("Sending data after", curTime.Sub(previousSend))
		previousSend = curTime

		sendTime := time.Now()
		s.Write(writeData)
		fmt.Println("Send took ", time.Now().Sub(sendTime))

		if stepPosition == len(stepData) && !loop {
			break
		}
	}
}
