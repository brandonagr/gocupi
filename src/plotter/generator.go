package plotter

// Generates StepData either from a GetPosition func or from GCode data

import (
	"fmt"
	"math"
)

// Generate spirograph
func GenerateParametric(posFunc func(float64) Coordinate, plotCoords chan<- Coordinate) {
	defer close(plotCoords)

	initialPosition := posFunc(0)
	//moveDist := Settings.MaxSpeed_MM_S * Settings.TimeSlice_US / 10000000.0
	thetaDelta := (2.0 * math.Pi) / 4000 //(moveDist / setup.BigR) * 100.0
	numberSteps := 0

	theta := thetaDelta
	curPosition := posFunc(theta)
	plotCoords <- curPosition.Minus(initialPosition)

	for !curPosition.Equals(initialPosition) {

		numberSteps++
		if numberSteps > 100000000 {
			fmt.Println("Hitting", numberSteps, " step limit")
			break
		}

		theta += thetaDelta

		curPosition = posFunc(theta)
		plotCoords <- curPosition.Minus(initialPosition)
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
	moveDist := 4.0 * Settings.MaxSpeed_MM_S * Settings.TimeSlice_US / 1000000.0
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
	plotCoords <- Coordinate{0, 0}
}

// Parameters needed to generate a sliding circle
type SlidingCircle struct {

	// Radius of the circle
	Radius float64

	// Distance traveled while one circle is traced
	CircleDisplacement float64

	// Number of cirlces that will be drawn
	NumbCircles int
}

// Generate a circle that slides along a given axis
func GenerateSlidingCircle(setup SlidingCircle, plotCoords chan<- Coordinate) {

	defer close(plotCoords)

	// MM that will be moved in a single step, used to calc what the new position along spiral will be after one time slice
	moveDist := 4.0 * Settings.MaxSpeed_MM_S * Settings.TimeSlice_US / 1000000.0
	displacement := Coordinate{setup.CircleDisplacement, 0}

	theta := 0.0
	thetaDelta := 2.8 * math.Asin(moveDist/(2.0*setup.Radius))
	origin := Coordinate{X: 0.0, Y: 0.0}

	for drawnCircles := 0; drawnCircles < setup.NumbCircles; {

		circlePos := Coordinate{X: setup.Radius * math.Cos(theta), Y: setup.Radius * math.Sin(theta)}
		origin = origin.Add(displacement.Scaled(thetaDelta / (2.0 * math.Pi)))

		plotCoords <- circlePos.Add(origin)

		theta += thetaDelta
		if theta > 2.0*math.Pi {
			theta -= 2.0 * math.Pi
			drawnCircles++
		}
	}
	plotCoords <- Coordinate{0, 0}
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
	plotCoords <- Coordinate{0, 0}
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
	Radius float64

	// number of faces on polygon
	PolygonEdgeCount float64

	// Number of lines
	Lines float64
}

// Generate parabolic curve out of a bunch of straight lines
func GenerateParabolic(setup Parabolic, plotCoords chan<- Coordinate) {

	defer close(plotCoords)

	edgeCountInt := int(setup.PolygonEdgeCount)
	linesCount := int(setup.Lines)

	points := make([]Coordinate, edgeCountInt)
	for edge := 0; edge < edgeCountInt; edge++ {
		angle := ((2.0 * math.Pi) / setup.PolygonEdgeCount) * float64(edge)
		points[edge] = Coordinate{setup.Radius * math.Cos(angle), setup.Radius * math.Sin(angle)}
	}

	for edge := 0; edge < edgeCountInt; edge++ {

		sourceBegin := points[edge]
		sourceEnd := Coordinate{0, 0}

		destBegin := Coordinate{0, 0}
		destEnd := points[(edge+1)%edgeCountInt]

		//fmt.Println("Source", sourceBegin, sourceEnd, "Dest", destBegin, destEnd)

		for lineIndex := 0; lineIndex < linesCount; lineIndex++ {
			startPercentage := float64(lineIndex) / float64(linesCount)
			endPercentage := float64(lineIndex+1) / float64(linesCount)

			//fmt.Println("Line", lineIndex, "StartFactor", startPercentage, "EndFactor", endPercentage)

			start := sourceEnd.Minus(sourceBegin).Scaled(startPercentage).Add(sourceBegin)
			end := destEnd.Minus(destBegin).Scaled(endPercentage).Add(destBegin)

			if lineIndex%2 == 0 {
				plotCoords <- start
				plotCoords <- end
			} else {
				plotCoords <- end
				plotCoords <- start
			}
		}
		plotCoords <- Coordinate{0, 0}
	}

}
