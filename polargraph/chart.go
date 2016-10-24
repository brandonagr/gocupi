package polargraph

import (
	chart "github.com/p/plotinum/plot"
    chartplotter "github.com/gonum/plot/plotter"
	vg "github.com/gonum/plot/vg"
	"image/color"
)

// Writes step data and position to a graph
func WriteStepsToChart(stepData <-chan int8) {

	maxNumberSteps := 1500

	leftVel := make(chartplotter.XYs, maxNumberSteps)
	rightVel := make(chartplotter.XYs, maxNumberSteps)
	leftPos := make(chartplotter.XYs, maxNumberSteps)
	rightPos := make(chartplotter.XYs, maxNumberSteps)

	var byteDataL, byteDataR int8
	stepIndex := 0
	for stepDataOpen := true; stepDataOpen; {

		byteDataL, stepDataOpen = <-stepData
		byteDataR, stepDataOpen = <-stepData

		leftVel[stepIndex].X = float64(stepIndex)
		leftVel[stepIndex].Y = float64(byteDataL) * Settings.StepSize_MM / (32.0 * 0.002)

		rightVel[stepIndex].X = float64(stepIndex)
		rightVel[stepIndex].Y = float64(byteDataR) * Settings.StepSize_MM / (32.0 * 0.002)

		leftPos[stepIndex].X = float64(stepIndex)
		if stepIndex > 0 {
			leftPos[stepIndex].Y = leftPos[stepIndex-1].Y + (leftVel[stepIndex].Y * .002)
		} else {
			leftPos[stepIndex].Y = 0
		}

		rightPos[stepIndex].X = float64(stepIndex)
		if stepIndex > 0 {
			rightPos[stepIndex].Y = rightPos[stepIndex-1].Y + (rightVel[stepIndex].Y * .002)
		} else {
			rightPos[stepIndex].Y = 0
		}

		stepIndex++
		if stepIndex >= maxNumberSteps {
			break
		}
	}

	// chop down slices if maxNumberSteps wasn't needed
	if stepIndex < maxNumberSteps {
		leftVel = leftVel[:stepIndex]
		rightVel = rightVel[:stepIndex]
		leftPos = leftPos[:stepIndex]
		rightPos = rightPos[:stepIndex]
	}

	// Create a new plot, set its title and
	// axis labels.
	p, err := chart.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "Polargraph Position & Velocity"
	p.X.Label.Text = "2ms Slice"
	p.Y.Label.Text = "Position(mm) / Velocity (mm/s)"

	// Draw a grid behind the data
	p.Add(chartplotter.NewGrid())

	leftPosLine, _ := chartplotter.NewLine(leftPos)
	leftPosLine.LineStyle.Width = vg.Points(1)
	leftPosLine.LineStyle.Color = color.RGBA{B: 255, A: 255}

	rightPosLine, _ := chartplotter.NewLine(rightPos)
	rightPosLine.LineStyle.Width = vg.Points(1)
	rightPosLine.LineStyle.Color = color.RGBA{R: 255, A: 255}

	leftVelLine, _ := chartplotter.NewLine(leftVel)
	leftVelLine.LineStyle.Width = vg.Points(1)
	leftVelLine.LineStyle.Color = color.RGBA{B: 150, G: 150, A: 255}

	rightVelLine, _ := chartplotter.NewLine(rightVel)
	rightVelLine.LineStyle.Width = vg.Points(1)
	rightVelLine.LineStyle.Color = color.RGBA{R: 150, G: 150, A: 255}

	p.Add(leftPosLine, rightPosLine)
	p.Legend.Add("Left Pos", leftPosLine)
	p.Legend.Add("Right Pos", rightPosLine)

	p.Add(leftVelLine, rightVelLine)
	p.Legend.Add("Left Vel", leftVelLine)
	p.Legend.Add("Right Vel", rightVelLine)

	// Save the plot to a PNG file.
	if err := p.Save(14, 8.5, "chart.png"); err != nil {
		panic(err)
	}
}
