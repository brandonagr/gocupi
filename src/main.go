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
	speedSlowFactor := flag.Float64("slowfactor", 1.0, "Divide max speed by this number")
	flag.Parse()

	if *speedSlowFactor < 1.0 {
		panic("slowfactor must be greater than 1")
	}
	// apply slow factor to max speed
	Settings.MaxSpeed_MM_S /= *speedSlowFactor
	Settings.Acceleration_Seconds *= *speedSlowFactor
	Settings.Acceleration_MM_S2 /= *speedSlowFactor
	fmt.Println("MaxSpeed:", Settings.MaxSpeed_MM_S)

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
		params := GetArgsAsFloats(args[1:], 3)
		bigR := params[0]
		littleR := params[1]
		pen := params[2]

		posFunc := func(t float64) Coordinate {
			return Coordinate{
				(bigR-littleR)*math.Cos(t) + pen*math.Cos(((bigR-littleR)/littleR)*t),
				(bigR-littleR)*math.Sin(t) - pen*math.Sin(((bigR-littleR)/littleR)*t),
			}
		}

		fmt.Println("Generating spiro")
		go GenerateParametric(posFunc, plotCoords)

	case "lissa":
		params := GetArgsAsFloats(args[1:], 3)
		posFunc := func(t float64) Coordinate {
			return Coordinate{
				params[0] * math.Cos(params[1]*t+math.Pi/2.0),
				params[0] * math.Sin(params[2]*t),
			}
		}

		fmt.Println("Generating Lissajous curve")
		go GenerateParametric(posFunc, plotCoords)

	case "spiral":
		params := GetArgsAsFloats(args[1:], 3)
		spiralSetup := Spiral{
			RadiusBegin:       params[0],
			RadiusEnd:         params[1],
			RadiusDeltaPerRev: params[2],
		}

		fmt.Println("Generating spiral")
		go GenerateSpiral(spiralSetup, plotCoords)

	case "circle":
		params := GetArgsAsFloats(args[1:], 3)
		circleSetup := SlidingCircle{
			Radius:             params[0],
			CircleDisplacement: params[1],
			NumbCircles:        int(params[2]),
		}

		fmt.Println("Generating sliding circle")
		go GenerateSlidingCircle(circleSetup, plotCoords)

	case "hilbert":
		params := GetArgsAsFloats(args[1:], 2)
		hilbertSetup := HilbertCurve{
			Size:   params[0],
			Degree: int(params[1]),
		}

		fmt.Println("Generating hilbert curve")
		go GenerateHilbertCurve(hilbertSetup, plotCoords)

	case "parabolic":
		params := GetArgsAsFloats(args[1:], 3)
		parabolicSetup := Parabolic{
			Radius:           params[0],
			PolygonEdgeCount: params[1],
			Lines:            params[2],
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

	stepData := make(chan int8, 1024)
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

// Parse a series of numbers as floats
func GetArgsAsFloats(args []string, expectedCount int) []float64 {

	if len(args) < expectedCount {
		PrintUsage()
		panic(fmt.Sprint("Expected at least", expectedCount, "numeric parameters and only saw", len(args)))
	}

	numbers := make([]float64, expectedCount)

	var err error
	for argIndex := 0; argIndex < expectedCount; argIndex++ {
		if numbers[argIndex], err = strconv.ParseFloat(args[argIndex], 64); err != nil {
			panic(fmt.Sprint("Unable to parse", args[argIndex], "as a float: ", err))
		}

		if numbers[argIndex] == 0 {
			panic(fmt.Sprint("0 is not a valid value for parameter", argIndex))
		}
	}

	return numbers
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
-slowfactor=#, slow down rendering by #x, 2x, 4x slower etc

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
