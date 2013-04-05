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

// Takes in coordinates and outputs stepData
func GenerateSteps(plotCoords <-chan Coordinate, stepData chan<- int8) {

	defer close(stepData)

	polarSystem := PolarSystemFromSettings()
	previousPolarPos := PolarCoordinate{LeftDist: Settings.StartingLeftDist_MM, RightDist: Settings.StartingRightDist_MM}
	startingLocation := previousPolarPos.ToCoord(polarSystem)

	fmt.Println("Start Location", startingLocation, "Initial Polar", previousPolarPos)
	fmt.Println("MinY", Settings.MinVertical_MM, "MaxY", Settings.MaxVertical_MM)

	// setup 0,0 as the initial location of the plot head
	polarSystem.XOffset = startingLocation.X
	polarSystem.YOffset = startingLocation.Y

	//var interp PositionInterpolater = new(LinearInterpolater)
	var interp PositionInterpolater = new(TrapezoidInterpolater)

	//var previousLeft int = 0
	//var sliceTotal int = 0

	origin := Coordinate{X: 0, Y: 0}
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

		//fmt.Println("Slice", sliceTotal, "------------------------")

		for slice := 1.0; slice <= interp.Slices(); slice++ {

			sliceTarget := interp.Position(slice)
			polarSliceTarget := sliceTarget.ToPolar(polarSystem)

			// calc integer number of steps * 32 that will be made this time slice
			sliceSteps := polarSliceTarget.Minus(previousPolarPos).Scaled(32.0/Settings.StepSize_MM).Ceil().Clamp(127, -127)
			previousPolarPos = previousPolarPos.Add(sliceSteps.Scaled(Settings.StepSize_MM / 32.0))

			//fmt.Println("Slice", slice, "From", previousPolarPos, "to", polarSliceTarget, "steps", sliceSteps)
			//if previousLeft-int(sliceSteps.LeftDist) < -30 {
			//panic("TOO FAR")
			//}
			//previousLeft = int(sliceSteps.LeftDist)

			stepData <- int8(-sliceSteps.LeftDist)
			stepData <- int8(sliceSteps.RightDist)
			//sliceTotal++
		}
		origin = previousPolarPos.ToCoord(polarSystem)
		//fmt.Println("Reprojection error", origin.Minus(target).Len(), "target", target, "actual", origin)
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

	previousSend := time.Now()
	var totalSends int = 0
	var byteData int8 = 0

	// send a -128 to force the arduino to restart and rerequest data
	s.Write([]byte{0x80})

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

		interp := new(TrapezoidInterpolater)
		interp.Setup(Coordinate{}, Coordinate{X: distance, Y: 0}, Coordinate{})
		position := 0.0

		for slice := 1.0; slice <= interp.Slices(); slice++ {

			sliceTarget := interp.Position(slice)

			// calc integer number of steps that will be made this time slice
			sliceSteps := math.Ceil((sliceTarget.X - position) * (32.0 / Settings.StepSize_MM))
			position = position + sliceSteps*(Settings.StepSize_MM/32.0)

			if side == "l" {
				alignStepData <- int8(-sliceSteps)
				alignStepData <- 0
			} else {
				alignStepData <- 0
				alignStepData <- int8(sliceSteps)
			}
		}

		close(alignStepData)
	}
}

