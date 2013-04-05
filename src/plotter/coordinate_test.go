package plotter

// Tests for PolarCoordinate and Coordinate

import (
	"testing"
)

// Minus should provide the correct result
func TestMinus(t *testing.T) {
	lhs := Coordinate{X: 2, Y: 2}
	rhs := Coordinate{X: 1, Y: 1}
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

	coord := Coordinate{X: 0, Y: 0, PenUp: true}
	polarCoord := coord.ToPolar(system)

	if polarCoord.LeftDist != 5 {
		t.Error("Unexpected value for LeftDist", polarCoord.LeftDist)
	}

	if polarCoord.RightDist != 5 {
		t.Error("Unexpected value for RightDist", polarCoord.RightDist)
	}

	if !polarCoord.PenUp {
		t.Error("Unexpected value for PenUp", polarCoord.PenUp)
	}

	coord = Coordinate{X: 0, Y: 0, PenUp: false}
	polarCoord = coord.ToPolar(system)

	if polarCoord.PenUp {
		t.Error("Unexpected value for PenUp", polarCoord.PenUp)
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

	polarCoord := PolarCoordinate{LeftDist: 5, RightDist: 5, PenUp: true}
	coord := polarCoord.ToCoord(system)

	if coord.X != 0 {
		t.Error("Unexpected value for X", coord.X)
	}

	if coord.Y != 0 {
		t.Error("Unexpected value for Y", coord.Y)
	}

	if !coord.PenUp {
		t.Error("Unexpected value for PenUp", coord.PenUp)
	}

	polarCoord = PolarCoordinate{LeftDist: 5, RightDist: 5, PenUp: false}
	coord = polarCoord.ToCoord(system)

	if coord.PenUp {
		t.Error("Unexpected value for PenUp", coord.PenUp)
	}
}

// Circle.Intersection(Line) should return expected results
func TestCircleLineIntersection(t *testing.T) {

	circle := Circle{Coordinate{X: 0, Y: 0}, 5}
	line := LineSegment{Coordinate{X: 0, Y: 0}, Coordinate{X: 2, Y: 0}}

	p1, p1Valid, p2, p2Valid := circle.Intersection(line)
	if p1Valid || p2Valid {
		t.Error("Should have detected intersection for ", circle, line)
	}

	line = LineSegment{Coordinate{X: 0, Y: 0}, Coordinate{X: 6, Y: 0}}
	p1, p1Valid, p2, p2Valid = circle.Intersection(line)
	if (!p1Valid || p1 != Coordinate{X: 5, Y: 0} || p2Valid || p2 != Coordinate{X: 0, Y: 0}) {
		t.Error("Expected one intersection", p1, p2)
	}

	line = LineSegment{Coordinate{X: -5, Y: 0}, Coordinate{X: 6, Y: 0}}
	p1, p1Valid, p2, p2Valid = circle.Intersection(line)
	if (!p1Valid || p1 != Coordinate{X: 5, Y: 0} || !p2Valid || p2 != Coordinate{X: -5, Y: 0}) {
		t.Error("Expected two intersections", p1, p2)
	}

	line = LineSegment{Coordinate{X: 5, Y: 0}, Coordinate{X: 6, Y: 0}}
	p1, p1Valid, p2, p2Valid = circle.Intersection(line)
	if (!p1Valid || p1 != Coordinate{X: 5, Y: 0} || p2Valid || p2 != Coordinate{X: 0, Y: 0}) {
		t.Error("Expected one intersection", p1, p2)
	}

	line = LineSegment{Coordinate{X: -6, Y: 5}, Coordinate{X: 6, Y: 5}}
	p1, p1Valid, p2, p2Valid = circle.Intersection(line)
	if (!p1Valid || p1 != Coordinate{X: 0, Y: 5} || p2Valid || p2 != Coordinate{X: 0, Y: 0}) {
		t.Error("Expected one intersection", p1, p2)
	}

	circle = Circle{Coordinate{X: 5, Y: 0}, 5}
	line = LineSegment{Coordinate{X: 0, Y: 0}, Coordinate{X: 0, Y: 1}}
	p1, p1Valid, p2, p2Valid = circle.Intersection(line)
	if (!p1Valid || p1 != Coordinate{X: 0, Y: 0} || p2Valid || p2 != Coordinate{X: 0, Y: 0}) {
		t.Error("Expected one intersection", p1, p2)
	}

	line = LineSegment{Coordinate{X: 5, Y: 0}, Coordinate{X: 5, Y: 10}}
	p1, p1Valid, p2, p2Valid = circle.Intersection(line)
	if (!p1Valid || p1 != Coordinate{X: 5, Y: 5} || p2Valid || p2 != Coordinate{X: 0, Y: 0}) {
		t.Error("Expected one intersection", p1, p2)
	}

	line = LineSegment{Coordinate{X: 5, Y: -10}, Coordinate{X: 5, Y: 10}}
	p1, p1Valid, p2, p2Valid = circle.Intersection(line)
	if (!p1Valid || p1 != Coordinate{X: 5, Y: 5} || !p2Valid || p2 != Coordinate{X: 5, Y: -5}) {
		t.Error("Expected one intersection", p1, p2)
	}
}
