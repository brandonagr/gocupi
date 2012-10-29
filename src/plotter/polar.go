package plotter

// Handles converting X,Y coordinate into polar coordinates

import (
	"fmt"
	"math"
)

// A Cartession coordinate or vector
type Coordinate struct {
	X, Y float64
}

// Coordinate ToString
func (coord Coordinate) String() string {
	return fmt.Sprintf("[ %.2f, %.2f ]", coord.X, coord.Y)
}

// Calculates length of vector
func (coord Coordinate) Len() float64 {
	return math.Sqrt(coord.X*coord.X + coord.Y*coord.Y)
}

// Add two coordinates together
func (source Coordinate) Add(dest Coordinate) Coordinate {
	return Coordinate{X: dest.X + source.X, Y: dest.Y + source.Y}
}

// Return the vector from source to dest
func (source Coordinate) Minus(dest Coordinate) Coordinate {
	return Coordinate{X: source.X - dest.X, Y: source.Y - dest.Y}
}

// Scales the Coordinate by the specified factor
func (coord Coordinate) Scaled(factor float64) Coordinate {
	return Coordinate{X: coord.X * factor, Y: coord.Y * factor}
}

// Apply math.Ceil to each value
func (coord Coordinate) Ceil() Coordinate {
	return Coordinate{X: math.Ceil(coord.X), Y: math.Ceil(coord.Y)}
}

// Clamp the values of X,Y to the given max/min
func (coord Coordinate) Clamp(max, min float64) Coordinate {
	return Coordinate{X: math.Min(max, math.Max(coord.X, min)), Y: math.Min(max, math.Max(coord.Y, min))}
}

// Normalize the vector
func (coord Coordinate) Normalized() Coordinate {
	len := coord.Len()
	return Coordinate{coord.X / len, coord.Y / len}
}

// Dot product between two vectors
func (coord Coordinate) DotProduct(other Coordinate) float64 {
	return coord.X*other.X + coord.Y*other.Y
}

// Test if the two coordinates are equal within a constant epsilon
func (coord Coordinate) Equals(other Coordinate) bool {
	diff := coord.Minus(other)
	return diff.Len() < 0.0001
}

// PolarSystem information, 0,0 is always the upper left motor
type PolarSystem struct {
	XOffset, YOffset float64 // The location of X,Y origin relative to the motors

	MinXMotorDist float64 // minimum amount of space from motors
	YMin, YMax    float64 // minimum vertical Y location

	RightMotorDist float64
}

// Create a PolarSystem from the settings object
func PolarSystemFromSettings() PolarSystem {
	return PolarSystem{
		XOffset:        0,
		YOffset:        0,
		MinXMotorDist:  0,
		YMin:           Settings.MinVertical_MM,
		YMax:           Settings.MaxVertical_MM,
		RightMotorDist: Settings.HorizontalDistance_MM,
	}
}

// A polar coordinate
type PolarCoordinate struct {
	LeftDist, RightDist float64
}

// Coordinate ToString
func (polarCoord PolarCoordinate) String() string {
	return fmt.Sprintf("( L %.2f, R %.2f )", polarCoord.LeftDist, polarCoord.RightDist)
}

// Add two coordinates together
func (source PolarCoordinate) Add(dest PolarCoordinate) PolarCoordinate {
	return PolarCoordinate{dest.LeftDist + source.LeftDist, dest.RightDist + source.RightDist}
}

// Return the vector from source to dest
func (source PolarCoordinate) Minus(dest PolarCoordinate) PolarCoordinate {
	return PolarCoordinate{source.LeftDist - dest.LeftDist, source.RightDist - dest.RightDist}
}

// Scales the PolarCoordinate bRightDist the specified factor
func (coord PolarCoordinate) Scaled(factor float64) PolarCoordinate {
	return PolarCoordinate{coord.LeftDist * factor, coord.RightDist * factor}
}

// ApplRightDist math.Ceil to each value
func (coord PolarCoordinate) Ceil() PolarCoordinate {
	return PolarCoordinate{math.Ceil(coord.LeftDist), math.Ceil(coord.RightDist)}
}

// Clamp the values of LeftDist,RightDist to the given maLeftDist/min
func (coord PolarCoordinate) Clamp(max, min float64) PolarCoordinate {
	return PolarCoordinate{math.Min(max, math.Max(coord.LeftDist, min)), math.Min(max, math.Max(coord.RightDist, min))}
}

// Convert the given coordinate from X,Y to polar in the given PolarSystem
func (coord Coordinate) ToPolar(system PolarSystem) (polarCoord PolarCoordinate) {

	coord.X += system.XOffset
	coord.Y += system.YOffset

	// clip coordinates to system's area
	coord.X = math.Max(coord.X, system.MinXMotorDist)
	coord.X = math.Min(coord.X, system.RightMotorDist-system.MinXMotorDist)
	coord.Y = math.Max(coord.Y, system.YMin)
	coord.Y = math.Min(coord.Y, system.YMax)

	//fmt.Println("Coord ToPolar", coord, system.RightMotorDist, system.MinXMotorDist)

	polarCoord.LeftDist = math.Sqrt(coord.X*coord.X + coord.Y*coord.Y)
	xDiff := system.RightMotorDist - coord.X
	polarCoord.RightDist = math.Sqrt(xDiff*xDiff + coord.Y*coord.Y)
	return
}

// Convert the given polarCoordinate from polar to X,Y in the given PolarSystem
func (polarCoord PolarCoordinate) ToCoord(system PolarSystem) (coord Coordinate) {

	coord.X = ((polarCoord.LeftDist * polarCoord.LeftDist) - (polarCoord.RightDist * polarCoord.RightDist) + (system.RightMotorDist * system.RightMotorDist)) / (2.0 * system.RightMotorDist)
	coord.Y = math.Sqrt((polarCoord.LeftDist * polarCoord.LeftDist) - (coord.X * coord.X))

	//fmt.Println("Polar ToCoord", polarCoord, system.RightMotorDist, coord)

	coord.X -= system.XOffset
	coord.Y -= system.YOffset

	return
}
