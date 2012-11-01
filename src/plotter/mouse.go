package plotter

// Allows mouse movement to be used to generate a plotCoords path

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

// Event data returned from event file
type InputEvent struct {
	Time  syscall.Timeval // time in seconds since epoch at which event occurred
	Type  uint16          // event type - one of ecodes.EV_*
	Code  uint16          // event code related to the event type
	Value int32           // event value related to the event type
}

// Generates a stream of Coordinates that follow the mouse movement
func GenerateMousePath(eventPath string, plotCoords chan<- Coordinate) {
	defer close(plotCoords)

	mouse, err := os.Open(eventPath)
	if err != nil {
		fmt.Println("Unable to open", eventPath)
		panic(err)
	}
	defer mouse.Close()

	fmt.Println("Left click to exit")
	fmt.Println("Right click to exit and manually enter X,Y position")

	currentPosition := Coordinate{0, 0}
	previousSent := Coordinate{0, 0}
	previousSendTime := time.Now()

	event := InputEvent{}
	buffer := make([]byte, int(unsafe.Sizeof(event)))
	for {
		mouse.Read(buffer)
		b := bytes.NewBuffer(buffer)
		binary.Read(b, binary.LittleEndian, &event)

		//fmt.Println("Read event Type", event.Type, "Code", event.Code, "Value", event.Value)

		switch event.Type {
		case 1: // EV_KEY button press
			switch event.Code {
			case 272: // BTN_LEFT left click
				polarSystem := PolarSystemFromSettings()
				startingPolarPos := PolarCoordinate{LeftDist: Settings.StartingLeftDist_MM, RightDist: Settings.StartingRightDist_MM}
				finalLocation := startingPolarPos.ToCoord(polarSystem).Add(currentPosition)
				finalPolarPos := finalLocation.ToPolar(polarSystem)

				fmt.Println("Updating Left from", Settings.StartingLeftDist_MM, "to", finalPolarPos.LeftDist)
				fmt.Println("Updating Right from", Settings.StartingRightDist_MM, "to", finalPolarPos.RightDist)

				Settings.StartingLeftDist_MM = finalPolarPos.LeftDist
				Settings.StartingRightDist_MM = finalPolarPos.RightDist
				Settings.Write()

				return
			case 273: // BTN_RIGHT right click
				fmt.Print("Enter X,Y location of pen:")
				var finalLocation Coordinate
				if _, err := fmt.Scanln(&finalLocation.X, &finalLocation.Y); err != nil {
					panic(err)
				}
				polarSystem := PolarSystemFromSettings()
				finalPolarPos := finalLocation.ToPolar(polarSystem)

				fmt.Println("Updating Left from", Settings.StartingLeftDist_MM, "to", finalPolarPos.LeftDist)
				fmt.Println("Updating Right from", Settings.StartingRightDist_MM, "to", finalPolarPos.RightDist)

				Settings.StartingLeftDist_MM = finalPolarPos.LeftDist
				Settings.StartingRightDist_MM = finalPolarPos.RightDist
				Settings.Write()

				return
			}

		case 2: // EV_REL movement event
			switch event.Code {
			case 0: // REL_X
				currentPosition.X += float64(event.Value) / 4.0
			case 1: // REL_Y
				currentPosition.Y += float64(event.Value) / 4.0
			}

			if currentPosition.Minus(previousSent).Len() > 10.0 && time.Since(previousSendTime).Seconds() > 1.0 {
				// send twice so that command doesnt get stuck in interpolation pipeline
				plotCoords <- currentPosition
				plotCoords <- currentPosition

				previousSent = currentPosition
				previousSendTime = time.Now()
			}
		}

		if event.Type != 2 {
			continue
		}
	}
}
