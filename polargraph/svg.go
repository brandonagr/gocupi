package polargraph

// Reads an SVG file with path data and converts that to a series of Coordinates
// PathParser is based on the canvg javascript code from http://code.google.com/p/canvg/

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Used to decode xml data into a readable struct
type Path struct {
	Style string `xml:"style,attr"`
	Data  string `xml:"d,attr"`
}

type Group struct {
	Transform string `xml:"transform,attr"`
	Paths     []Path `xml:"path"`
}

type GroupStipple struct {
	Transform string    `xml:"transform,attr"`
	Stipples  []Stipple `xml:"circle"`
}

type Stipple struct {
	Style string  `xml:"style,attr"`
	DataX float64 `xml:"cx,attr"`
	DataY float64 `xml:"cy,attr"`
	DataR float64 `xml:"r,attr"`
	Id    string  `xml:"id,attr"`
}

// All supported Path Commands
type PathCommand int

const (
	NotAValidCommand PathCommand = iota
	MoveToAbsolute
	MoveToRelative
	ClosePath
	LineToAbsolute
	LineToRelative
)

// PathCommand ToString
func (command PathCommand) String() string {
	switch command {
	case NotAValidCommand:
		return "NotAValidCommand"
	case MoveToAbsolute:
		return "MoveToAbsolute"
	case MoveToRelative:
		return "MoveToRelative"
	case ClosePath:
		return "ClosePath"
	case LineToAbsolute:
		return "LineToAbsolute"
	case LineToRelative:
		return "LineToRelative"
	}
	return "UNKNOWN"
}

// True if the given PathCommand is relative
func (command PathCommand) IsRelative() bool {
	switch command {
	case MoveToRelative, LineToRelative:
		return true
	default:
		return false
	}
	panic("Not reachable")
}

// Convert string to command, returns NotAValidCommand if not valid
func ParseCommand(commandString string) PathCommand {

	switch commandString {
	case "M":
		return MoveToAbsolute
	case "m":
		return MoveToRelative
	case "Z", "z":
		return ClosePath
	case "L":
		return LineToAbsolute
	case "l":
		return LineToRelative
	default:
		return NotAValidCommand
	}
	panic("Not reachable")
}

// Used to parse a path string
type PathParser struct {
	// All of the tokens, strings could be numbers or commands
	tokens []string

	// The token the parser is currently at
	tokenIndex int

	// The last PathCommand that was seen
	currentCommand PathCommand

	// Track current position for relative moves
	currentPosition Coordinate

	// Used to apply a scale factor to all coordinates
	scaleX float64
	scaleY float64

	// The coordinates read for the path
	coordinates []Coordinate
}

// Create new parser
func NewParser(originalPathData string, scaleX, scaleY float64) (parser *PathParser) {

	parser = &PathParser{}

	seperateLetters, _ := regexp.Compile(`([^\s])?([MmZzLlHhVvCcSsQqTtAa])([^\s])?`)
	seperateNumbers, _ := regexp.Compile(`([0-9])([+\-])`)

	pathData := seperateLetters.ReplaceAllString(originalPathData, "$1 $2 $3")
	pathData = seperateNumbers.ReplaceAllString(pathData, "$1 $2")
	pathData = strings.Replace(pathData, ",", " ", -1)
	parser.tokens = strings.Fields(pathData)

	parser.coordinates = make([]Coordinate, 0)
	parser.scaleX = scaleX
	parser.scaleY = scaleY

	return parser
}

// Parse the data
func (this *PathParser) Parse() []Coordinate {

	for this.ReadCommand() {

		switch this.currentCommand {
		case MoveToAbsolute, MoveToRelative:
			this.ReadCoord(true)
			for this.PeekHasMoreArguments() { // can have multiple implicit line coords
				this.ReadCoord(false)
			}

		case LineToAbsolute, LineToRelative:
			for this.PeekHasMoreArguments() {
				this.ReadCoord(false)
			}

		case ClosePath:
			firstPosition := this.coordinates[0]
			this.currentPosition = Coordinate{X: firstPosition.X, Y: firstPosition.Y, PenUp: false}
			this.coordinates = append(this.coordinates, this.currentPosition.ScaledBoth(this.scaleX, this.scaleY))

		default:
			panic(fmt.Sprint("Unsupported command:", this.currentCommand))
		}
	}

	return this.coordinates
}

