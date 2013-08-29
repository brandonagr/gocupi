package plotter

import (
	"fmt"
	"testing"
)

// Test that timeline generation is vertical
func Test_TimelineVertical(t *testing.T) {

	settings := &SettingsData{
		DrawingSurfaceMinY_MM:      1,
		DrawingSurfaceMaxY_MM:      9,
		DrawingSurfaceMinX_MM:      0,
		DrawingSurfaceMaxX_MM:      10,
		StepSize_MM:                1,
		StartingLeftDist_MM:        1,
		StartingRightDist_MM:       10.04987562112089027021926491276,
		SpoolHorizontalDistance_MM: 10,
	}

	plotCoords := make(chan Coordinate, 10)
	plotCoords <- Coordinate{X: 0, Y: 1}
	plotCoords <- Coordinate{X: 10, Y: 9}
	close(plotCoords)

	results := make(chan TimelineEvent, 100)

	go GenerateTimeline(plotCoords, results, settings)

	currentTime := 0.0
	for event := range results {
		if event.Time < currentTime {
			t.Error("Expected time to always grow, saw", event.Time, "after", currentTime)
		}
		currentTime = event.Time
		fmt.Println(event)
	}

	t.Fail()
}
