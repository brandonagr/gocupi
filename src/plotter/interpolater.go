package plotter

// Manages the trapezoidal interpolation

import (
	"fmt"
	"math"
)

// Given an origin, dest, nextDest, returns how many slices it will takes to traverse it and what the position at a given slice is
type PositionInterpolater interface {
	Setup(origin, dest, nextDest Coordinate)
	Slices() float64
	Position(slice float64) Coordinate
}

type LinearInterpolater struct {
	origin, destination Coordinate // positions currently interpolating between
	movement            Coordinate

	distance float64
	time     float64
	slices   float64
}

// setup
func (data *LinearInterpolater) Setup(origin, dest, nextDest Coordinate) {

	data.origin = origin
	data.destination = dest
	data.movement = data.destination.Minus(data.origin)
	data.distance = data.movement.Len()

	data.time = data.distance / Settings.MaxSpeed_MM_S
	data.slices = math.Ceil(data.time / (Settings.TimeSlice_US / 1000000))
}

// number of slices needed
func (data *LinearInterpolater) Slices() float64 {
	return data.slices
}

// position along line at this slice
func (data *LinearInterpolater) Position(slice float64) Coordinate {

	percentage := slice / data.slices
	return data.origin.Add(data.movement.Scaled(percentage))
}

// Data needed by the interpolater
type TrapezoidInterpolater struct {
	origin, destination Coordinate // positions currently interpolating between
	direction           Coordinate // unit direction vector from origin to destination

	entrySpeed  float64
	cruiseSpeed float64
	exitSpeed   float64

	distance float64
	time     float64
	slices   float64 // number of Settings.TIME_SLICE_US slices

	accelTime  float64
	accelDist  float64
	cruiseTime float64
	cruiseDist float64
	decelTime  float64
	decelDist  float64
}

func (data *TrapezoidInterpolater) WriteData() {
	fmt.Println("Origin:", data.origin, "Dest:", data.destination)
	fmt.Println("Dir:", data.direction, "Slices:", data.slices)
	fmt.Println()

	fmt.Println("Entry", data.entrySpeed, "Cruise", data.cruiseSpeed, "Exit", data.exitSpeed)

	fmt.Println("Taccel", data.accelTime, "Tcruise", data.cruiseTime, "Tdecel", data.decelTime)
	fmt.Println("Daccel", data.accelDist, "Dcruise", data.cruiseDist, "Ddecel", data.decelDist)

	fmt.Println("Total distance", data.distance)
}

