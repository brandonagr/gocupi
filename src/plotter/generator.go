package plotter
// Generates StepData either from a GetPosition func or from GCode data

import (
	"math"
	"fmt"
)

//// StepData, left then right # of steps
//type StepData []byte
//
//// An X,Y position
//type Position float64
//
//// Position tostring
//func (pos Position) String() string {
//	return fmt.Sprint(float64(pos), "mm")
//}
//
//// A function that takes time and returns a position
//type GetPosition func(percentage float64) Position
//
//// Given a value from 0 to 1, use cubic spline to smooth it out, where derivative at 0 and 1 is 0
//// See http://www.paulinternet.nl/?page=bicubic
//// results in ~50% higher max speed
//func CubicSmooth(x float64) float64 {
//	if x < 0 || x > 1 {
//		panic("Argument x out of range, must be between 0 and 1")
//	}
//
//	xSquared := x * x
//	return -2.0 * xSquared * x + 3 * xSquared
//}
//
//func QuadraticSmooth(x float64) float64 {
//	if x < 0 || x > 1 {
//		panic("Argument x out of range, must be between 0 and 1")
//	}
//
//	xCubed := x * x * x
//	return -(xCubed * x - 2*xCubed) / 2.0
//}
//
//// Evaluates the posFunc over time totalTime, generating necessary steps
//func GenStepProfile(totalTime time.Duration, positionCalculator GetPosition) (stepProfile []byte) {
//
//	totalSeconds := totalTime.Seconds()
//	stepProfile = make([]byte, 0, int(totalSeconds / float64(TIME_SLICE_US)))
//	previousActualPos := positionCalculator(0)
//
//	for curTime := TIME_SLICE_US; curTime <= totalTime; curTime += TIME_SLICE_US {
//
//		//smoothTime := CubicSmooth(curTime.Seconds() / totalSeconds) * totalSeconds
//		//smoothTime := QuadraticSmooth(curTime.Seconds() / totalSeconds) * totalSeconds
//		smoothTime := curTime.Seconds() / totalSeconds
//		newPos := positionCalculator(smoothTime)
//
//		//steps := int(math.Floor(float64(newPos - previousActualPos) / STEPSIZE_MM))
//
//		steps := int(float64(newPos - previousActualPos) / STEPSIZE_MM)
//
//		// Cap steps to the max, just hope that a future value will catch up, will probably want to panic in the case of 2 axis movement
//		if steps > 32 {
//			steps = 32
//		}
//		if steps < -32 {
//			steps = -32
//		}
//		previousActualPos = previousActualPos + Position(float64(steps) * STEPSIZE_MM);
//
//		var encodedSteps byte
//		if (steps < 0) {
//			encodedSteps = byte(-steps) | 0x80
//		} else {
//			encodedSteps = byte(steps)
//		}
//		stepProfile = append(stepProfile, encodedSteps)
//
//		//fmt.Println(curTime, "Target", newPos, "Actual", previousActualPos, "steps", steps)
//		fmt.Printf("%v\t%v\t%v\t%v\t%v\r\n", curTime, newPos, previousActualPos, steps, encodedSteps);
//	}
//
//	return
//}

// Evaluates the posFunc over time totalTime, generating necessary steps
func GenStepProfile(data GcodeData) (stepProfile []byte) {

	stepProfile = make([]byte, 0, 1000) // guess on size, it will be expanded if needed

	polarSystem := PolarSystemFromSettings()
	previousPolarPos := PolarCoordinate{ LeftDist: Settings.StartingLeftDist_MM, RightDist: Settings.StartingRightDist_MM }
	startingLocation := previousPolarPos.ToCoord(polarSystem)

	fmt.Println("Start Location", startingLocation, "Initial Polar", previousPolarPos)

	// setup 0,0 as the initial location of the plot head
	polarSystem.XOffset = startingLocation.X
	polarSystem.YOffset = startingLocation.Y

	origin := Coordinate{}
	previousActualPos := origin

	for _, curTarget := range data.Lines {

		target := curTarget.Dest
		targetVector := origin.Minus(target)

		actualDistance := targetVector.Len()
		idealTime := actualDistance / Settings.MaxSpeed_MM_S
		numberOfSlices := math.Ceil(idealTime / (Settings.TimeSlice_US / 1000000))

		for slice := 1.0; slice <= numberOfSlices; slice++ {
			percentageAlongLine := slice / numberOfSlices
			sliceTarget := origin.Add(targetVector.Scaled(percentageAlongLine))

			polarSliceTarget := sliceTarget.ToPolar(polarSystem)

			sliceSteps := previousPolarPos.Minus(polarSliceTarget).Scaled(1 / Settings.StepSize_MM)
			sliceSteps = sliceSteps.Ceil()//.Clamp(32, -32)

			previousPolarPos = previousPolarPos.Add(sliceSteps.Scaled(Settings.StepSize_MM))
			previousActualPos = previousPolarPos.ToCoord(polarSystem)

			//fmt.Println("Coord target", sliceTarget, "actual", previousActualPos, "Abs Actual", previousActualPos.Add(startingLocation))
			//fmt.Println("Polar target", polarSliceTarget, "actual", previousPolarPos)
			fmt.Println("Steps", sliceSteps, "Actual", previousActualPos)

			var encodedSteps byte
			if (sliceSteps.LeftDist < 0) {
				encodedSteps = byte(-sliceSteps.LeftDist)
			} else {
				encodedSteps = byte(sliceSteps.LeftDist) | 0x80
			}
			stepProfile = append(stepProfile, encodedSteps)
			if (sliceSteps.RightDist < 0) {
				encodedSteps = byte(-sliceSteps.RightDist) | 0x80
			} else {
				encodedSteps = byte(sliceSteps.RightDist)
			}
			stepProfile = append(stepProfile, encodedSteps)
		}
		origin = previousActualPos
		fmt.Println("NEXT STEP --------------------------------------------")
	}


//	totalSeconds := totalTime.Seconds()
//	stepProfile = make([]byte, 0, int(totalSeconds / float64(TIME_SLICE_US)))
//	previousActualPos := positionCalculator(0)
//
//	for curTime := TIME_SLICE_US; curTime <= totalTime; curTime += TIME_SLICE_US {
//
//		//smoothTime := CubicSmooth(curTime.Seconds() / totalSeconds) * totalSeconds
//		//smoothTime := QuadraticSmooth(curTime.Seconds() / totalSeconds) * totalSeconds
//		smoothTime := curTime.Seconds() / totalSeconds
//		newPos := positionCalculator(smoothTime)
//
//		//steps := int(math.Floor(float64(newPos - previousActualPos) / STEPSIZE_MM))
//
//		steps := int(float64(newPos - previousActualPos) / STEPSIZE_MM)
//
//		// Cap steps to the max, just hope that a future value will catch up, will probably want to panic in the case of 2 axis movement
//		if steps > 32 {
//			steps = 32
//		}
//		if steps < -32 {
//			steps = -32
//		}
//		previousActualPos = previousActualPos + Position(float64(steps) * STEPSIZE_MM);
//
//		var encodedSteps byte
//		if (steps < 0) {
//			encodedSteps = byte(-steps) | 0x80
//		} else {
//			encodedSteps = byte(steps)
//		}
//		stepProfile = append(stepProfile, encodedSteps)
//
//		//fmt.Println(curTime, "Target", newPos, "Actual", previousActualPos, "steps", steps)
//		fmt.Printf("%v\t%v\t%v\t%v\t%v\r\n", curTime, newPos, previousActualPos, steps, encodedSteps);
//	}

	return
}
