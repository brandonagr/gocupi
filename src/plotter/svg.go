package plotter

// Reads a text file and generates a program representation of the Gcode

import (
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type G struct {
	Transform string `xml:"transform,attr"`
	PathData  []Path `xml:"path"`
}
type Path struct {
	Style string `xml:"style,attr"`
	Data  string `xml:"d,attr"`
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

		//fmt.Println("Got token", t)

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "g" {
				var g G
				decoder.DecodeElement(&g, &se)

				data = g.ParsePaths(data)
			}
		}

	}

	return data
}

// Parse string paths into list of coordinates
func (g G) ParsePaths(data []Coordinate) []Coordinate {
	for _, path := range g.PathData {

		if path.Data[0] != 'm' {
			panic(fmt.Sprint("Unexpected first character in path data, expected m and got", path.Data[0]))
		}

		//fmt.Println("Parsing", path.Data)

		currentCoord := Coordinate{0, 0}
		for index, parts := range strings.Split(path.Data[2:], " ") {

			coordParts := strings.Split(parts, ",")
			if len(coordParts) != 2 {
				panic(fmt.Sprint("Expected comma seperated pair of coords and saw", parts))
			}

			var coord Coordinate
			var err interface{}
			coord.X, err = strconv.ParseFloat(coordParts[0], 64)
			if err != nil {
				panic(err)
			}
			coord.X = -coord.X
			coord.Y, err = strconv.ParseFloat(coordParts[1], 64)
			if err != nil {
				panic(err)
			}

			if index == 0 {
				currentCoord = coord
			} else {
				currentCoord = currentCoord.Add(coord)
			}

			data = append(data, currentCoord)
		}
	}

	return data
}

// Send svg path points to channel, uses whatever the first Coordinate is as the current location of the head
func GenerateSvgPath(data []Coordinate, size float64, plotCoords chan<- Coordinate) {

	defer close(plotCoords)

	initialPosition := data[0]
	minPoint := Coordinate{100000, 100000}
	maxPoint := Coordinate{-100000, -10000}

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

	imageSize := minPoint.Minus(maxPoint)
	var scale float64
	if imageSize.X > imageSize.Y {
		scale = size / imageSize.X
	} else {
		scale = size / imageSize.Y
	}

	for _, curTarget := range data {
		plotCoords <- curTarget.Minus(initialPosition).Scaled(scale)
	}
}
