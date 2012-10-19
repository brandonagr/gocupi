package main

import (
	. "plotter"
)

func main() {
	ReadSettings("../settings.xml")

	//PerformManualAlignment()

	plotCoords := make(chan Coordinate, 1024)
	//go GenerateSpiral(Spiral{RadiusBegin: 100, RadiusEnd: 2, RadiusDeltaPerRev: 5}, plotCoords)

	//go GenerateSlidingCircle(SlidingCircle{Radius: 50, CircleDisplacement: Coordinate{5, 0}, NumbCircles: 50}, plotCoords)

	//go GenerateHilbertCurve(HilbertCurve{Degree: 3, Size: 100.0}, plotCoords)

	data := ParseGcodeFile("../data/allegro lines hires.ngc")
	go GenerateGcodePath(data, plotCoords)

	DrawToImage("output.png", plotCoords)

	//stepData := make(chan byte, 1024)
	//go GenerateStepsLinear(plotCoords, stepData)

	//CountSteps(stepData)
	//WriteStepsToSerial(stepData)
}
