package plotter

// Takes a stream of plotCoords and generates

import (
	"fmt"
	"math"
)

// Convert MM to Steps
func ConvertToSteps(plotCoords <-chan Coordinate, stepCoords chan<- Coordinate) {
	defer close(stepCoords)
	mmToSteps := 1 / Settings.StepSize_MM

	for coord := range plotCoords {
		stepCoords <- coord.Scaled(mmToSteps)
	}
}

// Takes in coordinates and outputs a timeilne of step events
func GenerateTimeline(plotCoords <-chan Coordinate, timeEvents chan<- float64, settings *SettingsData) {

	defer close(timeEvents)

	polarSystem := PolarSystemFrom(settings)
	previousPolarPos := PolarCoordinate{LeftDist: settings.StartingLeftDist_MM, RightDist: settings.StartingRightDist_MM}
	startingLocation := previousPolarPos.ToCoord(polarSystem)
	mmToSteps := 1 / settings.StepSize_MM

	fmt.Println("Start Location", startingLocation, "Initial Polar", previousPolarPos)

	if startingLocation.IsNaN() {
		panic(fmt.Sprint("Starting location is not a valid number, your settings.xml has impossible values"))
	}

	// setup 0,0 as the initial location of the plot head
	polarSystem.XOffset = startingLocation.X
	polarSystem.YOffset = startingLocation.Y

	target, chanOpen := <-plotCoords
	if !chanOpen {
		return
	}
	target = target.Add(startingLocation).Scaled(mmToSteps)
	anotherTarget := true

	for anotherTarget {
		nextTarget, chanOpen := <-plotCoords
		nextTarget = nextTarget.Add(startingLocation).Scaled(mmToSteps)
		if !chanOpen {
			anotherTarget = false
			nextTarget = target
		}

		currentPolarPosSteps := previousPolarPos.Scaled(mmToSteps).Ceil()

		for {
			lineSegment := LineSegment{target, nextTarget}
			intersectionTime := math.MaxFloat64
			intersectionIndex := -1

			for index, circle := range []Circle{
				Circle{Radius: currentPolarPosSteps.LeftDist - 1},
				Circle{Radius: currentPolarPosSteps.LeftDist + 1},
				Circle{Radius: currentPolarPosSteps.RightDist - 1, Center: Coordinate{X: polarSystem.RightMotorDist * mmToSteps, Y: 0}},
				Circle{Radius: currentPolarPosSteps.RightDist + 1, Center: Coordinate{X: polarSystem.RightMotorDist * mmToSteps, Y: 0}}} {

				fmt.Println("Intersecting ", lineSegment, circle)

				if time, valid := circle.IntersectionTime(lineSegment); valid && time > 0 && time < intersectionTime {
					intersectionTime = time
					intersectionIndex = index
				}
			}

			if intersectionIndex == -1 {
				break
			}

			timeEvents <- intersectionTime

			fmt.Println("Intersection at", intersectionTime, target, "for", intersectionIndex)
			target = target.Add(nextTarget.Minus(target).Scaled(intersectionTime))
			fmt.Println("Moved to", target)

			fmt.Println("Polar from", currentPolarPosSteps)
			switch intersectionIndex {
			case 0:
				currentPolarPosSteps.LeftDist -= 1
				break
			case 1:
				currentPolarPosSteps.LeftDist += 1
				break
			case 2:
				currentPolarPosSteps.RightDist -= 1
				break
			case 3:
				currentPolarPosSteps.RightDist += 1
				break
			}
			fmt.Println("to", currentPolarPosSteps)

		}

		target = nextTarget
	}
	fmt.Println("Done generating timeline")
}
