package plotter

// Tests for Polar

import (
	"testing"
)

func TestToPolar(t *testing.T) {
	system := PolarSystem{
		XOffset:        3,
		YOffset:        4,
		MinXMotorDist:  0,
		YMin:           0,
		YMax:           8,
		RightMotorDist: 6,
	}

	coord := Coordinate{X: 0, Y: 0}
	polarCoord := coord.ToPolar(system)

	if polarCoord.LeftDist != 5 {
		t.Error("Unexpected value for LeftDist", polarCoord.LeftDist)
	}

	if polarCoord.RightDist != 5 {
		t.Error("Unexpected value for RightDist", polarCoord.RightDist)
	}
}

func TestToCoord(t *testing.T) {
	system := PolarSystem{
		XOffset:        3,
		YOffset:        4,
		MinXMotorDist:  0,
		YMin:           0,
		YMax:           8,
		RightMotorDist: 6,
	}

	polarCoord := PolarCoordinate{LeftDist: 5, RightDist: 5}
	coord := polarCoord.ToCoord(system)

	if coord.X != 0 {
		t.Error("Unexpected value for X", coord.X)
	}

	if coord.Y != 0 {
		t.Error("Unexpected value for Y", coord.Y)
	}
}
