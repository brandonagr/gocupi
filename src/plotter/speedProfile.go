package plotter

import (
	"fmt"
)

// Calculates the time elapsed as progressing over a series of line segments
type SpeedProfile interface {
	// Returns true if there is another segment
	MoveNext() bool

	// The current segment that is moving over
	Current() LineSegment

	// Return a unit length vector in current direction of movement
	Dir() Coordinate

	// Update current location to be percentage along the current line segment
	UpdateStart(percentage float64)

	// Get the time elapsed after passing over the given percentage of the current line segment
	Time(percentage float64) float64
}

// Create new linear speed profile object
func NewLinearSpeedProfile(coords <-chan Coordinate, settings *SettingsData) SpeedProfile {
	fmt.Println("Test")
	return &LinearSpeedProfile{}
}

// Uses a constant speed for all movement
type LinearSpeedProfile struct {
	coords <-chan Coordinate
	speed  float64
}

func (this *LinearSpeedProfile) MoveNext() bool {
	return false
}

func (this *LinearSpeedProfile) Current() LineSegment {
	return LineSegment{}
}

func (this *LinearSpeedProfile) Dir() Coordinate {
	return Coordinate{}
}

func (this *LinearSpeedProfile) UpdateStart(percentage float64) {

}

func (this *LinearSpeedProfile) Time(percentge float64) float64 {
	return percentge
}
