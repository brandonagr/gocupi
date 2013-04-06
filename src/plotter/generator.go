package plotter

// Generates StepData either from a GetPosition func or from GCode data

import (
	"fmt"
	"image"
	"math"
)

// Generate spirograph
func GenerateParametric(posFunc func(float64) Coordinate, plotCoords chan<- Coordinate) {
	defer close(plotCoords)

	initialPosition := posFunc(0)
	moveDist := 100.0 * Settings.MaxSpeed_MM_S * Settings.TimeSlice_US / 10000000.0
	thetaDelta := (2.0 * math.Pi) / 2000 //(moveDist / setup.BigR) * 100.0
	numberSteps := 0

	theta := thetaDelta
	curPosition := posFunc(theta)
	previousPosition := curPosition
	plotCoords <- curPosition.Minus(initialPosition)

	for !curPosition.Equals(initialPosition) {

		numberSteps++
		if numberSteps > 100000000 {
			fmt.Println("Hitting", numberSteps, " step limit")
			break
		}

		theta += thetaDelta
		curPosition = posFunc(theta)

		if curPosition.Minus(previousPosition).Len() > moveDist {
			plotCoords <- curPosition.Minus(initialPosition)
			previousPosition = curPosition
		}
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
	plotCoords <- Coordinate{X: 0, Y: 0}
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
	displacement := Coordinate{X: setup.CircleDisplacement, Y: 0}

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
	plotCoords <- Coordinate{X: 0, Y: 0}
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

		plotCoords <- Coordinate{X: float64(x), Y: float64(y)}.Scaled(scale)
	}
	plotCoords <- Coordinate{X: 0, Y: 0}
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
		points[edge] = Coordinate{X: setup.Radius * math.Cos(angle), Y: setup.Radius * math.Sin(angle)}
	}

	for edge := 0; edge < edgeCountInt; edge++ {

		sourceBegin := points[edge]
		sourceEnd := Coordinate{X: 0, Y: 0}

		destBegin := Coordinate{X: 0, Y: 0}
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
		plotCoords <- Coordinate{X: 0, Y: 0}
	}
}

// Parameters for grid
type Grid struct {

	// Size of each axis
	Width float64

	// Number of cells to divide grid into, will be Cells x Cells total cells
	Cells float64
}

// Generate parabolic curve out of a bunch of straight lines
func GenerateGrid(setup Grid, plotCoords chan<- Coordinate) {
	defer close(plotCoords)

	cellInt := int(setup.Cells)
	cellWidth := setup.Width / setup.Cells

	for y := 0; y < cellInt; y++ {
		yf := float64(y)

		if y%2 == 0 {
			plotCoords <- Coordinate{X: setup.Cells * cellWidth, Y: yf * cellWidth}
			plotCoords <- Coordinate{X: setup.Cells * cellWidth, Y: (yf + 1) * cellWidth}
		} else {
			plotCoords <- Coordinate{X: 0, Y: yf * cellWidth}
			plotCoords <- Coordinate{X: 0, Y: (yf + 1) * cellWidth}
		}
	}

	plotCoords <- Coordinate{X: 0, Y: setup.Cells * cellWidth}

	for x := 0; x < cellInt; x++ {
		xf := float64(x)

		if x%2 == 0 {
			plotCoords <- Coordinate{X: xf * cellWidth, Y: 0}
			plotCoords <- Coordinate{X: (xf + 1) * cellWidth, Y: 0}
		} else {
			plotCoords <- Coordinate{X: xf * cellWidth, Y: setup.Cells * cellWidth}
			plotCoords <- Coordinate{X: (xf + 1) * cellWidth, Y: setup.Cells * cellWidth}
		}
	}

	if cellInt%2 == 0 {
		plotCoords <- Coordinate{X: setup.Cells * cellWidth, Y: 0}
	}
	plotCoords <- Coordinate{X: 0, Y: 0}
}

// Parameters for arc
type Arc struct {

	// Size of longest axis
	Size float64

	// Distance between arcs
	ArcDist float64
}

