package main

import (
	"flag"
	"fmt"
	. "plotter"
	"strconv"
)

func main() {
	ReadSettings("../settings.xml")

	alignFlag := flag.Bool("align", false, "Run interactive spool alignment mode")
	toImageFlag := flag.Bool("toimage", false, "Output result to an image file instead of to the stepper")
	toFileFlag := flag.Bool("tofile", false, "Output steps to a text file")
	countFlag := flag.Bool("count", false, "Outputs the time it would take to draw")
	flag.Parse()

	if *alignFlag {
		PerformManualAlignment()
		return
	}

	args := flag.Args()
	if len(args) < 2 {
		PrintUsage()
		return
	}

	plotCoords := make(chan Coordinate, 1024)

	switch args[0] {
	case "gcode":
		scale, _ := strconv.ParseFloat(args[1], 64)
		if scale == 0 {
			scale = 1
		}

		fmt.Println("Generating Gcode path")
		data := ParseGcodeFile(args[2])
		go GenerateGcodePath(data, scale, plotCoords)

	case "spiro":
		spiroSetup := Spiro{}
		spiroSetup.BigR, _ = strconv.ParseFloat(args[1], 64)
		spiroSetup.LittleR, _ = strconv.ParseFloat(args[2], 64)
		spiroSetup.Pen, _ = strconv.ParseFloat(args[3], 64)

		if spiroSetup.BigR == 0 || spiroSetup.LittleR == 0 || spiroSetup.Pen == 0 {
			panic("Missing parameters")
		}
		fmt.Println("Generating spiro")
		go GenerateSpiro(spiroSetup, plotCoords)

	case "lissa":
		lissajous := Lissajous{}
		lissajous.Scale, _ = strconv.ParseFloat(args[1], 64)
		lissajous.A, _ = strconv.ParseFloat(args[2], 64)
		lissajous.B, _ = strconv.ParseFloat(args[3], 64)

		if lissajous.Scale == 0 || lissajous.A == 0 || lissajous.B == 0 {
			panic("Missing parameters")
		}
		fmt.Println("Generating Lissajous curve")
		go GenerateLissajous(lissajous, plotCoords)

	case "spiral":
		spiralSetup := Spiral{}
		spiralSetup.RadiusBegin, _ = strconv.ParseFloat(args[1], 64)
		spiralSetup.RadiusEnd, _ = strconv.ParseFloat(args[2], 64)
		spiralSetup.RadiusDeltaPerRev, _ = strconv.ParseFloat(args[3], 64)

		if spiralSetup.RadiusBegin == 0 || spiralSetup.RadiusEnd == 0 || spiralSetup.RadiusDeltaPerRev == 0 {
			panic("Missing parameters")
		}
		fmt.Println("Generating spiral")
		go GenerateSpiral(spiralSetup, plotCoords)

	case "circle":
		circleSetup := SlidingCircle{}
		circleSetup.Radius, _ = strconv.ParseFloat(args[1], 64)
		circleSetup.CircleDisplacement, _ = strconv.ParseFloat(args[2], 64)
		n, _ := strconv.ParseInt(args[3], 10, 32)
		circleSetup.NumbCircles = int(n)

		if circleSetup.Radius == 0 || circleSetup.CircleDisplacement == 0 || circleSetup.NumbCircles == 0 {
			panic("Missing parameters")
		}
		fmt.Println("Generating sliding circle")
		go GenerateSlidingCircle(circleSetup, plotCoords)

	case "hilbert":
		hilbertSetup := HilbertCurve{}
		d, _ := strconv.ParseInt(args[1], 10, 32)
		hilbertSetup.Degree = int(d)
		hilbertSetup.Size, _ = strconv.ParseFloat(args[2], 64)

		if hilbertSetup.Degree == 0 || hilbertSetup.Size == 0 {
			panic("Missing parameters")
		}

		fmt.Println("Generating hilbert curve")
		go GenerateHilbertCurve(hilbertSetup, plotCoords)

		// if there are multiple segments making up a single straight line, combine into just one line
		combineStraightCoords := make(chan Coordinate, 1024)
		go SmoothStraightCoords(plotCoords, combineStraightCoords)
		plotCoords = combineStraightCoords

	default:
		PrintUsage()
		return
	}

	if *toImageFlag {
		fmt.Println("Outputting to image")
		DrawToImage("output.png", plotCoords)
		return
	}

	stepData := make(chan byte, 1024)
	go GenerateSteps(plotCoords, stepData)
	switch {
	case *countFlag:
		CountSteps(stepData)
	case *toFileFlag:
		WriteStepsToFile(stepData)
	default:
		WriteStepsToSerial(stepData)
	}
}

// output valid command line arguments
func PrintUsage() {
	fmt.Println(`Usage:
-align
-toimage
-tofile
-count
gcode s "path" (s scale)
spiro R r p (R first circle radius) (r second circle radius) (p pen distance)
lissa s a b (s scale of drawing) (a factor) (b factor)
spiral R r d (R begin radius) (r end radius) (d radius delta per revolution)
circle R d n (R radius) (d displacement per revolution) (n number of circles)
hilbert d s (d degree(ie 1 to 6)) (s size)
parabolic h n (h height of square) (n number of lines)`)
}
