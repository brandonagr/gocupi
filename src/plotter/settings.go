package plotter

import (
	"encoding/xml"
	"io/ioutil"
)

type SettingsData struct {
	// Minimum time step used to control motion
	TimeSlice_US float64

	// Circumference of the motor spool
	SpoolCircumference_MM float64

	// Degrees in a single step, set based on the stepper motor & microstepping
	SpoolSingleStep_Degrees float64

	// Number of seconds to accelerate from 0 to MaxSpeed_MM_S
	Acceleration_Seconds float64

	// Distance between the two motor spools
	HorizontalDistance_MM float64

	// Minimum distance below motors that can be drawn
	MinVertical_MM float64

	// Maximum distance below motors that can be drawn
	MaxVertical_MM float64

	// Initial distance from head to left motor
	StartingLeftDist_MM float64

	// Initial distance from head to right motor
	StartingRightDist_MM float64

	// MM traveled by a single step
	StepSize_MM float64 `xml:"-"`

	// Max speed of the plot head
	MaxSpeed_MM_S float64 `xml:"-"`

	// Acceleration in mm / s^2, derived from Acceleration_Seconds and MaxSpeed_MM_S
	Acceleration_MM_S2 float64 `xml:"-"`
}

// Global settings variable
var Settings SettingsData

// location of the settings file
var settingsFile string = "../settings.xml"

// Read settings from file, setting the global variable
func (settings *SettingsData) Read() {

	fileData, err := ioutil.ReadFile(settingsFile)
	if err != nil {
		panic(err)
	}
	if err := xml.Unmarshal(fileData, settings); err != nil {
		panic(err)
	}

	// setup default values
	if settings.TimeSlice_US == 0 {
		settings.TimeSlice_US = 2048
	}
	if settings.SpoolCircumference_MM == 0 {
		settings.SpoolCircumference_MM = 60
	}
	if settings.Acceleration_Seconds == 0 {
		settings.Acceleration_Seconds = 1
	}

	// setup derived fields
	settings.StepSize_MM = (settings.SpoolSingleStep_Degrees / 360.0) * settings.SpoolCircumference_MM

	// use 4 because packing data into a byte is done by multiplying it by 32, so 128 is the max value
	stepsPerRevolution := 360.0 / settings.SpoolSingleStep_Degrees
	settings.MaxSpeed_MM_S = ((4 / (settings.TimeSlice_US / 1000000)) / stepsPerRevolution) * settings.SpoolCircumference_MM
	settings.MaxSpeed_MM_S *= 0.98 // give max speed some extra room to not hit 127 limit
	settings.Acceleration_MM_S2 = settings.MaxSpeed_MM_S / settings.Acceleration_Seconds
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
