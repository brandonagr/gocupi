package plotter

import (
	"fmt"
	serial "github.com/tarm/goserial"
	"io"
	"math"
	"os"
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

// // Takes in coordinates and outputs stepData
// func GenerateSteps(plotCoords <-chan Coordinate, stepData chan<- byte, useCubicSmooth bool) {

// 	defer close(stepData)

// 	polarSystem := PolarSystemFromSettings()
// 	previousPolarPos := PolarCoordinate{LeftDist: Settings.StartingLeftDist_MM, RightDist: Settings.StartingRightDist_MM}
// 	startingLocation := previousPolarPos.ToCoord(polarSystem)

// 	fmt.Println("Start Location", startingLocation, "Initial Polar", previousPolarPos)
// 	fmt.Println("MinY", Settings.MinVertical_MM, "MaxY", Settings.MaxVertical_MM)

// 	// setup 0,0 as the initial location of the plot head
// 	polarSystem.XOffset = startingLocation.X
// 	polarSystem.YOffset = startingLocation.Y

// 	origin := Coordinate{}
// 	previousActualPos := origin

// 	for {
// 		curTarget, chanOpen := <-plotCoords
// 		if !chanOpen {
// 			break
// 		}

// 		targetVector := curTarget.Minus(origin)

// 		actualDistance := targetVector.Len()
// 		idealTime := actualDistance / Settings.MaxSpeed_MM_S
// 		numberOfSlices := math.Ceil(idealTime / (Settings.TimeSlice_US / 1000000))

// 		for slice := 1.0; slice <= numberOfSlices; slice++ {

// 			percentageAlongLine := slice / numberOfSlices
// 			if useCubicSmooth && numberOfSlices > 3 {
// 				percentageAlongLine = CubicSmooth(percentageAlongLine)
// 			}

// 			sliceTarget := origin.Add(targetVector.Scaled(percentageAlongLine))
// 			polarSliceTarget := sliceTarget.ToPolar(polarSystem)

// 			// calc integer number of steps that will be made this time slice
// 			sliceSteps := polarSliceTarget.Minus(previousPolarPos).Scaled(1 / Settings.StepSize_MM)
// 			sliceSteps = sliceSteps.Ceil().Clamp(127, -127)

// 			previousPolarPos = previousPolarPos.Add(sliceSteps.Scaled(Settings.StepSize_MM))
// 			previousActualPos = previousPolarPos.ToCoord(polarSystem)

// 			//fmt.Println("Coord target", sliceTarget, "actual", previousActualPos, "Abs Actual", previousActualPos.Add(startingLocation))
// 			//fmt.Println("Polar target", polarSliceTarget, "actual", previousPolarPos)
// 			//fmt.Println("Steps", sliceSteps, "Actual", previousActualPos)

// 			var encodedSteps byte
// 			if sliceSteps.LeftDist < 0 {
// 				encodedSteps = byte(-sliceSteps.LeftDist)
// 			} else {
// 				encodedSteps = byte(sliceSteps.LeftDist) | 0x80
// 			}
// 			stepData <- encodedSteps
// 			if sliceSteps.RightDist < 0 {
// 				encodedSteps = byte(-sliceSteps.RightDist) | 0x80
// 			} else {
// 				encodedSteps = byte(sliceSteps.RightDist)
// 			}
// 			stepData <- encodedSteps
// 		}
// 		origin = previousActualPos
// 	}
// 	fmt.Println("Done generating steps")
// }

// Takes in coordinates and outputs stepData
func GenerateStepsUsingInterpolation(plotCoords <-chan Coordinate, stepData chan<- int8, useCubicSmooth bool) {

	defer close(stepData)

	polarSystem := PolarSystemFromSettings()
	previousPolarPos := PolarCoordinate{LeftDist: Settings.StartingLeftDist_MM, RightDist: Settings.StartingRightDist_MM}
	startingLocation := previousPolarPos.ToCoord(polarSystem)

	fmt.Println("Start Location", startingLocation, "Initial Polar", previousPolarPos)
	fmt.Println("MinY", Settings.MinVertical_MM, "MaxY", Settings.MaxVertical_MM)

	// setup 0,0 as the initial location of the plot head
	polarSystem.XOffset = startingLocation.X
	polarSystem.YOffset = startingLocation.Y

	interp := new(InterpolaterData)

	//origin := Coordinate{}

	origin, chanOpen := <-plotCoords
	if !chanOpen {
		return
	}
	target, chanOpen := <-plotCoords
	if !chanOpen {
		return
	}
	var anotherTarget bool = true

	for anotherTarget {
		nextTarget, chanOpen := <-plotCoords
		if !chanOpen {
			anotherTarget = false
			nextTarget = target
		}

		interp.Setup(origin, target, nextTarget)

		for slice := 1.0; slice <= interp.Slices(); slice++ {

			sliceTarget := interp.Position(slice)
			polarSliceTarget := sliceTarget.ToPolar(polarSystem)

			// calc integer number of steps * 32 that will be made this time slice
			sliceSteps := polarSliceTarget.Minus(previousPolarPos).Scaled(32.0/Settings.StepSize_MM).Ceil().Clamp(127, -127)

			//fmt.Println("Cur slice target:", sliceTarget, "Previous", previousPolarPos, "Polar target", polarSliceTarget)
			//fmt.Println("Steps", sliceSteps)
			// if sliceSteps.LeftDist == 127 || sliceSteps.LeftDist == -127 {

			// 	fmt.Println(interp.Position(slice - 1))
			// 	fmt.Println(interp.Position(slice))
			// 	fmt.Println("Origin:", origin, "Target", target, "NextTarget", nextTarget)
			// 	// fmt.Println("Cur slice target:", sliceTarget, "Previous", previousPolarPos, "Polar target", polarSliceTarget)
			// 	// fmt.Println("Steps", sliceSteps)
			// 	interp.WriteData()
			// 	panic("Exceeded speed")
			// }

			previousPolarPos = previousPolarPos.Add(sliceSteps.Scaled(Settings.StepSize_MM / 32.0))

			//fmt.Println("Coord target", sliceTarget, "actual", previousActualPos, "Abs Actual", previousActualPos.Add(startingLocation))
			//fmt.Println("Polar target", polarSliceTarget, "actual", previousPolarPos)
			//fmt.Println("Steps", sliceSteps, "Polar target", polarSliceTarget, "Polar actual", previousPolarPos)

			stepData <- int8(sliceSteps.LeftDist)
			stepData <- int8(-sliceSteps.RightDist)
		}
		origin = previousPolarPos.ToCoord(polarSystem)
		target = nextTarget
	}
	fmt.Println("Done generating steps")
}

// Count steps
func CountSteps(stepData <-chan int8) {

	sliceCount := 0
	for _ = range stepData {

		sliceCount++
	}
	fmt.Println("Steps ", sliceCount/2, "Time", time.Duration(float64(sliceCount)*0.5*Settings.TimeSlice_US)*time.Microsecond)
}

// Sends the given stepData to a file
func WriteStepsToFile(stepData <-chan int8) {

	file, err := os.OpenFile("stepData.txt", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var byteDataL, byteDataR int8
	size := 0
	for stepDataOpen := true; stepDataOpen; {

		byteDataL, stepDataOpen = <-stepData
		byteDataR, stepDataOpen = <-stepData

		io.WriteString(file, fmt.Sprintln(byteDataL, byteDataR))
		size++
		if size > 10000 {
			return
		}
	}
}

// Sends the given stepData to the stepper driver
func WriteStepsToSerial(stepData <-chan int8) {
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
	var totalSends int = 0
	var byteData int8 = 0

	for stepDataOpen := true; stepDataOpen; {

		// wait for next data request
		n, err := s.Read(readData)
		if err != nil {
			panic(err)
		}
		if n != 1 {
			panic(err)
		}

		dataToWrite := int(readData[0])
		for i := 0; i < dataToWrite; i += 2 {

			byteData, stepDataOpen = <-stepData
			writeData[i] = byte(byteData)

			byteData, stepDataOpen = <-stepData
			writeData[i+1] = byte(byteData)
		}

		totalSends++
		if totalSends >= 100 {
			curTime := time.Now()

			fmt.Println("Sent 100 messages after", curTime.Sub(previousSend))
			totalSends = 0

			previousSend = curTime
		}

		s.Write(writeData)
	}
}

// Used to manually adjust length of each step
func PerformManualAlignment() {

	for {

		var side string
		var distance float64
		fmt.Print("Input L/R DIST:")
		if _, err := fmt.Scanln(&side, &distance); err != nil {
			return
		}

		side = strings.ToLower(side)
		fmt.Println("Moving ", side, distance)

		alignStepData := make(chan int8, 1024)
		go WriteStepsToSerial(alignStepData)

		interp := new(InterpolaterData)
		interp.Setup(Coordinate{}, Coordinate{distance, 0}, Coordinate{})
		position := 0.0

		for slice := 1.0; slice <= interp.Slices(); slice++ {

			sliceTarget := interp.Position(slice)

			// calc integer number of steps that will be made this time slice
			sliceSteps := math.Ceil((sliceTarget.X - position) * (32.0 / Settings.StepSize_MM))
			position = position + sliceSteps*(Settings.StepSize_MM/32.0)

			if side == "l" {
				alignStepData <- int8(sliceSteps)
				alignStepData <- 0
			} else {
				alignStepData <- 0
				alignStepData <- int8(-sliceSteps)
			}
		}

		close(alignStepData)
	}
}
