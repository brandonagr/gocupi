package polargraph

// Tests for svg loading

import (
	"strings"
	"testing"
)

// Should read and apply any scale transforms
func TestSVGScale(t *testing.T) {

	svgText := `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<!-- Created with the Eggbot TSP art toolkit (http://egg-bot.com) -->

<svg xmlns="http://www.w3.org/2000/svg"
     xmlns:inkscape="http://www.inkscape.org/namespaces/inkscape"
     xmlns:sodipodi="http://sodipodi.sourceforge.net/DTD/sodipodi-0.dtd"
     xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
     xmlns:dc="http://purl.org/dc/elements/1.1/"
     xmlns:cc="http://creativecommons.org/ns#"
     height="800"
     width="3200">
  <g      transform="translate(1339, 506) scale(1, -1)">
    <path style="fill:none;stroke:#000000;stroke-width:1"
          d="m 213,152 3,0 2,0 3,0 2,2"/>
  </g>
</svg>`

	var expectedResult, result []Coordinate
	result = ParseSvg(strings.NewReader(svgText))

	expectedResult = []Coordinate{
		Coordinate{X: 213, Y: -152, PenUp: true},
		Coordinate{X: 216, Y: -152, PenUp: false},
		Coordinate{X: 218, Y: -152, PenUp: false},
		Coordinate{X: 221, Y: -152, PenUp: false},
		Coordinate{X: 223, Y: -154, PenUp: false},
	}
	assertAreEqual(expectedResult, result, t)
}

// Sample should correctly average values together
func TestSVGPath(t *testing.T) {

	var p *PathParser
	var expectedResult, result []Coordinate

	p = NewParser("M364.51173 507.85272L364.283279688 508.984281094L363.660275 509.90832375", 1, 1)
	expectedResult = []Coordinate{
		Coordinate{X: 364.51173, Y: 507.85272, PenUp: true},
		Coordinate{X: 364.283279688, Y: 508.984281094, PenUp: false},
		Coordinate{X: 363.660275, Y: 509.90832375, PenUp: false},
	}
	result = p.Parse()
	assertAreEqual(expectedResult, result, t)

	p = NewParser("m 213,152 3,0 2,0 3,0 2,2", 1, 1)
	expectedResult = []Coordinate{
		Coordinate{X: 213, Y: 152, PenUp: true},
		Coordinate{X: 216, Y: 152, PenUp: false},
		Coordinate{X: 218, Y: 152, PenUp: false},
		Coordinate{X: 221, Y: 152, PenUp: false},
		Coordinate{X: 223, Y: 154, PenUp: false},
	}
	result = p.Parse()
	assertAreEqual(expectedResult, result, t)

	p = NewParser("m 213,152 3,0 2,0 3,0 2,2     z", 1, 1)
	expectedResult = []Coordinate{
		Coordinate{X: 213, Y: 152, PenUp: true},
		Coordinate{X: 216, Y: 152, PenUp: false},
		Coordinate{X: 218, Y: 152, PenUp: false},
		Coordinate{X: 221, Y: 152, PenUp: false},
		Coordinate{X: 223, Y: 154, PenUp: false},
		Coordinate{X: 213, Y: 152, PenUp: false},
	}
	result = p.Parse()
	assertAreEqual(expectedResult, result, t)
}

// assert that the two slices are equal
func assertAreEqual(expected, actual []Coordinate, t *testing.T) {

	if len(expected) != len(actual) {
		t.Error("[]Coordinate length difference", len(expected), "actual", len(actual))
	}

	for index := range expected {
		if !expected[index].Equals(actual[index]) {
			t.Error("Index", index, "expected", expected[index], "actual", actual[index])
		}
	}
}
