package plotter

import ()

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

	firstPos, chanOpen := <-coords
	if !chanOpen {
		return nil
	}

	return &LinearSpeedProfile{
		end:    firstPos, // will get moved into begin on call to MoveNext
		coords: coords,
		speed:  settings.MaxSpeed_MM_S,
	}
}

// Uses a constant speed for all movement
type LinearSpeedProfile struct {
	begin  Coordinate
	end    Coordinate
	coords <-chan Coordinate
	speed  float64
}

func (this *LinearSpeedProfile) MoveNext() bool {
	this.begin = this.end
	var chanOpen bool
	this.end, chanOpen = <-this.coords
	return chanOpen
}

func (this *LinearSpeedProfile) Current() LineSegment {
	return LineSegment{this.begin, this.end}
}

func (this *LinearSpeedProfile) Dir() Coordinate {
	return this.end.Minus(this.begin)
}

func (this *LinearSpeedProfile) UpdateStart(percentage float64) {
	this.begin = this.begin.Add(this.Dir().Scaled(percentage))
}

func (this *LinearSpeedProfile) Time(percentage float64) float64 {
	return (this.Dir().Len() * percentage) / this.speed
}
