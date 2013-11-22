package main

import (
	"flag"
	"fmt"
	"github.com/qpliu/qrencode-go/qrencode"
	"math"
	. "plotter"
	"strconv"
	"strings"
)

// set flag usage variable so that entire help will be output
func init() {
	flag.Usage = PrintUsage
}

// main
func main() {
	Settings.Read()

	timelineFlag := flag.Bool("timeline", false, "Generate timeline of step events")
	toImageFlag := flag.Bool("toimage", false, "Output result to an image file instead of to the stepper")
	toFileFlag := flag.Bool("tofile", false, "Output steps to a text file")
	toChartFlag := flag.Bool("tochart", false, "Output a chart of the movement and velocity")
	countFlag := flag.Bool("count", false, "Outputs the time it would take to draw")
	speedSlowFactor := flag.Float64("slowfactor", 1.0, "Divide max speed by this number")
	flag.Parse()

	if *speedSlowFactor < 1.0 {
		panic("slowfactor must be greater than 1")
	}
	// apply slow factor to max speed
	Settings.MaxSpeed_MM_S /= *speedSlowFactor
	Settings.Acceleration_Seconds *= *speedSlowFactor
	Settings.Acceleration_MM_S2 /= *speedSlowFactor
	fmt.Printf("MaxSpeed: %.3f mm/s Accel: %.3f mm/s^2", Settings.MaxSpeed_MM_S, Settings.Acceleration_MM_S2)
	fmt.Println()

	args := flag.Args()
	if len(args) < 1 {
		PrintUsage()
		return
	}

	plotCoords := make(chan Coordinate, 1024)

	switch args[0] {

	case "test":
		plotCoords <- Coordinate{X: 0, Y: 0}
		plotCoords <- Coordinate{X: 10, Y: 0}
		plotCoords <- Coordinate{X: 10.1, Y: 0}
		plotCoords <- Coordinate{X: 10.1, Y: 10}
		plotCoords <- Coordinate{X: 10.1, Y: 10, PenUp: true}
		plotCoords <- Coordinate{X: 20.1, Y: 10, PenUp: true}
		plotCoords <- Coordinate{X: 20.1, Y: 15}
		plotCoords <- Coordinate{X: 20.1, Y: 15, PenUp: true}
		plotCoords <- Coordinate{X: 0, Y: 0, PenUp: true}
		close(plotCoords)

	case "circle":
		params := GetArgsAsFloats(args[1:], 3)
		circleSetup := SlidingCircle{
			Radius:             params[0],
			CircleDisplacement: params[1],
			NumbCircles:        int(params[2]),
		}

		fmt.Println("Generating sliding circle")
		go GenerateSlidingCircle(circleSetup, plotCoords)

	case "gcode":
		scale, _ := strconv.ParseFloat(args[1], 64)
		if scale == 0 {
			scale = 1
		}

		fmt.Println("Generating Gcode path")
		data := ParseGcodeFile(args[2])
		go GenerateGcodePath(data, scale, plotCoords)

	case "grid":
		params := GetArgsAsFloats(args[1:], 2)
		gridSetup := Grid{
			Width: params[0],
			Cells: params[1],
		}

		fmt.Println("Generating grid")
		go GenerateGrid(gridSetup, plotCoords)

	case "hilbert":
		params := GetArgsAsFloats(args[1:], 2)
		hilbertSetup := HilbertCurve{
			Size:   params[0],
			Degree: int(params[1]),
		}

		fmt.Println("Generating hilbert curve")
		go GenerateHilbertCurve(hilbertSetup, plotCoords)

	case "imagearc":
		params := GetArgsAsFloats(args[1:], 2)
		arcSetup := Arc{
			Size:    params[0],
			ArcDist: params[1],
		}

		fmt.Println("Generating image arc path")
		data := LoadImage(args[3])
		data = GaussianImage(data)
		go GenerateArc(arcSetup, data, plotCoords)

	case "imageraster":
		params := GetArgsAsFloats(args[1:], 2)
		rasterSetup := Raster{
			Size:     params[0],
			PenWidth: params[1],
		}

		fmt.Println("Generating image raster path")
		data := LoadImage(args[3])
		go GenerateRaster(rasterSetup, data, plotCoords)

	case "lissa":
		params := GetArgsAsFloats(args[1:], 3)
		posFunc := func(t float64) Coordinate {
			return Coordinate{
				X: params[0] * math.Cos(params[1]*t+math.Pi/2.0),
				Y: params[0] * math.Sin(params[2]*t),
			}
		}

		fmt.Println("Generating Lissajous curve")
		go GenerateParametric(posFunc, plotCoords)

	case "move":
		PerformMouseTracking()
		return

	case "parabolic":
		params := GetArgsAsFloats(args[1:], 3)
		parabolicSetup := Parabolic{
			Radius:           params[0],
			PolygonEdgeCount: params[1],
			Lines:            params[2],
		}

		fmt.Println("Generating parabolic graph")
		go GenerateParabolic(parabolicSetup, plotCoords)

	case "spiral":
		params := GetArgsAsFloats(args[1:], 2)
		spiralSetup := Spiral{
			RadiusBegin:       params[0],
			RadiusEnd:         0.01,
			RadiusDeltaPerRev: params[1],
		}

		fmt.Println("Generating spiral")
		go GenerateSpiral(spiralSetup, plotCoords)

	case "spiro":
		params := GetArgsAsFloats(args[1:], 3)
		bigR := params[0]
		littleR := params[1]
		pen := params[2]

		posFunc := func(t float64) Coordinate {
			return Coordinate{
				X: (bigR-littleR)*math.Cos(t) + pen*math.Cos(((bigR-littleR)/littleR)*t),
				Y: (bigR-littleR)*math.Sin(t) - pen*math.Sin(((bigR-littleR)/littleR)*t),
			}
		}

		fmt.Println("Generating spiro")
		go GenerateParametric(posFunc, plotCoords)

	case "spool":
		PerformManualAlignment()
		return

	case "svg":
		size, _ := strconv.ParseFloat(args[1], 64)
		if size == 0 {
			size = 1
		}

		svgType := "top"
		if len(args) > 3 {
			svgType = strings.ToLower(args[3])
		}

		fmt.Println("Generating svg path")
		data := ParseSvgFile(args[2])
		switch svgType {
		case "top":
			go GenerateSvgTopPath(data, size, plotCoords)

		case "box":
			go GenerateSvgBoxPath(data, size, plotCoords)

		default:
			fmt.Println("Expected top or box as the svg type, and saw", svgType)
			return
		}

	case "text":
		height, _ := strconv.ParseFloat(args[1], 64)
		if height == 0 {
			height = 40
		}

		fmt.Println("Generating text path")
		go GenerateTextPath(args[2], height, plotCoords)

	case "qr":
		params := GetArgsAsFloats(args[1:], 2)
		rasterSetup := Raster{
			Size:     params[0],
			PenWidth: params[1],
		}

		fmt.Println("Generating qr raster path for ", args[3])
		data, err := qrencode.Encode(args[3], qrencode.ECLevelQ)
		if err != nil {
			panic(err)
		}
		imageData := data.ImageWithMargin(1, 0)
		go GenerateRaster(rasterSetup, imageData, plotCoords)

	default:
		PrintUsage()
		return
	}

	if *timelineFlag {
		fmt.Println("Outputting to timeline")
		outputTimeline := make(chan TimelineEvent, 1024)
		go func() {
			for event := range outputTimeline {
				fmt.Println(event)
			}
		}()

		GenerateTimeline(plotCoords, outputTimeline, &Settings)
		return
	}

	if *toImageFlag {
		fmt.Println("Outputting to image")
		DrawToImage("output.png", plotCoords)
		return
	}

	stepData := make(chan int8, 1024)
	go GenerateSteps(plotCoords, stepData)
	switch {
	case *countFlag:
		CountSteps(stepData)
	case *toFileFlag:
		WriteStepsToFile(stepData)
	case *toChartFlag:
		WriteStepsToChart(stepData)
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
-toimage, outputs data to an image of what the render should look like
-tochart, outputs a graph of velocity and position
-tofile, outputs step data to a file
-count, outputs number of steps and render time
-slowfactor=#, slow down rendering by #x, 2x, 4x slower etc

Commands:
circle R d n
	R - radius of circle
	d - displacement per revolution
	n - number of circles

gcode s "path"
	s - scale
	
grid s c
	s - size of square grid
	c - number of cells in grid
 
hilbert s d
	s - size of square
	d - degree of hilbert curve, 2 to 6

imagearc s a "path"
	s - size of long axis
	a - distance between each arc

imageraster s p "path"
	s - size of long axis
	p - pen thickness / distance between rows

lissa s a b
	s - size of drawing
	a - first factor
	b - second factor
 
move

parabolic R c l
	R - radius of shape
	c - count of polygon edges
	l - number of lines per edges
	
spiral R d
	R - initial outter radius
	d - radius delta per revolution

spiro R r p
	R - first circle radius
	r - second circle radius
	p - pen distance

spool

svg s "path" t
	s - size of long axis
	path - path to svg file
	t - type of drawing, either top or box
		top (default) - best for TSP single loop drawings, pen starts on loop at top
		box - pen starts in upper left corner, drawing boundary extents first

text h "string"
	h - letter height
	string - text to print
 
qr s p "string"
	s - size of square
	p - pen thickness, determines how much it fills in solid squares
	string - the text that will be encoded`)
}
