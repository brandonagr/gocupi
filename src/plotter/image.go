package plotter

// Draws a series of coordinates to an image

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

// Draw coordinates to image
func DrawToImage(imageName string, plotCoords <-chan Coordinate) {

	// buffer all of the coordinates into a slice in order to figure out the min and max points to know how big the image needs to be
	points := make([]Coordinate, len(plotCoords))
	minPoint := Coordinate{100000, 100000}
	maxPoint := Coordinate{-100000, -10000}

	for point := range plotCoords {
		point = point.Scaled(4.0)
		points = append(points, point)

		if point.X < minPoint.X {
			minPoint.X = point.X
		} else if point.X > maxPoint.X {
			maxPoint.X = point.X
		}

		if point.Y < minPoint.Y {
			minPoint.Y = point.Y
		} else if point.Y > maxPoint.Y {
			maxPoint.Y = point.Y
		}
	}

	maxPoint = maxPoint.Add(Coordinate{50, 50})
	minPoint = minPoint.Add(Coordinate{-50, -50})

	image := image.NewRGBA(image.Rect(0, 0, int(maxPoint.X-minPoint.X), int(maxPoint.Y-minPoint.Y)))

	// plot each point in the image
	previousPoint := Coordinate{0, 0}
	for _, point := range points {
		//image.Set(int(point.X-minPoint.X), int(-(point.Y-minPoint.Y)+2*maxPoint.Y), color.RGBA{0, 0, 0, 255})
		drawLine(previousPoint, point, minPoint, maxPoint, image)

		previousPoint = point
	}

	file, err := os.OpenFile(imageName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if err = png.Encode(file, image); err != nil {
		panic(err)
	}
}

// Draw a line, from http://41j.com/blog/2012/09/bresenhams-line-drawing-algorithm-implemetations-in-go-and-c/
func drawLine(start Coordinate, end Coordinate, minPoint Coordinate, maxPoint Coordinate, image *image.RGBA) {

	start_x := int(start.X - minPoint.X)
	start_y := int(start.Y - minPoint.Y)
	end_x := int(end.X - minPoint.X)
	end_y := int(end.Y - minPoint.Y)

	//  highlight the end
	image.Set(end_x+1, end_y+1, color.RGBA{255, 0, 0, 128})
	image.Set(end_x+1, end_y-1, color.RGBA{255, 0, 0, 128})
	image.Set(end_x-1, end_y+1, color.RGBA{255, 0, 0, 128})
	image.Set(end_x-1, end_y-1, color.RGBA{255, 0, 0, 128})

	// Bresenham's
	cx := start_x
	cy := start_y

	dx := end_x - cx
	dy := end_y - cy
	if dx < 0 {
		dx = 0 - dx
	}
	if dy < 0 {
		dy = 0 - dy
	}

	var sx int
	var sy int
	if cx < end_x {
		sx = 1
	} else {
		sx = -1
	}
	if cy < end_y {
		sy = 1
	} else {
		sy = -1
	}
	err := dx - dy

	var n int
	for n = 0; n < 10000; n++ {

		image.Set(cx, cy, color.RGBA{0, 128, 0, 255})
		if (cx == end_x) && (cy == end_y) {
			return
		}
		e2 := 2 * err
		if e2 > (0 - dy) {
			err = err - dy
			cx = cx + sx
		}
		if e2 < dx {
			err = err + dx
			cy = cy + sy
		}
	}
}
