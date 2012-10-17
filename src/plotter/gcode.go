package plotter
// Reads a text file and generates a program representation of the Gcode

// Any of the possible command types that are supported
type GcodeCommand int32
const (
	MOVE_RAPID GcodeCommand = 0
	MOVE = 1

	SET_UNITS_INCHES = 20
	SET_UNITS_MM = 21
)

// Data on a single line, the command followed by any values on the line
type GcodeLine struct {
	Command GcodeCommand
	//data map[string]float64
	Dest Coordinate
}

// All of the data from a file
type GcodeData struct {
	Lines []GcodeLine
}

// read all of the fileData lines, generating a GcodeData object
func ParseGcode(fileData []string) GcodeData {

	return GcodeData{}
}



