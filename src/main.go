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

	// channels that will output
	plotCoords := make(chan Coordinate, 1024)

	//go GenerateGcodePath(data, plotCoords)
	go GenerateSpiral(100, 2, 10, plotCoords)

	//OutputCoords(plotCoords)
	RenderCoords(plotCoords, GenerateStepsLinear)
}
