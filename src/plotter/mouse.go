package plotter

// Allows mouse movement to be used to generate a plotCoords path

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"syscall"
	"unsafe"
)

// Interface needed
type MouseReader interface {
	Start(string)
	Close()
	GetPos() Coordinate
	GetLeftButton() bool
	GetRightButton() bool
}

// Data needed by MouseReader
type UnixMouseReader struct {
	eventFile *os.File

	currentPos         Coordinate
	leftButtonPressed  bool
	rightButtonPressed bool

	lock sync.Mutex
}

// Launches a gorountine which updates
func (mouse *UnixMouseReader) Start(eventPath string) {

	var err error
	mouse.eventFile, err = os.Open(eventPath)
	if err != nil {
		fmt.Println("Unable to open", eventPath)
		panic(err)
	}

	go mouse.readDriver()
}

// Event data returned from event file
type inputEvent struct {
	Time  syscall.Timeval // time in seconds since epoch at which event occurred
	Type  uint16          // event type - one of ecodes.EV_*
	Code  uint16          // event code related to the event type
	Value int32           // event value related to the event type
}

// Runs an infinite loop reading from the event file
func (mouse *UnixMouseReader) readDriver() {
	event := inputEvent{}
	buffer := make([]byte, int(unsafe.Sizeof(event)))

	for {
		_, err := mouse.eventFile.Read(buffer)
		if err != nil {
			return
		}

		b := bytes.NewBuffer(buffer)
		binary.Read(b, binary.LittleEndian, &event)

		switch event.Type {
		case 1: // EV_KEY button press

			switch event.Code {
			case 272: // BTN_LEFT left click
				mouse.leftButtonPressed = true

			case 273: // BTN_RIGHT right click
				mouse.rightButtonPressed = true
			}

		case 2: // EV_REL movement event
			switch event.Code {
			case 0: // REL_X
				mouse.movePos(Coordinate{float64(event.Value) / 5.0, 0})
			case 1: // REL_Y
				mouse.movePos(Coordinate{0, float64(event.Value) / 5.0})
			}
		}
	}
}

// Stop the mouse
func (mouse *UnixMouseReader) Close() {

	mouse.eventFile.Close()
}

// Return the current mouse position
func (mouse *UnixMouseReader) GetPos() Coordinate {
	mouse.lock.Lock()
	defer mouse.lock.Unlock()

	return mouse.currentPos
}

// Update the current position
func (mouse *UnixMouseReader) movePos(delta Coordinate) {
	mouse.lock.Lock()
	defer mouse.lock.Unlock()

	mouse.currentPos = mouse.currentPos.Add(delta)
}

// Return true if left button has ever been pressed
func (mouse *UnixMouseReader) GetLeftButton() bool {
	return mouse.leftButtonPressed
}

// Return true if right button has ever been pressed
func (mouse *UnixMouseReader) GetRightButton() bool {
	return mouse.rightButtonPressed
}
