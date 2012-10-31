package plotter

// Allows mouse movement to be used to generate a plotCoords path

import (
	"bytes"
	"encoding/binary"
	"os"
	"syscall"
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
		panic(err)
	}
	defer mouse.Close()

	currentPosition := Coordinate{0, 0}
	event := InputEvent{}
	buffer := make([]byte, int(unsafe.Sizeof(event)))
	for {
		mouse.Read(buffer)
		b := bytes.NewBuffer(buffer)
		binary.Read(b, binary.LittleEndian, &event)

		if event.Type != 2 {
			continue
		}

		switch event.Code {
		case 0:
			currentPosition.X += float64(event.Value)
		case 1:
			currentPosition.Y += float64(event.Value)
		}

		plotCoords <- currentPosition
	}
}