// Calculate all fields needed
func (data *TrapezoidInterpolater) Setup(origin, dest, nextDest Coordinate) {

	// entry speed is whatever the previous exit speed was
	data.entrySpeed = data.exitSpeed

	data.origin = origin
	data.destination = dest
	data.direction = dest.Minus(origin)
	data.distance = data.direction.Len()
	data.direction = data.direction.Normalized()

	nextDirection := nextDest.Minus(dest)
	if nextDirection.Len() == 0 {
		// if there is no next direction, make the exit speed 0 by pretending the next move will be backwards from current direction
		nextDirection = Coordinate{-data.direction.X, -data.direction.Y}
	} else {
		nextDirection = nextDirection.Normalized()
	}
	cosAngle := data.direction.DotProduct(nextDirection)
	data.exitSpeed = Settings.MaxSpeed_MM_S * math.Max(cosAngle, 0.0)

	data.cruiseSpeed = Settings.MaxSpeed_MM_S

	data.accelTime = (data.cruiseSpeed - data.entrySpeed) / Settings.Acceleration_MM_S2
	data.accelDist = 0.5*Settings.Acceleration_MM_S2*data.accelTime*data.accelTime + data.entrySpeed*data.accelTime

	data.decelTime = (data.cruiseSpeed - data.exitSpeed) / Settings.Acceleration_MM_S2
	data.decelDist = 0.5*-Settings.Acceleration_MM_S2*data.decelTime*data.decelTime + data.cruiseSpeed*data.decelTime

	data.cruiseDist = data.distance - (data.accelDist + data.decelDist)
	data.cruiseTime = data.cruiseDist / data.cruiseSpeed

	// we dont have enough room to reach max velocity, just cruise at exit speed
	if data.distance < data.accelDist+data.decelDist {

		// equation derived by http://www.numberempire.com/equationsolver.php from equations:
		// distanceAccel = 0.5 * accel * timeAccel^2 + entrySpeed * timeAccel
		// distanceDecel = 0.5 * -accel * timeDecel^2 + maxSpeed * timeDecel
		// totalDistance = distanceAccel + distanceDecel
		// maxSpeed = entrySpeed + accel * timeAccel
		// maxSpeed = exitSpeed + accel * timeDecel
		data.decelTime = (math.Sqrt(2)*math.Sqrt(data.exitSpeed*data.exitSpeed+data.entrySpeed*data.entrySpeed+2*Settings.Acceleration_MM_S2*data.distance) - 2*data.exitSpeed) / (2 * Settings.Acceleration_MM_S2)
		data.cruiseTime = 0
		data.cruiseDist = 0
		data.cruiseSpeed = data.exitSpeed + Settings.Acceleration_MM_S2*data.decelTime
		data.accelTime = (data.cruiseSpeed - data.entrySpeed) / Settings.Acceleration_MM_S2

		data.accelDist = 0.5*Settings.Acceleration_MM_S2*data.accelTime*data.accelTime + data.entrySpeed*data.accelTime
		data.decelDist = 0.5*-Settings.Acceleration_MM_S2*data.decelTime*data.decelTime + data.cruiseSpeed*data.decelTime

		// don't have enough room to accelerate to exitSpeed over the given distance, have to change exit speed
		if data.decelTime < 0 || data.accelTime < 0 {

			if data.exitSpeed > data.entrySpeed { // need to accelerate to max exit speed possible

				data.decelDist = 0
				data.decelTime = 0
				data.cruiseDist = 0
				data.cruiseTime = 0

				//given distance, how much can you accelerate from entrySpeed
				data.accelTime = (math.Sqrt(data.entrySpeed*data.entrySpeed+2*Settings.Acceleration_MM_S2*data.distance) - data.entrySpeed) / Settings.Acceleration_MM_S2
				data.exitSpeed = data.entrySpeed + Settings.Acceleration_MM_S2*data.accelTime
				data.cruiseSpeed = data.exitSpeed
				data.accelDist = 0.5*Settings.Acceleration_MM_S2*data.accelTime*data.accelTime + data.entrySpeed*data.accelTime
			} else { // need to decelerate to min exit speed possible

				data.accelDist = 0
				data.accelTime = 0
				data.cruiseDist = 0
				data.cruiseTime = 0

				data.decelTime = (math.Sqrt(data.entrySpeed*data.entrySpeed-2*Settings.Acceleration_MM_S2*data.distance) + data.entrySpeed) / Settings.Acceleration_MM_S2
				data.cruiseSpeed = data.entrySpeed
				data.exitSpeed = data.entrySpeed - Settings.Acceleration_MM_S2*data.decelTime
				data.decelDist = 0.5*-Settings.Acceleration_MM_S2*data.decelTime*data.decelTime + data.cruiseSpeed*data.decelTime
			}
		}
	}

	data.time = data.accelTime + data.cruiseTime + data.decelTime
	data.slices = math.Ceil(data.time / (Settings.TimeSlice_US / 1000000))
}

// Calculate current position at the given time
func (data *TrapezoidInterpolater) Position(slice float64) Coordinate {

	// on last interp just return the destination
	if slice == data.slices {
		return data.destination
	}

	time := (slice / data.slices) * data.time
	var distanceAlongMovement float64 = 0

	if time < data.accelTime { // in acceleration

		distanceAlongMovement = 0.5*Settings.Acceleration_MM_S2*time*time + data.entrySpeed*time
	} else if time < data.accelTime+data.cruiseTime { // in cruise

		time = time - data.accelTime
		distanceAlongMovement = data.accelDist + time*data.cruiseSpeed

	} else { // in deceleration

		time = time - (data.accelTime + data.cruiseTime)
		distanceAlongMovement = data.accelDist + data.cruiseDist + 0.5*-Settings.Acceleration_MM_S2*time*time + data.cruiseSpeed*time
	}

	return data.origin.Add(data.direction.Scaled(distanceAlongMovement))
}

// Get total time it takes to move
func (data *TrapezoidInterpolater) Slices() float64 {
	return data.slices
}
