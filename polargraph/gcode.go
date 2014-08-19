package polargraph

// Reads a text file and generates a program representation of the Gcode

import (
	"bufio"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

// Any of the possible command types that are supported
type GcodeCommand int32
var penupcode float64

const (
	MOVE_RAPID GcodeCommand = 0
	MOVE                    = 1

	SET_UNITS_INCHES = 20
	SET_UNITS_MM     = 21
)

// Data on a single line, the command followed by any values on the line
type GcodeLine struct {
	Command GcodeCommand
	//data map[string]float64
	Dest Coordinate
}

// All of the data from a file
type GcodeData struct {
	Lines []GcodeLine
}

// read a file and parse its Gcode
func ParseGcodeFile(fileName string) GcodeData {

	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(file)

	lines := make([]string, 0)
	for {
		l, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		lines = append(lines, strings.TrimSpace(l))
	}

	return ParseGcode(lines)
}

// read all of the fileData lines, generating a GcodeData object
func ParseGcode(fileData []string) (data GcodeData) {

	data = GcodeData{make([]GcodeLine, 0)}
	

	for _, fileLine := range fileData {

		if strings.HasPrefix(fileLine, "G00") || strings.HasPrefix(fileLine, "G01") {

			coord := Coordinate{X: math.MaxFloat64, Y: math.MaxFloat64, PenUp: false}
			
			for _, part := range strings.Split(fileLine, " ") {
				
				var err interface{}
				if strings.HasPrefix(part, "X") {
					coord.X, err = strconv.ParseFloat(part[1:], 64)
					if err != nil {
						panic(err)
					}
				} else if strings.HasPrefix(part, "Y") {
					if strings.HasSuffix(part, ";") {
				        part = part[:len(part)-len(";")]
					}
					coord.Y, err = strconv.ParseFloat(part[1:], 64)
					if err != nil {
						panic(err)
					}
					coord.Y = -coord.Y
				} else if strings.HasPrefix(part, "Z") {
				    if strings.HasSuffix(part, ";") {
				        part = part[:len(part)-len(";")]
				    }
				    penupcode, err = strconv.ParseFloat(part[1:],64)
					if err != nil {
						panic(err)
					}
					
					
				}
				
			}
			if penupcode == 50 {
				coord.PenUp = true
				//fmt.Println("PENUP")
			} else {
				coord.PenUp = false
				//fmt.Println("PENDOWN")
			}
			if coord.X != math.MaxFloat64 && coord.Y != math.MaxFloat64 {
				data.Lines = append(data.Lines, GcodeLine{Command: MOVE, Dest: coord})
			}
		}
	}

	return
}

// Given GCodeData, returns all of the
func GenerateGcodePath(data GcodeData, scale float64, plotCoords chan<- Coordinate) {

	defer close(plotCoords)

	for _, curTarget := range data.Lines {
	   plotCoords <- curTarget.Dest.Scaled(scale)
	}
}
