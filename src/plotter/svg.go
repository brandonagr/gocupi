package plotter

// Reads a text file and generates a program representation of the Gcode

import (
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Initialize regular expressions
func init() {

}

// Used to decode xml data into a readable struct
type Path struct {
	Style string `xml:"style,attr"`
	Data  string `xml:"d,attr"`
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
	case MoveToRelative:
	case LineToRelative:
		return true
	}
	return false
}

// Convert string to command, returns NotAValidCommand if not valid
func ParseCommand(commandString string) PathCommand {

	fmt.Println("Parsing", commandString)

	switch commandString {
	case "M":
		return MoveToAbsolute
	case "m":
		return MoveToRelative
	case "Z":
	case "z":
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
}

// Create new parser
func NewParser(originalPathData string) (parser *PathParser) {

	parser = &PathParser{}

	seperateLetters, _ := regexp.Compile(`([^\s])?([MmZzLlHhVvCcSsQqTtAa])([^\s])?`)
	seperateNumbers, _ := regexp.Compile(`([0-9])([+\-])`)

	pathData := seperateLetters.ReplaceAllString(originalPathData, "$1 $2 $3")
	pathData = seperateNumbers.ReplaceAllString(pathData, "$1 $2")
	pathData = strings.Replace(pathData, ",", " ", -1)
	parser.tokens = strings.Fields(pathData)

	return parser
}

// Parse the data
func (this *PathParser) Parse(data []Coordinate) []Coordinate {

	fmt.Println("Parse!")

	for this.ReadCommand() {

		fmt.Println("currentCommand", this.currentCommand)
		fmt.Println("MoveToAbsolute is ", MoveToAbsolute)

		//switch this.currentCommand {

		//case MoveToAbsolute, MoveToRelative:

		fmt.Println("Handling")

		this.ReadCoords(true)
		data = append(data, this.currentPosition)

		for !this.PeekHasMoreArguments() {
			this.ReadCoords(false)
			data = append(data, this.currentPosition)
		}

		//default:
		//	panic(fmt.Sprint("Unsupported command, saw ", commandString))
		//}
	}

	return data
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

	fmt.Println("tokenIndex", this.tokenIndex, "len", len(this.tokens))

	if this.tokenIndex >= len(this.tokens) {
		return false
	}
	return ParseCommand(this.tokens[this.tokenIndex]) != NotAValidCommand
}

// Read two strings as a pair of doubles
func (this *PathParser) ReadCoords(penUp bool) {

	number := this.tokens[this.tokenIndex]
	x, err := strconv.ParseFloat(number, 64)
	if err != nil {
		panic(fmt.Sprint("Expected a parseable number, but saw", number, "which got parse error", err))
	}

	number = this.tokens[this.tokenIndex+1]
	y, err := strconv.ParseFloat(number, 64)
	if err != nil {
		panic(fmt.Sprint("Expected a parseable number, but saw", number, "which got parse error", err))
	}

	this.tokenIndex += 2

	if this.currentCommand.IsRelative() {
		x += this.currentPosition.X
		y += this.currentPosition.Y
	}

	this.currentPosition = Coordinate{X: x, Y: y, PenUp: penUp}

	fmt.Println("Read currentPosition", this.currentPosition)
}

// read a file and parse its Gcode
func ParseSvgFile(fileName string) (data []Coordinate) {

	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	data = make([]Coordinate, 0)
	decoder := xml.NewDecoder(file)
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:

			fmt.Println("Saw element", se.Name.Local)

			if se.Name.Local == "path" {
				var pathData Path
				decoder.DecodeElement(&pathData, &se)

				parser := NewParser(pathData.Data)

				data = parser.Parse(data)
			}
		}
	}

	if len(data) == 0 {
		panic("SVG contained no Path elements! Only Paths are supported")
	}

	return data
}

// Send svg path points to channel, uses whatever the first Coordinate is as the current location of the head
func GenerateSvgPath(data []Coordinate, size float64, plotCoords chan<- Coordinate) {

	defer close(plotCoords)

	// find top most svg point, so that the path can start there
	topMostPointIndex := 0
	topMostPoint := -100000.0
	for index, point := range data {
		if point.Y > topMostPoint {
			topMostPointIndex = index
			topMostPoint = point.Y
		}
	}

	initialPosition := data[topMostPointIndex]
	minPoint := Coordinate{X: 100000, Y: 100000}
	maxPoint := Coordinate{X: -100000, Y: -10000}

	fmt.Println("Starting location is", initialPosition, "index", topMostPointIndex)

	for _, curTarget := range data {
		point := curTarget.Minus(initialPosition)

		if point.X < minPoint.X {
			minPoint.X = point.X
		} else if point.X > maxPoint.X {
			maxPoint.X = point.X
		}

		if point.Y < minPoint.Y {
			minPoint.Y = point.Y
		} else if point.Y > maxPoint.Y {
			maxPoint.Y = point.Y
		}
	}

	imageSize := maxPoint.Minus(minPoint)
	scale := -size / math.Max(imageSize.X, imageSize.Y)

	fmt.Println("Min", minPoint, "Max", maxPoint, "Scale", scale)

	for index := 0; index < len(data); index++ {
		curTarget := data[(index+topMostPointIndex)%len(data)]
		plotCoords <- curTarget.Minus(initialPosition).Scaled(scale)
	}
}
