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

	start, chanOpen := <-plotCoords
	if !chanOpen {
		return
	}
	anotherSegment := true

	for anotherSegment {
		end, chanOpen := <-plotCoords
		if !chanOpen {
			anotherSegment = false
			end = start
		}

		// Min difference between points to move to it
		minTolerance := 0.000001

		for {
			penPolarPos := start.ToPolar(polarSystem)

			intersectTime := math.MaxFloat64
			intersectIndex := -1
			lineSegment := LineSegment{start, end}
			var circle Circle

			fmt.Println("LineSegment:", lineSegment)

			prevLeftDist := math.Floor(penPolarPos.LeftDist)
			if math.Abs(penPolarPos.LeftDist-prevLeftDist) < minTolerance {
				prevLeftDist -= 1
			}
			circle.Center = Coordinate{}
			circle.Radius = prevLeftDist
			if time, valid := circle.IntersectionTime(lineSegment); valid && time > minTolerance && time < intersectTime {
				intersectTime = time
				intersectIndex = 0
			}

			nextLeftDist := math.Ceil(penPolarPos.LeftDist)
			if math.Abs(penPolarPos.LeftDist-nextLeftDist) < minTolerance {
				nextLeftDist += 1
			}
			circle.Center = Coordinate{}
			circle.Radius = nextLeftDist
			if time, valid := circle.IntersectionTime(lineSegment); valid && time > minTolerance && time < intersectTime {
				intersectTime = time
				intersectIndex = 1
			}

			prevRightDist := math.Floor(penPolarPos.RightDist)
			if math.Abs(penPolarPos.RightDist-prevRightDist) < minTolerance {
				prevRightDist -= 1
			}
			circle.Center = Coordinate{X: settings.SpoolHorizontalDistance_MM}
			circle.Radius = prevRightDist
			if time, valid := circle.IntersectionTime(lineSegment); valid && time > minTolerance && time < intersectTime {
				intersectTime = time
				intersectIndex = 2
			}

			nextRightDist := math.Ceil(penPolarPos.RightDist)
			if math.Abs(penPolarPos.RightDist-nextRightDist) < minTolerance {
				nextRightDist += 1
			}
			circle.Center = Coordinate{X: settings.SpoolHorizontalDistance_MM}
			circle.Radius = nextRightDist
			if time, valid := circle.IntersectionTime(lineSegment); valid && time > minTolerance && time < intersectTime {
				intersectTime = time
				intersectIndex = 3
			}

			fmt.Println("Pos", penPolarPos, prevLeftDist, nextLeftDist, prevRightDist, nextRightDist)

			if intersectIndex == -1 {
				fmt.Println("failed to find any intersections, exiting")
				break
			}

			intersectPos := start.Add(end.Minus(start).Scaled(intersectTime))

			fmt.Println("From", start, "to", end, "intersect", intersectTime, "at", intersectPos)

			start = intersectPos

			switch intersectIndex {
			case 0:
				//currentPolarPosSteps.LeftDist -= 1
				fmt.Println("Left -= 1")
				break
			case 1:
				//currentPolarPosSteps.LeftDist += 1
				fmt.Println("Left += 1")
				break
			case 2:
				//currentPolarPosSteps.RightDist -= 1
				fmt.Println("Right -= 1")
				break
			case 3:
				//currentPolarPosSteps.RightDist += 1
				fmt.Println("Right += 1")
				break
			}

			timeEvents <- TimelineEvent{
				Time:      intersectTime,
				LeftStep:  0,
				RightStep: 0,
			}
		}

		start = end
	}

	fmt.Println("Done generating timeline")
}