// Do mouse tracking, must open up serial port directly in order to send steps in realtime as requested
func PerformMouseTracking() {

	fmt.Println("Opening mouse reader")
	mouse := CreateAndStartMouseReader()
	defer mouse.Close()

	fmt.Println("Opening com port")
	c := &serial.Config{Name: "/dev/ttyAMA0", Baud: 57600}
	s, err := serial.OpenPort(c)
	if err != nil {
		panic(err)
	}
	defer s.Close()

	fmt.Println("Left click to exit, Right click to exit and enter X Y location of pen")

	// buffers to use during serial communication
	writeData := make([]byte, 128)
	readData := make([]byte, 1)

	polarSystem := PolarSystemFromSettings()
	previousPolarPos := PolarCoordinate{LeftDist: Settings.StartingLeftDist_MM, RightDist: Settings.StartingRightDist_MM}
	startingPos := previousPolarPos.ToCoord(polarSystem)
	polarSystem.XOffset = startingPos.X
	polarSystem.YOffset = startingPos.Y

	currentPos := Coordinate{X: 0, Y: 0}

	// max distance that can be travelled in one batch
	maxDistance := 64 * (Settings.MaxSpeed_MM_S * Settings.TimeSlice_US / 1000000.0)

	// send a -128 to force the arduino to restart and rerequest data
	s.Write([]byte{0x80})
	for stepDataOpen := true; stepDataOpen; {
		// wait for next data request
		n, err := s.Read(readData)
		if err != nil {
			panic(err)
		}
		if n != 1 {
			panic(err)
		}

		if mouse.GetLeftButton() {
			updateSettingsPosition(currentPos, polarSystem)
			return
		} else if mouse.GetRightButton() {
			promptForSettingsPosition(polarSystem)
			return
		}

		mouseX, mouseY := mouse.GetPos()
		mousePos := Coordinate{X: float64(mouseX) / 20.0, Y: float64(mouseY) / 20.0}
		direction := mousePos.Minus(currentPos)
		distance := direction.Len()
		if distance == 0.0 {
			direction = Coordinate{X: 1, Y: 0}
		} else {
			direction = direction.Normalized()
		}
		if distance > maxDistance {
			distance = maxDistance
		}
		//fmt.Println("Got mouse pos", mousePos)

		dataToWrite := int(readData[0])
		for i := 0; i < dataToWrite; i += 2 {

			sliceTarget := currentPos.Add(direction.Scaled(float64(i) * distance / 128.0))
			polarSliceTarget := sliceTarget.ToPolar(polarSystem)

			//fmt.Println("i", i, "pos", currentPos, "target", sliceTarget);

			sliceSteps := polarSliceTarget.Minus(previousPolarPos).Scaled(32.0/Settings.StepSize_MM).Ceil().Clamp(127, -127)
			previousPolarPos = previousPolarPos.Add(sliceSteps.Scaled(Settings.StepSize_MM / 32.0))

			writeData[i] = byte(int8(-sliceSteps.LeftDist))
			writeData[i+1] = byte(int8(sliceSteps.RightDist))
		}
		currentPos = previousPolarPos.ToCoord(polarSystem)

		s.Write(writeData)
	}
}

// Update settings with the current position of the pen
func updateSettingsPosition(currentPos Coordinate, polarSystem PolarSystem) {
	finalPolarPos := currentPos.ToPolar(polarSystem)

	fmt.Println("Updating Left from", Settings.StartingLeftDist_MM, "to", finalPolarPos.LeftDist)
	fmt.Println("Updating Right from", Settings.StartingRightDist_MM, "to", finalPolarPos.RightDist)

	Settings.StartingLeftDist_MM = finalPolarPos.LeftDist
	Settings.StartingRightDist_MM = finalPolarPos.RightDist
	Settings.Write()
}

// Ask user for X Y location and then update settings
func promptForSettingsPosition(polarSystem PolarSystem) {
	fmt.Print("Enter X Y location of pen:")
	var finalLocation Coordinate
	if _, err := fmt.Scanln(&finalLocation.X, &finalLocation.Y); err != nil {
		panic(err)
	}
	finalPolarPos := finalLocation.Minus(Coordinate{X: polarSystem.XOffset, Y: polarSystem.YOffset}).ToPolar(polarSystem)

	fmt.Println("Updating Left from", Settings.StartingLeftDist_MM, "to", finalPolarPos.LeftDist)
	fmt.Println("Updating Right from", Settings.StartingRightDist_MM, "to", finalPolarPos.RightDist)

	Settings.StartingLeftDist_MM = finalPolarPos.LeftDist
	Settings.StartingRightDist_MM = finalPolarPos.RightDist
	Settings.Write()
}
