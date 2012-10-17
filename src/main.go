package main

import (
	"fmt"
	. "plotter"
)

func main() {

	ReadSettings("/home/pi/GoTest/settings.xml")

	//Settings.TimeSlice_US = 10000
	//Settings.SpoolCircumference_MM = 59
	//Settings.MaxSpeed_MM_S = 300
	//WriteSettings("settings.xml")

	data := GcodeData{
		Lines: []GcodeLine{
			GcodeLine{Command: MOVE, Dest: Coordinate{X: 300, Y: 0}},
			GcodeLine{Command: MOVE, Dest: Coordinate{X: -300, Y: 400}},
			GcodeLine{Command: MOVE, Dest: Coordinate{X: 300, Y: 400}},
			GcodeLine{Command: MOVE, Dest: Coordinate{X: -300, Y: 0}},
			GcodeLine{Command: MOVE, Dest: Coordinate{X: 0, Y: 0}},
		}}

	//	// Generate vertical sinusoidal movement
	//	posGenerator := func(percentage float64) Position {
	//		return Position(math.Cos(percentage * 2. * math.Pi) * AMPLITUDE_MM)
	//	}

	//	posGenerator := func(percentage float64) Position {
	//		return Position(percentage * AMPLITUDE_MM)
	//	}

	//steps := GenStepProfile(7 * time.Second, posGenerator)

	steps := GenStepProfile(data)

	fmt.Println("Generated", len(steps), "steps")
	RunSteps(steps, true)
}