// Move to next token
func (this *PathParser) ReadCommand() bool {

	if this.tokenIndex >= len(this.tokens) {
		return false
	}

	commandString := this.tokens[this.tokenIndex]
	this.tokenIndex++
	this.currentCommand = ParseCommand(commandString)
	if this.currentCommand == NotAValidCommand {
		panic(fmt.Sprint("Unexpected command, saw ", commandString))
	}

	return true
}

// Return if the next token is a command or not
func (this *PathParser) PeekHasMoreArguments() bool {

	if this.tokenIndex >= len(this.tokens) {
		return false
	}
	return ParseCommand(this.tokens[this.tokenIndex]) == NotAValidCommand
}

// Read two strings as a pair of doubles
func (this *PathParser) ReadCoord(penUp bool) {

	if this.tokenIndex >= len(this.tokens)-1 {
		panic(fmt.Sprint("Not enough tokens to ReadCoord, at ", this.tokenIndex, " of ", len(this.tokens)))
	}

	number := this.tokens[this.tokenIndex]
	this.tokenIndex++
	x, err := strconv.ParseFloat(number, 64)
	if err != nil {
		panic(fmt.Sprint("Expected a parseable number, but saw", number, "which got parse error", err))
	}

	number = this.tokens[this.tokenIndex]
	this.tokenIndex++
	y, err := strconv.ParseFloat(number, 64)
	if err != nil {
		panic(fmt.Sprint("Expected a parseable number, but saw", number, "which got parse error", err))
	}

	if this.currentCommand.IsRelative() {
		x += this.currentPosition.X
		y += this.currentPosition.Y
	}

	this.currentPosition = Coordinate{X: x, Y: y, PenUp: penUp}
	this.coordinates = append(this.coordinates, this.currentPosition.ScaledBoth(this.scaleX, this.scaleY))
}

// read a file
func ParseSvgFile(fileName string) (data []Coordinate) {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	return ParseSvg(file)
}

// read svg xml data
func ParseSvg(svgData io.Reader) (data []Coordinate) {

	data = make([]Coordinate, 0)
	decoder := xml.NewDecoder(svgData)
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:

			if se.Name.Local == "path" {
				var pathData Path
				decoder.DecodeElement(&pathData, &se)
				parser := NewParser(pathData.Data, 1, 1)
				data = append(data, parser.Parse()...)
			} else if se.Name.Local == "g" {
				var groupData Group
				decoder.DecodeElement(&groupData, &se)

				var transformX, transformY, scaleX, scaleY float64
				if groupData.Transform != "" && strings.Contains(groupData.Transform, "scale") {
					if _, err := fmt.Sscanf(groupData.Transform, "translate(%f,%f) scale(%f,%f)", &transformX, &transformY, &scaleX, &scaleY); err != nil {
						fmt.Println("WARNING: Unable to parse svg group transform of ", groupData.Transform)
						scaleX = 1
						scaleY = 1
					}
				} else {
					scaleX = 1
					scaleY = 1
				}

				if groupData.Paths != nil {
					for _, pathElement := range groupData.Paths {
						parser := NewParser(pathElement.Data, scaleX, scaleY)
						data = append(data, parser.Parse()...)
					}
				}
			}
		}
	}

	if len(data) == 0 {
		panic("SVG contained no Path elements! Only Paths are supported")
	}

	return data
}

// read a file
func ParseSvgFileCircle(fileName string) (data []Circle) {
	//fmt.Println("filename",fileName)
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	return ParseSvgCircle(file)
}

