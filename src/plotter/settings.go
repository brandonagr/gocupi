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

	// Acceleration in mm / s^2
	Acceleration_MM_S2 float64 `xml:"-"`

	// Distance between the two motor spools
	HorizontalDistance_MM float64

	// Minimum distance below motors that can be drawn
	MinVertical_MM float64

	// Maximum distance below motors that can be drawn
	MaxVertical_MM float64

	// Max speed of the plot head
	MaxSpeed_MM_S float64 `xml:"-"`

	// Initial distance from head to left motor
	StartingLeftDist_MM float64

	// Initial distance from head to right motor
	StartingRightDist_MM float64

	// MM traveled by a single step
	StepSize_MM float64 `xml:"-"`
}

// Global settings variable
var Settings SettingsData

// Read settings from file, setting the global variable
func ReadSettings(settingsFile string) {

	fileData, err := ioutil.ReadFile(settingsFile)
	if err != nil {
		panic(err)
	}
	if err := xml.Unmarshal(fileData, &Settings); err != nil {
		panic(err)
	}

	// setup default values
	if Settings.TimeSlice_US == 0 {
		Settings.TimeSlice_US = 2048
	}
	if Settings.SpoolCircumference_MM == 0 {
		Settings.SpoolCircumference_MM = 60
	}
	if Settings.Acceleration_Seconds == 0 {
		Settings.Acceleration_Seconds = 1
	}

	// setup derived fields
	Settings.StepSize_MM = (Settings.SpoolSingleStep_Degrees / 360.0) * Settings.SpoolCircumference_MM

	// use 4 because packing data into a byte is done by multiplying it by 32, so 128 is the max value
	stepsPerRevolution := 360.0 / Settings.SpoolSingleStep_Degrees
	Settings.MaxSpeed_MM_S = ((4 / (Settings.TimeSlice_US / 1000000)) / stepsPerRevolution) * Settings.SpoolCircumference_MM
	Settings.Acceleration_MM_S2 = Settings.MaxSpeed_MM_S / Settings.Acceleration_Seconds
}

// Write settings to file
func WriteSettings(settingsFile string) {
	fileData, err := xml.MarshalIndent(Settings, "", "\t")
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(settingsFile, fileData, 0777); err != nil {
		panic(err)
	}
}
