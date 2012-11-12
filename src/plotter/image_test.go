package plotter

// Tests for image contour path and helper functions

import (
	"image"
	"image/color"
	"math"
	"testing"
)

// Sample should correctly average values together
func TestSampleImageAt(t *testing.T) {
	r := image.Rectangle{image.Point{0, 0}, image.Point{2, 2}}
	data := image.NewRGBA(r)

	data.SetRGBA(0, 0, color.RGBA{0, 0, 0, 0})
	data.SetRGBA(1, 0, color.RGBA{64, 64, 64, 64})
	data.SetRGBA(0, 1, color.RGBA{128, 128, 128, 128})
	data.SetRGBA(1, 1, color.RGBA{191, 191, 191, 191})

	// sampling at exactly a coordinate should give the expected value
	assertAreClose(0, sampleImageAt(data, Coordinate{0, 0}), t)
	assertAreClose(0.25, sampleImageAt(data, Coordinate{1, 0}), t)
	assertAreClose(0.5, sampleImageAt(data, Coordinate{0, 1}), t)
	assertAreClose(0.75, sampleImageAt(data, Coordinate{1, 1}), t)

	// sampling at boundary of two points
	assertAreClose(0.125, sampleImageAt(data, Coordinate{0.5, 0}), t)
	assertAreClose(0.625, sampleImageAt(data, Coordinate{0.5, 1}), t)
	assertAreClose(0.25, sampleImageAt(data, Coordinate{0, 0.5}), t)
	assertAreClose(0.5, sampleImageAt(data, Coordinate{1, 0.5}), t)

	// sampling at middle
	assertAreClose(0.375, sampleImageAt(data, Coordinate{0.5, 0.5}), t)

	// samping at combination
	assertAreClose(0.0625, sampleImageAt(data, Coordinate{0.25, 0}), t)
	assertAreClose(0.5625, sampleImageAt(data, Coordinate{0.25, 1}), t)
}

// assert that the two values are equal
func assertAreClose(expected, actual float64, t *testing.T) {
	if math.Abs(expected-actual) > 0.01 {
		t.Error("Expected", expected, "and got", actual)
	}
}