// read svg xml data
func ParseSvgCircle(svgData io.Reader) (data []Circle) {

	data = make([]Circle, 0)
	decoder := xml.NewDecoder(svgData)
	for {
		t, _ := decoder.Token()
		//fmt.Println("token: ",t)
		if t == nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			//fmt.Println("xml element name:",se.Name.Local)
			//fmt.Println("se:",se)

			if se.Name.Local == "circle" {
				var stippleData Stipple
				decoder.DecodeElement(&stippleData, &se)

				data = append(data)
			} else if se.Name.Local == "g" {

				var groupData GroupStipple
				decoder.DecodeElement(&groupData, &se)
				//fmt.Println("group data:",groupData)

				var transformX, transformY, scaleX, scaleY float64
				if groupData.Transform != "" && strings.Contains(groupData.Transform, "scale") {
					if _, err := fmt.Sscanf(groupData.Transform, "translate(%f,%f) scale(%f,%f)", &transformX, &transformY, &scaleX, &scaleY); err != nil {
						fmt.Println("WARNING: Unable to parse svg group transform of ", groupData.Transform)
						scaleX = 1
						scaleY = 1
					}
				} else {
					scaleX = 1
					scaleY = 1
				}
				/*
					data = groupData.Stipples
				*/

				if groupData.Stipples != nil {
					for _, stipple := range groupData.Stipples {
						//fmt.Println("id: ",stipple.Id)
						penup := false
						start := false
						if strings.Contains(stipple.Id, "penup") {
							fmt.Println("id: ", stipple.Id)
							penup = true
						}
						if strings.Contains(stipple.Id, "start") {
							fmt.Println("id to start: ", stipple.Id)
							start = true
						}
						circle := Circle{Coordinate{X: stipple.DataX, Y: stipple.DataY, PenUp: penup}, stipple.DataR, start}
						data = append(data, circle)
					}
				}

			}
		} //end switch
	}

	//search id with "start" tag in it
	initialPositionIndex := 0
	for index, point := range data {
		if point.Start == true {
			initialPositionIndex = index
		}
	}

	dataSorted := make([]Circle, 0)
	for index := 0; index < len(data); index++ {
		curTarget := data[(index+initialPositionIndex)%len(data)]
		dataSorted = append(dataSorted, curTarget)
	}
	fmt.Println("initialPositionIndex: ", initialPositionIndex, "len(data): ", len(data), "len(dataSorted): ", len(dataSorted))
	fmt.Println("initialPosition: ", dataSorted[initialPositionIndex], "nextPosition: ", dataSorted[initialPositionIndex+1])

	if len(dataSorted) == 0 {
		panic("SVG contained no Circle elements! Only Circle are supported")
	}
	//fmt.Println("[]Circles",data)
	return dataSorted
}

// Send svg path points to channel
func GenerateSvgCenterPath(data Coordinates, size float64, plotCoords chan<- Coordinate) {

	defer close(plotCoords)

	minPoint, maxPoint := data.Extents()

	imageSize := maxPoint.Minus(minPoint)
	scale := size / imageSize.X

	fmt.Println("SVG Min:", minPoint, "Max:", maxPoint, "Scale:", scale)

	if imageSize.X*scale > (Settings.DrawingSurfaceMaxX_MM-Settings.DrawingSurfaceMinX_MM) || imageSize.Y*scale > (Settings.DrawingSurfaceMaxY_MM-Settings.DrawingSurfaceMinY_MM) {
		panic(fmt.Sprint(
			"SVG coordinates extend past drawable surface, as defined in setup. Scaled svg size was: ",
			imageSize,
			" And settings bounds are, X: ", Settings.DrawingSurfaceMaxX_MM, " - ", Settings.DrawingSurfaceMinX_MM,
			" Y: ", Settings.DrawingSurfaceMaxY_MM, " - ", Settings.DrawingSurfaceMinY_MM))
	}

	// want to center the image horizontally, so need actual world space location of gondola at start
	polarSystem := PolarSystemFromSettings()
	previousPolarPos := PolarCoordinate{LeftDist: Settings.StartingLeftDist_MM, RightDist: Settings.StartingRightDist_MM}
	startingLocation := previousPolarPos.ToCoord(polarSystem)
	surfaceWidth := Settings.DrawingSurfaceMaxX_MM - Settings.DrawingSurfaceMinX_MM

	// actual starting location - desired = offset
	imageDistanceFromLeftMargin := (surfaceWidth - (imageSize.X * scale)) / 2
	centeringOffset := Coordinate{X: startingLocation.X - imageDistanceFromLeftMargin}
	fmt.Println("Pen starting position:", startingLocation, "Drawing surface width:", surfaceWidth, "Centering Offset:", centeringOffset)

	plotCoords <- Coordinate{X: 0, Y: 0, PenUp: true}

	for index := 0; index < len(data); index++ {
		curTarget := data[index]
		plotCoords <- curTarget.Minus(minPoint).Scaled(scale).Minus(centeringOffset)
	}

	plotCoords <- Coordinate{X: 0, Y: 0, PenUp: true}
}

