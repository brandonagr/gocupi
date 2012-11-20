package plotter

// Tests for PolarCoordinate and Coordinate

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

// Circle.Intersection(Line) should return expected results
func TestCircleLineIntersection(t *testing.T) {

	circle := Circle{Coordinate{0,0}, 5}
	line := LineSegment{Coordinate{0,0}, Coordinate{2,0}}

	p1, p1Valid, p2, p2Valid := circle.Intersection(line)
	if (p1Valid || p2Valid) {
		t.Error("Should have detected intersection for ", circle, line)
	}

	line = LineSegment{Coordinate{0,0}, Coordinate{6,0}}
	p1, p1Valid, p2, p2Valid = circle.Intersection(line)
	if (!p1Valid || p1 != Coordinate{5,0} || p2Valid || p2 != Coordinate{0,0}) {
		t.Error("Expected one intersection", p1, p2)
	}

	line = LineSegment{Coordinate{-5,0}, Coordinate{6,0}}
	p1, p1Valid, p2, p2Valid = circle.Intersection(line)
	if (!p1Valid || p1 != Coordinate{5,0} || !p2Valid || p2 != Coordinate{-5,0}) {
		t.Error("Expected two intersections", p1, p2)
	}

	line = LineSegment{Coordinate{5,0}, Coordinate{6,0}}
	p1, p1Valid, p2, p2Valid = circle.Intersection(line)
	if (!p1Valid || p1 != Coordinate{5,0} || p2Valid || p2 != Coordinate{0,0}) {
		t.Error("Expected one intersection", p1, p2)
	}

	line = LineSegment{Coordinate{-6,5}, Coordinate{6,5}}
	p1, p1Valid, p2, p2Valid = circle.Intersection(line)
	if (!p1Valid || p1 != Coordinate{0,5} || p2Valid || p2 != Coordinate{0,0}) {
		t.Error("Expected one intersection", p1, p2)
	}

	circle = Circle{Coordinate{5,0}, 5}
	line = LineSegment{Coordinate{0,0}, Coordinate{0,1}}
	p1, p1Valid, p2, p2Valid = circle.Intersection(line)
	if (!p1Valid || p1 != Coordinate{0,0} || p2Valid || p2 != Coordinate{0,0}) {
		t.Error("Expected one intersection", p1, p2)
	}

	line = LineSegment{Coordinate{5,0}, Coordinate{5,10}}
	p1, p1Valid, p2, p2Valid = circle.Intersection(line)
	if (!p1Valid || p1 != Coordinate{5,5} || p2Valid || p2 != Coordinate{0,0}) {
		t.Error("Expected one intersection", p1, p2)
	}

	line = LineSegment{Coordinate{5,-10}, Coordinate{5,10}}
	p1, p1Valid, p2, p2Valid = circle.Intersection(line)
	if (!p1Valid || p1 != Coordinate{5,5} || !p2Valid || p2 != Coordinate{5,-5}) {
		t.Error("Expected one intersection", p1, p2)
	}
}