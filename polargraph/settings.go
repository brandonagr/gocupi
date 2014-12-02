package polargraph

import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// These constants are also set in StepperDriver.ino, must be changed in both places
const (
	// Time step used to control motion, ie the amount of time that the stepper motors will be going a constant speed
	// decreasing this increases CPU usage and serial communication
	// increasing it decreases rendering quality
	// when running on a raspberry pi 2048 us (2 milliseconds) seems like a good number
	TimeSlice_US float64 = 2048

	// The factor the steps are multiplied by, needs to be the same as set in the arduino code
	StepsFixedPointFactor float64 = 32.0

	// Determined because 1 byte is sent per value, so have range -128 to 127, and -128, -127, 127 are reserved values with special meanings
	StepsMaxValue float64 = 126.0

	// Special Steps value that when received causes the arduino to flush its buffers and reset its internal state
	ResetCommand byte = 0x80 // -128

	// Special Steps value that raises the pen
	PenUpCommand int8 = -127

	// Special Steps value that lowers the pen
	PenDownCommand int8 = 127
)

// User configurable settings
type SettingsData struct {
	// Circumference of the motor spool
	SpoolCircumference_MM float64

	// Degrees in a single step, set based on the stepper motor & microstepping
	SpoolSingleStep_Degrees float64

	// Number of seconds to accelerate from 0 to MaxSpeed_MM_S
	Acceleration_Seconds float64

	// Distance between the two motor spools
	SpoolHorizontalDistance_MM float64

	// Minimum distance below motors that can be drawn
	DrawingSurfaceMinY_MM float64

	// Maximum distance below motors that can be drawn
	DrawingSurfaceMaxY_MM float64

	// Distance from the left edge that the pen can go
	DrawingSurfaceMinX_MM float64

	// Calculated from SpoolHorizontalDistance_MM and DrawingSufaceMinX_MM
	DrawingSurfaceMaxX_MM float64 `xml:"-"`

	// Initial distance from head to left motor
	StartingLeftDist_MM float64

	// Initial distance from head to right motor
	StartingRightDist_MM float64

	// path to mouse event file, use evtest to find
	MousePath string

	// MM traveled by a single step
	StepSize_MM float64 `xml:"-"`

	// Max speed of the plot head
	MaxSpeed_MM_S float64 `xml:"-"`

	// Acceleration in mm / s^2, derived from Acceleration_Seconds and MaxSpeed_MM_S
	Acceleration_MM_S2 float64 `xml:"-"`
}

// Global settings variable
var Settings SettingsData

// name of the settings / config file
var settingsFile string = "gocupi_config.xml"

// Read settings from file, setting the global variable
func (settings *SettingsData) Read() {

	fileData, err := ioutil.ReadFile(settingsFile)
	if err != nil {
		// attempt to copy the default settings file from the gopath directory to the current working directory
		repoSettingsFile := filepath.Join(os.Getenv("GOPATH"), "src/github.com/brandonagr/gocupi", settingsFile)
		if copyErr := copyFile(repoSettingsFile, settingsFile); copyErr != nil {
			panic(copyErr)
		}

		fileData, err = ioutil.ReadFile(settingsFile)
		if err != nil {
			panic(err)
		}
	}
	if err := xml.Unmarshal(fileData, settings); err != nil {
		panic(err)
	}

	// setup default values
	if settings.SpoolCircumference_MM == 0 {
		settings.SpoolCircumference_MM = 60
	}
	if settings.Acceleration_Seconds == 0 {
		settings.Acceleration_Seconds = 1
	}

	settings.CalculateDerivedFields()
}

// setup derived fields
func (settings *SettingsData) CalculateDerivedFields() {
	settings.DrawingSurfaceMaxX_MM = settings.SpoolHorizontalDistance_MM - settings.DrawingSurfaceMinX_MM
	settings.StepSize_MM = (settings.SpoolSingleStep_Degrees / 360.0) * settings.SpoolCircumference_MM

	stepsPerRevolution := 360.0 / settings.SpoolSingleStep_Degrees
	stepsPerValue := StepsMaxValue / StepsFixedPointFactor
	settings.MaxSpeed_MM_S = ((stepsPerValue / (TimeSlice_US / 1000000.0)) / stepsPerRevolution) * settings.SpoolCircumference_MM
	settings.Acceleration_MM_S2 = settings.MaxSpeed_MM_S / settings.Acceleration_Seconds
}

// from https://gist.github.com/elazarl/5507969
func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	// no need to check errors on read only file, we already got everything
	// we need from the filesystem, so nothing can go wrong now.
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}

// Write settings to file
func (settings *SettingsData) Write() {
	fileData, err := xml.MarshalIndent(settings, "", "\t")
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(settingsFile, fileData, 0777); err != nil {
		panic(err)
	}
}
