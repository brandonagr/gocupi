package main

import (
	"fmt"
	"syscall"
	"unsafe"
	"encoding/binary"
	"os"
	"bytes"
)

type InputEvent struct {
	Time syscall.Timeval  // time in seconds since epoch at which event occurred
	Type  uint16  // event type - one of ecodes.EV_*
	Code  uint16  // event code related to the event type
	Value int32   // event value related to the event type
}

// Get a useful description for an input event. Example:
//   event at 1347905437.435795, code 01, type 02, val 02
func (ev *InputEvent) String() string {
	return fmt.Sprintf("event at %d.%d, code %02d, type %02d, val %02d",
		   ev.Time.Sec, ev.Time.Usec, ev.Code, ev.Type, ev.Value)
}

var eventsize = int(unsafe.Sizeof(InputEvent{}))

func main() {
	f, err := os.Open("/dev/input/event2")
	if err != nil {
		panic(err)
	}
	defer f.Close()


	var x,y int

	event := InputEvent{}
	buffer := make([]byte, eventsize)
	for {
		f.Read(buffer)
		b := bytes.NewBuffer(buffer)
		binary.Read(b, binary.LittleEndian, &event)

		if event.Type != 2 {
			continue
		}

		switch event.Code {
			case 0:
				x += int(event.Value)
			case 1:
				y += int(event.Value)
		}
		
		fmt.Println(x, y, event)
	}
}
