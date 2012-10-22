package main

import (
	. "plotter"
)

func main() {
	ReadSettings("../settings.xml")

	//PerformManualAlignment()
	//return

	plotCoords := make(chan Coordinate, 1024)
	//go GenerateSpiral(Spiral{RadiusBegin: 100, RadiusEnd: 0.1, RadiusDeltaPerRev: 10}, plotCoords)

	//go GenerateSlidingCircle(SlidingCircle{Radius: 25, CircleDisplacement: Coordinate{3, 0}, NumbCircles: 70}, plotCoords)

	go GenerateHilbertCurve(HilbertCurve{Degree: 3, Size: 270.0}, plotCoords)

	//go GenerateParabolic(Parabolic{Height: 270, Lines: 80}, plotCoords)

	//data := ParseGcodeFile("../data/allegro lines.ngc")
	//go GenerateGcodePath(data, plotCoords)

	// if there are multiple segments making up a single straight line, combine into just one line
	combineStraightCoords := make(chan Coordinate, 1024)
	go SmoothStraightCoords(plotCoords, combineStraightCoords)

	DrawToImage("output.png", combineStraightCoords)

	//stepData := make(chan byte, 1024)
	//go GenerateSteps(combineStraightCoords, stepData)

	//CountSteps(stepData)
	//WriteStepsToSerial(stepData)
	//WriteStepsToFile(stepData)
}