// Send svg path points to channel
func GenerateSvgBoxPath(data Coordinates, size float64, plotCoords chan<- Coordinate) {

	defer close(plotCoords)

	minPoint, maxPoint := data.Extents()

	imageSize := maxPoint.Minus(minPoint)
	scale := size / math.Max(imageSize.X, imageSize.Y)

	fmt.Println("SVG Min:", minPoint, "Max:", maxPoint, "Scale:", scale)

	if imageSize.X*scale > (Settings.DrawingSurfaceMaxX_MM-Settings.DrawingSurfaceMinX_MM) || imageSize.Y*scale > (Settings.DrawingSurfaceMaxY_MM-Settings.DrawingSurfaceMinY_MM) {
		panic(fmt.Sprint(
			"SVG coordinates extend past drawable surface, as defined in setup. Scaled svg size was: ",
			imageSize,
			" And settings bounds are, X: ", Settings.DrawingSurfaceMaxX_MM, " - ", Settings.DrawingSurfaceMinX_MM,
			" Y: ", Settings.DrawingSurfaceMaxY_MM, " - ", Settings.DrawingSurfaceMinY_MM))
	}

	plotCoords <- Coordinate{X: 0, Y: 0, PenUp: true}
	plotCoords <- Coordinate{X: 0, Y: 10, PenUp: true}
	plotCoords <- Coordinate{X: 0, Y: maxPoint.Y - minPoint.Y, PenUp: true}.Scaled(scale)
	plotCoords <- Coordinate{X: maxPoint.X - minPoint.X, Y: maxPoint.Y - minPoint.Y, PenUp: true}.Scaled(scale)
	plotCoords <- Coordinate{X: maxPoint.X - minPoint.X, Y: 0, PenUp: true}.Scaled(scale)
	plotCoords <- Coordinate{X: 0, Y: 0, PenUp: true}.Scaled(scale)

	for index := 0; index < len(data); index++ {
		curTarget := data[index]
		plotCoords <- curTarget.Minus(minPoint).Scaled(scale)
	}

	plotCoords <- Coordinate{X: 0, Y: 0, PenUp: true}
}

// Send svg path points to channel
func GenerateSvgTopPath(data Coordinates, size float64, plotCoords chan<- Coordinate) {

	defer close(plotCoords)

	minPoint, maxPoint := data.Extents()

	imageSize := maxPoint.Minus(minPoint)
	scale := size / math.Max(imageSize.X, imageSize.Y)

	fmt.Println("SVG Min:", minPoint, "Max:", maxPoint, "Scale:", scale)

	if imageSize.X*scale > (Settings.DrawingSurfaceMaxX_MM-Settings.DrawingSurfaceMinX_MM) || imageSize.Y*scale > (Settings.DrawingSurfaceMaxY_MM-Settings.DrawingSurfaceMinY_MM) {
		panic(fmt.Sprint(
			"SVG coordinates extend past drawable surface, as defined in setup. Scaled svg size was: ",
			imageSize,
			" And settings bounds are, X: ", Settings.DrawingSurfaceMaxX_MM, " - ", Settings.DrawingSurfaceMinX_MM,
			" Y: ", Settings.DrawingSurfaceMaxY_MM, " - ", Settings.DrawingSurfaceMinY_MM))
	}

	// find top most svg point, so that the path can start there	244		// find minPoint of coordinates, which will be upper left, where the pen will start
	initialPositionIndex := 0
	initialPosition := Coordinate{X: 100000.0, Y: 100000.0}
	for index, point := range data {
		if point.Y < initialPosition.Y || (point.Y == initialPosition.Y && point.X < initialPosition.X) {
			initialPositionIndex = index
			initialPosition = point
		}
	}
	initialPosition.PenUp = false

	fmt.Println("SVG initial top position:", initialPosition)

	for index := 0; index < len(data); index++ {
		curTarget := data[(index+initialPositionIndex)%len(data)]
		plotCoords <- curTarget.Minus(initialPosition).Scaled(scale)
	}

	plotCoords <- Coordinate{X: 0, Y: 0, PenUp: true}
}
