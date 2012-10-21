package main

import (
	. "plotter"
)

func main() {
	ReadSettings("../settings.xml")

	//PerformManualAlignment()
	//return

	plotCoords := make(chan Coordinate, 1024)
	//go GenerateSpiral(Spiral{RadiusBegin: 100, RadiusEnd: 0.1, RadiusDeltaPerRev: 2}, plotCoords)

	go GenerateSlidingCircle(SlidingCircle{Radius: 25, CircleDisplacement: Coordinate{3, 0}, NumbCircles: 70}, plotCoords)

	//go GenerateHilbertCurve(HilbertCurve{Degree: 4, Size: 270.0}, plotCoords)

	//go GenerateParabolic(Parabolic{Height: 270, Lines: 80}, plotCoords)

	//data := ParseGcodeFile("../data/star50.ngc")
	//go GenerateGcodePath(data, plotCoords)

	//DrawToImage("output.png", plotCoords)

	stepData := make(chan byte, 1024)
	go GenerateSteps(plotCoords, stepData)

	//CountSteps(stepData)
	WriteStepsToSerial(stepData)
	//WriteStepsToFile(stepData)
}
