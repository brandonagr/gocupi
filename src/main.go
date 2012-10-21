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

	//go GenerateSlidingCircle(SlidingCircle{Radius: 100, CircleDisplacement: Coordinate{2, 0}, NumbCircles: 50}, plotCoords)

	go GenerateHilbertCurve(HilbertCurve{Degree: 6, Size: 350.0}, plotCoords)

	//data := ParseGcodeFile("../data/allegro lines hires.ngc")
	//go GenerateGcodePath(data, plotCoords)

	//DrawToImage("output.png", plotCoords)

	stepData := make(chan byte, 1024)
	go GenerateSteps(plotCoords, stepData)

	//CountSteps(stepData)
	WriteStepsToSerial(stepData)
	//WriteStepsToFile(stepData)
}
