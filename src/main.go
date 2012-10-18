package main

import (
	. "plotter"
)

func main() {

	ReadSettings("../settings.xml")

	// data := GcodeData{
	// 	Lines: []GcodeLine{
	// 		GcodeLine{Command: MOVE, Dest: Coordinate{X: 300, Y: 0}},
	// 		GcodeLine{Command: MOVE, Dest: Coordinate{X: -300, Y: 400}},
	// 		GcodeLine{Command: MOVE, Dest: Coordinate{X: 300, Y: 400}},
	// 		GcodeLine{Command: MOVE, Dest: Coordinate{X: -300, Y: 0}},
	// 		GcodeLine{Command: MOVE, Dest: Coordinate{X: 0, Y: 0}},
	// 	},
	// }
	//go GenerateGcodePath(data, plotCoords)

	plotCoords := make(chan Coordinate, 1024)
	go GenerateSpiral(Spiral{RadiusBegin: 100, RadiusEnd: 2, RadiusDeltaPerRev: 10}, plotCoords)

	stepData := make(chan byte, 1024)
	go GenerateStepsLinear(plotCoords, stepData)

	CountSteps(stepData)
	//WriteStepsToSerial(stepData)
}
