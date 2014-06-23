package polargraph

// Allows current mouse position to be read

import (
	"bytes"
	"encoding/binary"
	"os"
	"sync/atomic"
	"syscall"
	"unsafe"
)

// Interface needed
type MouseReader interface {
	Start(string)
	Close()
	GetPos() (int32, int32)
	GetLeftButton() bool
	GetRightButton() bool
}

// Data needed by MouseReader
type linuxMouseReader struct {
	eventFile *os.File

	currentXPos        int32
	currentYPos        int32
	leftButtonPressed  bool
	rightButtonPressed bool
}

// Return an implementation of MouseReader specific to the current platform
func CreateAndStartMouseReader() MouseReader {
	// todo: implement windows supports?
	var mouse = &linuxMouseReader{}
	mouse.Start(Settings.MousePath)
	return mouse
}

// Launches a gorountine which updates
func (mouse *linuxMouseReader) Start(eventPath string) {
	var err error
	mouse.eventFile, err = os.Open(eventPath)
	if err != nil {
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
func (mouse *linuxMouseReader) readDriver() {
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
				x := atomic.LoadInt32(&mouse.currentXPos)
				x += event.Value
				atomic.StoreInt32(&mouse.currentXPos, x)
			case 1: // REL_Y
				y := atomic.LoadInt32(&mouse.currentYPos)
				y += event.Value
				atomic.StoreInt32(&mouse.currentYPos, y)
			}
		}
	}
}

// Stop the mouse
func (mouse *linuxMouseReader) Close() {
	mouse.eventFile.Close()
}

// Return the current mouse position
func (mouse *linuxMouseReader) GetPos() (x, y int32) {
	return atomic.LoadInt32(&mouse.currentXPos), atomic.LoadInt32(&mouse.currentYPos)
}

// Return true if left button has ever been pressed
func (mouse *linuxMouseReader) GetLeftButton() bool {
	return mouse.leftButtonPressed
}

// Return true if right button has ever been pressed
func (mouse *linuxMouseReader) GetRightButton() bool {
	return mouse.rightButtonPressed
}