// Draw image by generate a series of arcs, where darknes of a pixel is a movement along the arc
func GenerateArc(setup Arc, imageData image.Image, plotCoords chan<- Coordinate) {
	defer close(plotCoords)

	imageSize := imageData.Bounds().Max
	scale := setup.Size / math.Max(float64(imageSize.X), float64(imageSize.Y))
	width := float64(imageSize.X) * scale
	height := float64(imageSize.Y) * scale
	fmt.Println("Width", width, "Height", height, "Scale", scale)

	polarSystem := PolarSystemFromSettings()
	polarPos := PolarCoordinate{LeftDist: Settings.StartingLeftDist_MM, RightDist: Settings.StartingRightDist_MM}
	startingPos := polarPos.ToCoord(polarSystem)

	arcOrigin := Coordinate{X: 0, Y: 0}.Minus(startingPos)

	beginRadius := arcOrigin.Len()
	endRadius := Coordinate{X: width, Y: height}.Minus(arcOrigin).Len()

	fmt.Println("Origin", arcOrigin, "beginRadius", beginRadius, "endRadius", endRadius)

	// sides of the drawing box
	sides := [4]LineSegment{
		LineSegment{Coordinate{X: 0, Y: 0}, Coordinate{X: width, Y: 0}},           // top
		LineSegment{Coordinate{X: width, Y: 0}, Coordinate{X: width, Y: height}},  // right
		LineSegment{Coordinate{X: width, Y: height}, Coordinate{X: 0, Y: height}}, // bottom
		LineSegment{Coordinate{X: 0, Y: height}, Coordinate{X: 0, Y: 0}},          // left
	}

	flipDir := false

	for radius := beginRadius + setup.ArcDist; radius < endRadius; radius += setup.ArcDist {

		// find two points of intersection
		var topIntersection, botIntersection Coordinate
		arc := Circle{arcOrigin, radius}

		for _, side := range sides {
			p1, p1Valid, _, _ := arc.Intersection(side)
			if p1Valid {
				if (topIntersection == Coordinate{X: 0, Y: 0}) {
					topIntersection = p1
				} else {
					botIntersection = p1
				}
			}
		}
		if topIntersection.Y > botIntersection.Y {
			topIntersection, botIntersection = botIntersection, topIntersection
		}

		//fmt.Println("Radius", radius, topIntersection, botIntersection)

		// need beginAngle, endAngle, increment
		topAngle := math.Atan2(topIntersection.Y-arcOrigin.Y, topIntersection.X-arcOrigin.X)
		botAngle := math.Atan2(botIntersection.Y-arcOrigin.Y, botIntersection.X-arcOrigin.X)

		thetaDelta := 3.0 * math.Asin(1.0/(2.0*radius))

		//fmt.Println("topAngle", topAngle, "botAngle", botAngle, "thetaDelta", thetaDelta)

		flipDir = !flipDir
		if flipDir {
			for theta := topAngle; theta <= botAngle; theta += thetaDelta {

				pos := arcOrigin.Add(Coordinate{X: math.Cos(theta) * radius, Y: math.Sin(theta) * radius})
				imageValue := 1.0 - sampleImageAt(imageData, pos.Scaled(1/scale))
				offset := setup.ArcDist * 0.5 * imageValue

				plotCoords <- pos
				if imageValue > 0.05 {
					plotCoords <- arcOrigin.Add(Coordinate{X: math.Cos(theta) * (radius + offset), Y: math.Sin(theta) * (radius + offset)})
					plotCoords <- arcOrigin.Add(Coordinate{X: math.Cos(theta+thetaDelta/2.0) * (radius + offset), Y: math.Sin(theta+thetaDelta/2.0) * (radius + offset)})
					plotCoords <- arcOrigin.Add(Coordinate{X: math.Cos(theta+thetaDelta/2.0) * (radius - offset), Y: math.Sin(theta+thetaDelta/2.0) * (radius - offset)})
					plotCoords <- arcOrigin.Add(Coordinate{X: math.Cos(theta+thetaDelta) * (radius - offset), Y: math.Sin(theta+thetaDelta) * (radius - offset)})
				}
			}
			plotCoords <- arcOrigin.Add(Coordinate{X: math.Cos(botAngle) * radius, Y: math.Sin(botAngle) * radius})
		} else {
			for theta := botAngle; theta >= topAngle; theta -= thetaDelta {
				pos := arcOrigin.Add(Coordinate{X: math.Cos(theta) * radius, Y: math.Sin(theta) * radius})
				imageValue := 1.0 - sampleImageAt(imageData, pos.Scaled(1/scale))
				offset := setup.ArcDist * 0.5 * imageValue

				plotCoords <- pos
				if imageValue > 0.05 {
					plotCoords <- arcOrigin.Add(Coordinate{X: math.Cos(theta) * (radius + offset), Y: math.Sin(theta) * (radius + offset)})
					plotCoords <- arcOrigin.Add(Coordinate{X: math.Cos(theta-thetaDelta/2.0) * (radius + offset), Y: math.Sin(theta-thetaDelta/2.0) * (radius + offset)})
					plotCoords <- arcOrigin.Add(Coordinate{X: math.Cos(theta-thetaDelta/2.0) * (radius - offset), Y: math.Sin(theta-thetaDelta/2.0) * (radius - offset)})
					plotCoords <- arcOrigin.Add(Coordinate{X: math.Cos(theta-thetaDelta) * (radius - offset), Y: math.Sin(theta-thetaDelta) * (radius - offset)})
				}
			}
			plotCoords <- arcOrigin.Add(Coordinate{X: math.Cos(topAngle) * radius, Y: math.Sin(topAngle) * radius})
		}
	}

	plotCoords <- Coordinate{X: width, Y: height}
	plotCoords <- Coordinate{X: 0, Y: height}
	plotCoords <- Coordinate{X: 0, Y: 0}
}

