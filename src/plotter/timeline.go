package plotter

// Takes a stream of plotCoords and generates

import (
	"fmt"
	"math"
)

type TimelineEvent struct {
	// Time that the event occured, offset from the beginning of time
	Time float64

	// Left motor movement
	LeftStep int8

	// Right motor movement
	RightStep int8
}

// TimelineEvent ToString
func (event TimelineEvent) String() string {

	return fmt.Sprintf("[ %.3f, %d, %d ]", event.Time, event.LeftStep, event.RightStep)
}

// Convert MM to Steps
func ConvertToSteps(plotCoords <-chan Coordinate, stepCoords chan<- Coordinate) {
	defer close(stepCoords)
	mmToSteps := 1 / Settings.StepSize_MM

	for coord := range plotCoords {
		stepCoords <- coord.Scaled(mmToSteps)
	}
}

// Takes in coordinates and outputs a timeilne of step events
func GenerateTimeline(plotCoords <-chan Coordinate, timeEvents chan<- TimelineEvent, settings *SettingsData) {

	defer close(timeEvents)

	polarSystem := PolarSystemFrom(settings)
	penPolarSteps := PolarCoordinate{LeftDist: settings.StartingLeftDist_MM, RightDist: settings.StartingRightDist_MM}
	startingLocation := penPolarSteps.ToCoord(polarSystem)

	fmt.Println("Start Location", startingLocation, "Initial Polar", penPolarSteps)

	if startingLocation.IsNaN() {
		panic(fmt.Sprint("Starting location is not a valid number, your settings.xml has impossible values"))
	}

	// setup 0,0 as the initial location of the plot head
	//polarSystem.XOffset = startingLocation.X
	//polarSystem.YOffset = startingLocation.Y

	speedProfile := NewLinearSpeedProfile(plotCoords, settings)
	minTolerance := 0.000001 // Min difference between points to move to it
	currentTime := 0.0

	for speedProfile.MoveNext() {

		for {
			lineSegment := speedProfile.Current()
			penPolarPos := lineSegment.Begin.ToPolar(polarSystem)

			intersectPercentage := math.MaxFloat64
			intersectIndex := -1

			var circle Circle

			//fmt.Println("LineSegment:", lineSegment)

			prevLeftDist := math.Floor(penPolarPos.LeftDist)
			if math.Abs(penPolarPos.LeftDist-prevLeftDist) < minTolerance {
				prevLeftDist -= 1
			}
			circle.Center = Coordinate{}
			circle.Radius = prevLeftDist
			if time, valid := circle.IntersectionTime(lineSegment); valid && time > minTolerance && time < intersectPercentage {
				intersectPercentage = time
				intersectIndex = 0
			}

			nextLeftDist := math.Ceil(penPolarPos.LeftDist)
			if math.Abs(penPolarPos.LeftDist-nextLeftDist) < minTolerance {
				nextLeftDist += 1
			}
			circle.Center = Coordinate{}
			circle.Radius = nextLeftDist
			if time, valid := circle.IntersectionTime(lineSegment); valid && time > minTolerance && time < intersectPercentage {
				intersectPercentage = time
				intersectIndex = 1
			}

			prevRightDist := math.Floor(penPolarPos.RightDist)
			if math.Abs(penPolarPos.RightDist-prevRightDist) < minTolerance {
				prevRightDist -= 1
			}
			circle.Center = Coordinate{X: settings.SpoolHorizontalDistance_MM}
			circle.Radius = prevRightDist
			if time, valid := circle.IntersectionTime(lineSegment); valid && time > minTolerance && time < intersectPercentage {
				intersectPercentage = time
				intersectIndex = 2
			}

			nextRightDist := math.Ceil(penPolarPos.RightDist)
			if math.Abs(penPolarPos.RightDist-nextRightDist) < minTolerance {
				nextRightDist += 1
			}
			circle.Center = Coordinate{X: settings.SpoolHorizontalDistance_MM}
			circle.Radius = nextRightDist
			if time, valid := circle.IntersectionTime(lineSegment); valid && time > minTolerance && time < intersectPercentage {
				intersectPercentage = time
				intersectIndex = 3
			}

			//fmt.Println("Pos", penPolarPos, prevLeftDist, nextLeftDist, prevRightDist, nextRightDist)

			if intersectIndex == -1 {
				//fmt.Println("failed to find any intersections, exiting")

				// need to add remaining time to reach the end of current segment
				currentTime += speedProfile.Time(1.0)
				break
			}

			currentTime += speedProfile.Time(intersectPercentage)
			speedProfile.UpdateStart(intersectPercentage)

			event := TimelineEvent{
				Time: currentTime,
			}

			switch intersectIndex {
			case 0:
				event.LeftStep = -1
				break
			case 1:
				event.LeftStep = 1
				break
			case 2:
				event.RightStep = -1
				break
			case 3:
				event.RightStep = 1
				break
			}

			timeEvents <- event
		}
	}

	fmt.Println("Done generating timeline")
}
