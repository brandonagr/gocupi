package plotter

// Tests for Polar

import (
	"testing"
)

// Minus should provide the correct result
func TestMinus(t *testing.T) {
	lhs := Coordinate{2, 2}
	rhs := Coordinate{1, 1}
	if !lhs.Minus(rhs).Equals(rhs) {
		t.Error("Unexpected result for lhs - rhs", lhs.Minus(rhs))
	}
}

// ToPolar should return expected result when converting from cartesian to polar
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

// ToCoord should return expected result when converting from polar to cartessian
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
