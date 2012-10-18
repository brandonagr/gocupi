package plotter

import (
	"fmt"
	serial "github.com/tarm/goserial"
	"time"
)

// Output the coordinates to the screen
func OutputCoords(plotCoords chan Coordinate) {

	for {

		coord, chanOpen := <-plotCoords

		if !chanOpen {
			fmt.Println("Done plotting")
			return
		}

		fmt.Println(coord)
	}
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
