package plotter

// Generates StepData either from a GetPosition func or from GCode data

import (
	"fmt"
	"math"
)

// Given GCodeData, returns all of the 
func GenerateGcodePath(data GcodeData, scale float64, plotCoords chan<- Coordinate) {

	defer close(plotCoords)

	for _, curTarget := range data.Lines {
		fmt.Println("Sending", curTarget.Dest)
		plotCoords <- curTarget.Dest.Scaled(scale)
	}

}

// Parameters needed to generate a spiral
type Spiral struct {
	// Initial radius of spiral
	RadiusBegin float64

	// Spiral ends when this radius is reached
	RadiusEnd float64

	// How much each revolution changes the radius
	RadiusDeltaPerRev float64
}

// Generate a spiral
func GenerateSpiral(setup Spiral, plotCoords chan<- Coordinate) {

	defer close(plotCoords)

	// MM that will be moved in a single step, used to calc what the new position along spiral will be after one time slice
	moveDist := Settings.MaxSpeed_MM_S * Settings.TimeSlice_US / 1000000.0
	theta := 0.0

	for radius := setup.RadiusBegin; radius >= setup.RadiusEnd; {

		plotCoords <- Coordinate{X: radius * math.Cos(theta), Y: radius * math.Sin(theta)}

		// use right triangle to approximate arc distance along spiral, ignores radiusDelta for this calculation
		thetaDelta := 2.0 * math.Asin(moveDist/(2.0*radius))
		theta += thetaDelta
		if theta >= 2.0*math.Pi {
			theta -= 2.0 * math.Pi
		}

		radiusDelta := setup.RadiusDeltaPerRev * thetaDelta / (2.0 * math.Pi)
		radius -= radiusDelta

		//fmt.Println("Radius", radius, "Radius delta", radiusDelta, "Theta", theta, "Theta delta", thetaDelta)
	}
	plotCoords <- Coordinate{0,0}
}

// Parameters needed to generate a sliding circle
type SlidingCircle struct {

	// Radius of the circle
	Radius float64

	// Distance traveled while one circle is traced
	CircleDisplacement Coordinate

	// Number of cirlces that will be drawn
	NumbCircles int
}

// Generate a circle that slides along a given axis
func GenerateSlidingCircle(setup SlidingCircle, plotCoords chan<- Coordinate) {

	defer close(plotCoords)

	// MM that will be moved in a single step, used to calc what the new position along spiral will be after one time slice
	moveDist := Settings.MaxSpeed_MM_S * Settings.TimeSlice_US / 1000000.0

	theta := 0.0
	thetaDelta := 2.8 * math.Asin(moveDist/(2.0*setup.Radius))
	origin := Coordinate{X: 0.0, Y: 0.0}

	for drawnCircles := 0; drawnCircles < setup.NumbCircles; {

		circlePos := Coordinate{X: setup.Radius * math.Cos(theta), Y: setup.Radius * math.Sin(theta)}
		origin = origin.Add(setup.CircleDisplacement.Scaled(thetaDelta / (2.0 * math.Pi)))

		plotCoords <- circlePos.Add(origin)

		theta += thetaDelta
		if theta > 2.0*math.Pi {
			theta -= 2.0 * math.Pi
			drawnCircles++
		}
	}
	plotCoords <- Coordinate{0,0}
}

// Parameters needed for hilbert curve
type HilbertCurve struct {

	// Integer of how complex it is, 2^Degree is the order
	Degree int

	// Total size in mm
	Size float64
}

// Generate a Hilbert curve, based on code from http://en.wikipedia.org/wiki/Hilbert_curve
func GenerateHilbertCurve(setup HilbertCurve, plotCoords chan<- Coordinate) {

	defer close(plotCoords)

	order := int(math.Pow(2, float64(setup.Degree)))
	dimSize := order << 1
	length := dimSize * dimSize
	scale := setup.Size / float64(dimSize)

	fmt.Println("Hilbert DimSize", dimSize, "Length", length)

	for hilbertIndex := 0; hilbertIndex < length; hilbertIndex++ {
		var x, y int
		hilbert_d2xy(dimSize, hilbertIndex, &x, &y)

		plotCoords <- Coordinate{float64(x), float64(y)}.Scaled(scale)
	}
	plotCoords <- Coordinate{0,0}
}

//convert d to (x,y)
func hilbert_d2xy(n int, d int, x *int, y *int) {
	var rx, ry, s, t int
	t = d
	*x = 0
	*y = 0
	for s = 1; s < n; s *= 2 {
		rx = 1 & (t / 2)
		ry = 1 & (t ^ rx)
		hilbert_rot(s, x, y, rx, ry)
		*x += s * rx
		*y += s * ry
		t /= 4
	}
}

//rotate/flip a quadrant appropriately
func hilbert_rot(n int, x *int, y *int, rx int, ry int) {
	if ry == 0 {
		if rx == 1 {
			*x = n - 1 - *x
			*y = n - 1 - *y
		}

		//Swap x and y
		*x, *y = *y, *x
	}
}

// Parameters for parabolic curve
type Parabolic struct {

	// Height of each axis
	Height float64

	// Number of lines
	Lines int
}

// Generate parabolic curv
func GenerateParabolic(setup Parabolic, plotCoords chan<- Coordinate) {

	defer close(plotCoords)

	delta := setup.Height / float64(setup.Lines)

	for lineIndex := 0; lineIndex < setup.Lines; lineIndex++ {

		if lineIndex % 2 == 0 {
			plotCoords <- Coordinate{0, delta * float64(lineIndex)}
			plotCoords <- Coordinate{delta * float64(lineIndex+1), setup.Height}
		} else {
			plotCoords <- Coordinate{delta * float64(lineIndex+1), setup.Height}
			plotCoords <- Coordinate{0, delta * float64(lineIndex)} 
		}
	}
	plotCoords <- Coordinate{0,0}

	for lineIndex := 0; lineIndex < setup.Lines; lineIndex++ {

		if lineIndex % 2 == 0 {
			plotCoords <- Coordinate{delta * float64(lineIndex), 0}
			plotCoords <- Coordinate{setup.Height, delta * float64(lineIndex+1)}
		} else {
			plotCoords <- Coordinate{setup.Height, delta * float64(lineIndex+1)}
			plotCoords <- Coordinate{delta * float64(lineIndex), 0}
		}
	}
	plotCoords <- Coordinate{0,0}
}
