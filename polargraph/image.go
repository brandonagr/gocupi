package polargraph

// Draws a series of coordinates to an image

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
)

// Draw coordinates to image
func DrawToImage(imageName string, plotCoords <-chan Coordinate) {

	// buffer all of the coordinates into a slice in order to figure out the min and max points to know how big the image needs to be
	points := make([]Coordinate, len(plotCoords))
	minPoint := Coordinate{X: 100000, Y: 100000}
	maxPoint := Coordinate{X: -100000, Y: -10000}

	for point := range plotCoords {
		point = point.Scaled(4.0) // 4 pixels = 1mm
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

	// add some border to the image
	maxPoint = maxPoint.Add(Coordinate{X: 50, Y: 50})
	minPoint = minPoint.Add(Coordinate{X: -50, Y: -50})

	fmt.Println("Image Min:", minPoint, "Max:", maxPoint)

	image := image.NewRGBA(image.Rect(0, 0, int(maxPoint.X-minPoint.X), int(maxPoint.Y-minPoint.Y)))

	// plot each point in the image
	previousPoint := Coordinate{X: 0, Y: 0}
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
	var lineColor color.RGBA
	if end.PenUp {
		lineColor = color.RGBA{220, 220, 220, 255}
	} else {
		lineColor = color.RGBA{0, 0, 255, 255}
	}

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

		image.Set(cx, cy, lineColor)
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

// Load image data
func LoadImage(imageFileName string) image.Image {

	file, err := os.Open(imageFileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	image, format, err := image.Decode(file)
	if err != nil {
		panic(err)
	}

	fmt.Println("Loaded", format, "Size", image.Bounds())

	return image
	//return SobelImage(image) // turns out that runnign edge detection first does NOT help the image, it only removes information from the picture
}

// Return a new image that is the result of applying the sobel filter on the image
func SobelImage(imageData image.Image) image.Image {

	fmt.Println("Applying edge detection...")

	filtered := image.NewGray16(imageData.Bounds())
	imageSize := imageData.Bounds().Max

	for yPixel := 0; yPixel < imageSize.Y; yPixel++ {
		filtered.SetGray16(0, yPixel, color.Gray16{32768})
		filtered.SetGray16(imageSize.X-1, yPixel, color.Gray16{32768})
	}
	for xPixel := 0; xPixel < imageSize.X; xPixel++ {
		filtered.SetGray16(xPixel, 0, color.Gray16{32768})
		filtered.SetGray16(xPixel, imageSize.Y-1, color.Gray16{32768})
	}

	minColor := uint16(65535)
	maxColor := uint16(0)

	for yPixel := 1; yPixel < imageSize.Y-1; yPixel++ {
		for xPixel := 1; xPixel < imageSize.X-1; xPixel++ {

			var total float64
			total -= average(imageData.At(xPixel-1, yPixel-1))
			total -= 2 * average(imageData.At(xPixel-1, yPixel))
			total -= average(imageData.At(xPixel-1, yPixel+1))

			total += average(imageData.At(xPixel+1, yPixel-1))
			total += 2 * average(imageData.At(xPixel+1, yPixel))
			total += average(imageData.At(xPixel+1, yPixel+1))

			total /= 6

			newColor := uint16(total*32768 + 32768)
			filtered.SetGray16(xPixel, yPixel, color.Gray16{newColor})

			if newColor > maxColor {
				maxColor = newColor
			} else if newColor < minColor {
				minColor = newColor
			}
		}
	}

	// scale image so contrast is maximized
	scale := float64(maxColor-minColor) / 65535.0
	//fmt.Println("Max", maxColor, "Min", minColor, "Scale", scale)

	for yPixel := 1; yPixel < imageSize.Y-1; yPixel++ {
		for xPixel := 1; xPixel < imageSize.X-1; xPixel++ {
			oldColor, _, _, _ := filtered.At(xPixel, yPixel).RGBA()

			newColor := uint16((float64(oldColor) - float64(minColor)) / scale)
			//fmt.Println("From", oldColor, "to", newColor)
			filtered.SetGray16(xPixel, yPixel, color.Gray16{newColor})
		}
	}

	// dump test image to disk
	// file, err := os.OpenFile("test.png", os.O_CREATE|os.O_WRONLY, 0666)
	// if err != nil {
	// 	panic(err)
	// }
	// defer file.Close()
	// if err = png.Encode(file, filtered); err != nil {
	// 	panic(err)
	// }

	return filtered
}

// Return a new image that is the result of applying a 3x3 gaussion blur to the image
func GaussianImage(imageData image.Image) image.Image {

	fmt.Println("Applying gaussian blur...")

	filtered := image.NewGray16(imageData.Bounds())
	imageSize := imageData.Bounds().Max

	for yPixel := 0; yPixel < imageSize.Y; yPixel++ {
		newColor := uint16(average(imageData.At(0, yPixel)) * 65535)
		filtered.SetGray16(0, yPixel, color.Gray16{newColor})

		newColor = uint16(average(imageData.At(imageSize.X-1, yPixel)) * 65535)
		filtered.SetGray16(imageSize.X-1, yPixel, color.Gray16{newColor})
	}
	for xPixel := 0; xPixel < imageSize.X; xPixel++ {
		newColor := uint16(average(imageData.At(xPixel, 0)) * 65535)
		filtered.SetGray16(xPixel, 0, color.Gray16{newColor})

		newColor = uint16(average(imageData.At(xPixel, imageSize.Y-1)) * 65535)
		filtered.SetGray16(xPixel, imageSize.Y-1, color.Gray16{newColor})
	}

	minColor := uint16(65535)
	maxColor := uint16(0)

	for yPixel := 1; yPixel < imageSize.Y-1; yPixel++ {
		for xPixel := 1; xPixel < imageSize.X-1; xPixel++ {

			var total float64
			total += 1 * average(imageData.At(xPixel-1, yPixel-1))
			total += 2 * average(imageData.At(xPixel-1, yPixel))
			total += 1 * average(imageData.At(xPixel-1, yPixel+1))

			total += 2 * average(imageData.At(xPixel, yPixel-1))
			total += 4 * average(imageData.At(xPixel, yPixel))
			total += 2 * average(imageData.At(xPixel, yPixel+1))

			total += 1 * average(imageData.At(xPixel+1, yPixel-1))
			total += 2 * average(imageData.At(xPixel+1, yPixel))
			total += 1 * average(imageData.At(xPixel+1, yPixel+1))

			total /= 16

			newColor := uint16(total * 65535)
			filtered.SetGray16(xPixel, yPixel, color.Gray16{newColor})

			if newColor > maxColor {
				maxColor = newColor
			} else if newColor < minColor {
				minColor = newColor
			}
		}
	}

	// scale image so contrast is maximized
	scale := float64(maxColor-minColor) / 65535.0
	//fmt.Println("Max", maxColor, "Min", minColor, "Scale", scale)

	for yPixel := 1; yPixel < imageSize.Y-1; yPixel++ {
		for xPixel := 1; xPixel < imageSize.X-1; xPixel++ {
			oldColor, _, _, _ := filtered.At(xPixel, yPixel).RGBA()

			newColor := uint16((float64(oldColor) - float64(minColor)) / scale)
			//fmt.Println("From", oldColor, "to", newColor)
			filtered.SetGray16(xPixel, yPixel, color.Gray16{newColor})
		}
	}

	// dump test image to disk
	// file, err := os.OpenFile("test.png", os.O_CREATE|os.O_WRONLY, 0666)
	// if err != nil {
	// 	panic(err)
	// }
	// defer file.Close()
	// if err = png.Encode(file, filtered); err != nil {
	// 	panic(err)
	// }

	return filtered
}

// Test the value at a given point and return a single interpolated value
func sampleImageAt(imageData image.Image, coord Coordinate) float64 {

	imageBounds := imageData.Bounds()

	minCoord := coord.Floor()
	min := clamp(image.Point{int(minCoord.X), int(minCoord.Y)}, imageBounds)
	maxCoord := coord.Ceil()
	max := clamp(image.Point{int(maxCoord.X), int(maxCoord.Y)}, imageBounds)

	//fmt.Println("Sample at", coord, "Pixels", min, max)

	weight1 := (1.0 - (coord.X - minCoord.X)) * (1.0 - (coord.Y - minCoord.Y))
	weight2 := (1.0 - (coord.X - minCoord.X)) * (coord.Y - minCoord.Y)
	weight3 := (coord.X - minCoord.X) * (coord.Y - minCoord.Y)
	weight4 := (coord.X - minCoord.X) * (1.0 - (coord.Y - minCoord.Y))

	//fmt.Println("Weights", weight1, weight2, weight3, weight4)

	total := 0.0
	if weight1 != 0 {
		total += average(imageData.At(min.X, min.Y)) * weight1
	}
	if weight2 != 0 {
		total += average(imageData.At(min.X, max.Y)) * weight2
	}
	if weight3 != 0 {
		total += average(imageData.At(max.X, max.Y)) * weight3
	}
	if weight4 != 0 {
		total += average(imageData.At(max.X, min.Y)) * weight4
	}

	return total
}

// Clamp a bound to inside the given bounds
func clamp(point image.Point, bounds image.Rectangle) image.Point {
	if point.X < bounds.Min.X {
		point.X = bounds.Min.X
	}
	if point.X >= bounds.Max.X {
		point.X -= bounds.Max.X - 1
	}
	if point.Y < bounds.Min.Y {
		point.Y = bounds.Min.Y
	}
	if point.Y >= bounds.Max.Y {
		point.Y = bounds.Max.Y - 1
	}

	return point
}

// Returns an average of R,G,B from 0 to 1
func average(pixelColor color.Color) float64 {
	r, g, b, _ := pixelColor.RGBA()

	//fmt.Println("Sampling", r, g, b)

	return (float64(r)/65535.0 + float64(g)/65535.0 + float64(b)/65535.0) / 3.0
}
