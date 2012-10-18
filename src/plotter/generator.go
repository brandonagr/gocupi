package plotter

// Generates StepData either from a GetPosition func or from GCode data

import (
	"fmt"
	"math"
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

// Given GCodeData, returns all of the 
func GenerateGcodePath(data GcodeData, plotCoords chan Coordinate) {

	defer close(plotCoords)

	for _, curTarget := range data.Lines {
		fmt.Println("Sending", curTarget.Dest)
		plotCoords <- curTarget.Dest
	}

}

// Parameters needed to generate a spiral
type Spiral struct {
	RadiusBegin, RadiusEnd, RadiusDeltaPerRev float64
}

// Generate a spiral
func GenerateSpiral(setup Spiral, plotCoords chan Coordinate) {

	defer close(plotCoords)

	// MM that will be moved in a single step, used to calc what the new position along spiral will be after one time slice
	moveDist := Settings.MaxSpeed_MM_S * Settings.TimeSlice_US / 1000000.0
	theta := 0.0

	for radius := setup.RadiusBegin; radius >= setup.RadiusEnd; {

		// use right triangle to approximate arc distance along spiral, ignores radiusDelta for this calculation
		thetaDelta := 2.0 * math.Asin(moveDist/(2.0*radius))
		theta += thetaDelta
		if theta >= 2.0*math.Pi {
			theta -= 2.0 * math.Pi
		}

		radiusDelta := setup.RadiusDeltaPerRev * thetaDelta / (2.0 * math.Pi)
		radius -= radiusDelta

		//fmt.Println("Radius", radius, "Radius delta", radiusDelta, "Theta", theta, "Theta delta", thetaDelta)

		plotCoords <- Coordinate{X: radius * math.Cos(theta), Y: radius * math.Sin(theta)}
	}
}