// Parameters for raster
type Raster struct {

	// Size of longest axis
	Size float64

	// Width of the pen used when filling in a pixel
	PenWidth float64

	// Size of a given pixel
	pixelSize float64
}

// Draw image by generate a series of arcs, where darknes of a pixel is a movement along the arc
func GenerateRaster(setup Raster, imageData image.Image, plotCoords chan<- Coordinate) {
	defer close(plotCoords)

	imageSize := imageData.Bounds().Max
	scale := setup.Size / math.Max(float64(imageSize.X), float64(imageSize.Y))
	setup.pixelSize = scale / 2.0
	width := float64(imageSize.X) * scale
	height := float64(imageSize.Y) * scale
	fmt.Println("Width", width, "Height", height, "Scale", scale)

	//polarSystem := PolarSystemFromSettings()
	//polarPos := PolarCoordinate{Settings.StartingLeftDist_MM, Settings.StartingRightDist_MM}
	//startingPos := polarPos.ToCoord(polarSystem)

	for y := 0; ; {

		for x := 0; x < imageSize.X; x++ {
			pos := Coordinate{X: float64(x) * scale, Y: float64(y) * scale}
			plotCoords <- pos

			//fmt.Println(x, y, average(imageData.At(x, y)))
			if average(imageData.At(x, y)) < 0.2 {
				drawPixel(pos, setup, plotCoords)
			}
		}

		y++
		if y == imageSize.Y {
			plotCoords <- Coordinate{X: 0, Y: height - scale}
			break
		}

		for x := imageSize.X - 1; x >= 0; x-- {
			pos := Coordinate{X: float64(x) * scale, Y: float64(y) * scale}
			plotCoords <- pos
			//fmt.Println(x, y, average(imageData.At(x, y)))
			if average(imageData.At(x, y)) < 0.2 {
				drawPixel(pos, setup, plotCoords)
			}
		}

		y++
		if y == imageSize.Y {
			break
		}
	}

	plotCoords <- Coordinate{X: 0, Y: 0}
}

// Draw a pixel at the given location
func drawPixel(center Coordinate, setup Raster, plotCoords chan<- Coordinate) {

	for currentBoxSize := setup.PenWidth; currentBoxSize <= setup.pixelSize; currentBoxSize += setup.PenWidth {
		plotCoords <- center.Minus(Coordinate{X: currentBoxSize, Y: currentBoxSize})
		plotCoords <- center.Minus(Coordinate{X: currentBoxSize, Y: -currentBoxSize})
		plotCoords <- center.Minus(Coordinate{X: -currentBoxSize, Y: -currentBoxSize})
		plotCoords <- center.Minus(Coordinate{X: -currentBoxSize, Y: currentBoxSize})
		plotCoords <- center.Minus(Coordinate{X: currentBoxSize, Y: currentBoxSize})
	}

	plotCoords <- center
}
