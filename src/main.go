package main

import (
	"flag"
	"fmt"
	"math"
	. "plotter"
	"strconv"
)

func main() {
	ReadSettings("../settings.xml")

	alignFlag := flag.Bool("align", false, "Run interactive spool alignment mode")
	toImageFlag := flag.Bool("toimage", false, "Output result to an image file instead of to the stepper")
	toFileFlag := flag.Bool("tofile", false, "Output steps to a text file")
	countFlag := flag.Bool("count", false, "Outputs the time it would take to draw")
	cubicSmoothFlag := flag.Bool("cubicsmooth", false, "Uses cubic spline for straight lines to speed up / slow down")
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

	case "svg":
		size, _ := strconv.ParseFloat(args[1], 64)
		if size == 0 {
			size = 1
		}

		fmt.Println("Generating svg path")
		data := ParseSvgFile(args[2])
		go GenerateSvgPath(data, size, plotCoords)

	case "spiro":
		bigR, _ := strconv.ParseFloat(args[1], 64)
		littleR, _ := strconv.ParseFloat(args[2], 64)
		pen, _ := strconv.ParseFloat(args[3], 64)

		if bigR == 0 || littleR == 0 || pen == 0 {
			panic("Missing parameters")
		}
		posFunc := func(t float64) Coordinate {
			return Coordinate{
				(bigR-littleR)*math.Cos(t) + pen*math.Cos(((bigR-littleR)/littleR)*t),
				(bigR-littleR)*math.Sin(t) - pen*math.Sin(((bigR-littleR)/littleR)*t),
			}
		}

		fmt.Println("Generating spiro")
		go GenerateParametric(posFunc, plotCoords)

	case "lissa":
		scale, _ := strconv.ParseFloat(args[1], 64)
		factorA, _ := strconv.ParseFloat(args[2], 64)
		factorB, _ := strconv.ParseFloat(args[3], 64)

		if scale == 0 || factorA == 0 || factorB == 0 {
			panic("Missing parameters")
		}
		posFunc := func(t float64) Coordinate {
			return Coordinate{
				scale * math.Cos(factorA*t+math.Pi/2.0),
				scale * math.Sin(factorB*t),
			}
		}

		fmt.Println("Generating Lissajous curve")
		go GenerateParametric(posFunc, plotCoords)

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
		hilbertSetup.Size, _ = strconv.ParseFloat(args[1], 64)
		d, _ := strconv.ParseInt(args[2], 10, 32)
		hilbertSetup.Degree = int(d)

		if hilbertSetup.Degree == 0 || hilbertSetup.Size == 0 {
			panic("Missing parameters")
		}

		fmt.Println("Generating hilbert curve")
		go GenerateHilbertCurve(hilbertSetup, plotCoords)

		// if there are multiple segments making up a single straight line, combine into just one line
		combineStraightCoords := make(chan Coordinate, 1024)
		go SmoothStraightCoords(plotCoords, combineStraightCoords)
		plotCoords = combineStraightCoords

	case "parabolic":
		parabolicSetup := Parabolic{}
		parabolicSetup.Radius, _ = strconv.ParseFloat(args[1], 64)
		parabolicSetup.PolygonEdgeCount, _ = strconv.ParseFloat(args[2], 64)
		parabolicSetup.Lines, _ = strconv.ParseFloat(args[3], 64)

		if parabolicSetup.Radius == 0 || parabolicSetup.PolygonEdgeCount == 0 || parabolicSetup.Lines == 0 {
			panic("Missing parameters")
		}
		fmt.Println("Generating parabolic graph")
		go GenerateParabolic(parabolicSetup, plotCoords)

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
	go GenerateSteps(plotCoords, stepData, *cubicSmoothFlag)
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
	fmt.Println(`Usage: (flags) COMMAND PARAMS...
Flags:
-align, runs alignment mode where you can wind the spools a certain distance
-toimage, outputs data to an image of what the render should look like
-tofile, outputs step data to a file
-count, outputs number of steps and render time
-cubicsmooth, uses cubic spline smoothing when generating steps

Commands:
gcode s "path" (s scale)
svg s "path" (s size)
spiro R r p (R first circle radius) (r second circle radius) (p pen distance)
lissa s a b (s scale of drawing) (a factor) (b factor)
spiral R r d (R begin radius) (r end radius) (d radius delta per revolution)
circle R d n (R radius) (d displacement per revolution) (n number of circles)
hilbert s d (s size) (d degree(ie 1 to 6))
parabolic R c l (R radius) (c count of polygon edges) (l number of lines)`)
}
