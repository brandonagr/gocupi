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
	WriteData()
}

type LinearInterpolater struct {
	origin, destination Coordinate // positions currently interpolating between
	movement            Coordinate

	distance float64
	time     float64
	slices   float64
}

// setup data for the linear interpolater
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

// output data
func (data *LinearInterpolater) WriteData() {
	fmt.Println("Dist", data.distance, "Time", data.time, "Slices", data.slices)
}

// Data needed by the interpolater
type TrapezoidInterpolater struct {
	origin      Coordinate // positions currently interpolating from
	destination Coordinate // position currently interpolating towards
	direction   Coordinate // unit direction vector from origin to destination

	entrySpeed  float64 // speed at beginning at origin
	cruiseSpeed float64 // maximum speed reached
	exitSpeed   float64 // target speed when we reach destination

	acceleration float64 // acceleration, only differs from Settings.Acceleration_MM_S2 when decelerating and there is not enough distance to hit exit speed

	distance float64 // total distance travelled
	time     float64 // total time to go from origin to destination
	slices   float64 // number of Settings.TIME_SLICE_US slices

	accelTime  float64 // time accelerating
	accelDist  float64 // distance covered while accelerating
	cruiseTime float64 // time cruising at max speed
	cruiseDist float64 // distance covered while cruising
	decelTime  float64 // time decelerating
	decelDist  float64 // distance covered while decelerating
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

	// special case of not going anywhere
	if origin == dest {
		data.origin = origin
		data.destination = data.destination
		data.direction = Coordinate{X: 0, Y: 1}
		data.distance = 0
		data.exitSpeed = data.entrySpeed
		data.cruiseSpeed = data.entrySpeed
		data.accelDist = 0
		data.accelTime = 0
		data.cruiseDist = 0
		data.cruiseTime = 0
		data.decelDist = 0
		data.decelTime = 0
		data.acceleration = Settings.Acceleration_MM_S2
		data.time = 0
		data.slices = 0
		return
	}

	data.origin = origin
	data.destination = dest
	data.direction = data.destination.Minus(origin)
	data.distance = data.direction.Len()
	data.direction = data.direction.Normalized()

	nextDirection := nextDest.Minus(dest)
	if nextDirection.Len() == 0 {
		// if there is no next direction, make the exit speed 0 by pretending the next move will be backwards from current direction
		nextDirection = Coordinate{X: -data.direction.X, Y: -data.direction.Y}
	} else {
		nextDirection = nextDirection.Normalized()
	}
	cosAngle := data.direction.DotProduct(nextDirection)
	cosAngle = math.Pow(cosAngle, 3) // use cube in order to make it smaller for non straight lines
	data.exitSpeed = Settings.MaxSpeed_MM_S * math.Max(cosAngle, 0.0)

	data.cruiseSpeed = Settings.MaxSpeed_MM_S

	data.accelTime = (data.cruiseSpeed - data.entrySpeed) / Settings.Acceleration_MM_S2
	data.accelDist = 0.5*Settings.Acceleration_MM_S2*data.accelTime*data.accelTime + data.entrySpeed*data.accelTime

	data.decelTime = (data.cruiseSpeed - data.exitSpeed) / Settings.Acceleration_MM_S2
	data.decelDist = 0.5*-Settings.Acceleration_MM_S2*data.decelTime*data.decelTime + data.cruiseSpeed*data.decelTime

	data.cruiseDist = data.distance - (data.accelDist + data.decelDist)
	data.cruiseTime = data.cruiseDist / data.cruiseSpeed

	data.acceleration = Settings.Acceleration_MM_S2

	// we dont have enough room to reach max velocity, have to calculate what max speed we can reach
	if data.distance < data.accelDist+data.decelDist {

		// equation derived by http://www.numberempire.com/equationsolver.php from equations:
		// distanceAccel = 0.5 * accel * timeAccel^2 + entrySpeed * timeAccel
		// distanceDecel = 0.5 * -accel * timeDecel^2 + maxSpeed * timeDecel
		// totalDistance = distanceAccel + distanceDecel
		// maxSpeed = entrySpeed + accel * timeAccel
		// maxSpeed = exitSpeed + accel * timeDecel
		data.decelTime = (math.Sqrt2*math.Sqrt(data.exitSpeed*data.exitSpeed+data.entrySpeed*data.entrySpeed+2*Settings.Acceleration_MM_S2*data.distance) - 2*data.exitSpeed) / (2 * Settings.Acceleration_MM_S2)
		data.cruiseTime = 0
		data.cruiseSpeed = data.exitSpeed + Settings.Acceleration_MM_S2*data.decelTime
		data.accelTime = (data.cruiseSpeed - data.entrySpeed) / Settings.Acceleration_MM_S2

		// don't have enough room to accelerate to exitSpeed over the given distance, have to change exit speed
		if data.decelTime < 0 || data.accelTime < 0 {

			if data.exitSpeed > data.entrySpeed { // need to accelerate to max exit speed possible

				data.decelDist = 0
				data.decelTime = 0
				data.cruiseDist = 0
				data.cruiseTime = 0

				// determine time it will take to travel distance at the given acceleration
				data.accelTime = (math.Sqrt(data.entrySpeed*data.entrySpeed+2*Settings.Acceleration_MM_S2*data.distance) - data.entrySpeed) / Settings.Acceleration_MM_S2
				data.exitSpeed = data.entrySpeed + Settings.Acceleration_MM_S2*data.accelTime
				data.cruiseSpeed = data.exitSpeed
				data.accelDist = data.distance
			} else { // need to decelerate to exit speed, by changing acceleration

				//fmt.Println("Warning, unable to decelerate to target exit speed using acceleration, try adding -slowfactor=2")

				data.accelDist = 0
				data.accelTime = 0
				data.cruiseDist = 0
				data.cruiseTime = 0

				// determine time it will take to reach exit speed over the given distance
				data.decelTime = 2.0 * data.distance / (data.exitSpeed + data.entrySpeed)
				data.acceleration = (data.entrySpeed - data.exitSpeed) / data.decelTime
				data.cruiseSpeed = data.entrySpeed
				data.decelDist = data.distance
			}
		} else {
			data.accelDist = 0.5*Settings.Acceleration_MM_S2*data.accelTime*data.accelTime + data.entrySpeed*data.accelTime
			data.cruiseDist = 0
			data.decelDist = 0.5*-Settings.Acceleration_MM_S2*data.decelTime*data.decelTime + data.cruiseSpeed*data.decelTime
		}
	}

	data.time = data.accelTime + data.cruiseTime + data.decelTime
	data.slices = data.time / (Settings.TimeSlice_US / 1000000)
}

// Calculate current position at the given time
func (data *TrapezoidInterpolater) Position(slice float64) Coordinate {

	//fmt.Println("Slice", slice, "out of", data.slices, "percent", slice/data.slices)

	time := (slice / data.slices) * data.time
	var distanceAlongMovement float64 = 0

	if time < data.accelTime { // in acceleration

		distanceAlongMovement = 0.5*Settings.Acceleration_MM_S2*time*time + data.entrySpeed*time

		//fmt.Println("Accel", time, distanceAlongMovement, "speed is", Settings.Acceleration_MM_S2*time+data.entrySpeed)
	} else if time < data.accelTime+data.cruiseTime { // in cruise

		time = time - data.accelTime
		distanceAlongMovement = data.accelDist + time*data.cruiseSpeed

		//fmt.Println("Cruise", time, distanceAlongMovement)
	} else { // in deceleration

		time = time - (data.accelTime + data.cruiseTime)
		distanceAlongMovement = data.accelDist + data.cruiseDist + 0.5*-data.acceleration*time*time + data.cruiseSpeed*time

		//fmt.Println("Decel", time, distanceAlongMovement, "speed is", -data.acceleration*time+data.cruiseSpeed)
	}

	return data.origin.Add(data.direction.Scaled(distanceAlongMovement))
}

// Get total time it takes to move
func (data *TrapezoidInterpolater) Slices() float64 {
	return data.slices
}

// Test code to use a larger buffer to do the look ahead:
// package main

// import "fmt"

// func main() {
// 	firstChannel := make(chan int, 5)
// 	secondChannel := make(chan int, 5)

// 	//go DoProcessing(firstChannel, secondChannel)
// 	go DoProcessingBuffer(5, firstChannel, secondChannel)

// 	go func() {
// 		for data := 0; data < 10; data++ {
// 			firstChannel <- data
// 		}
// 		close(firstChannel)
// 	}()

// 	// write resulting data to screen
// 	for resultData := range secondChannel {
// 		fmt.Println("=", resultData)
// 	}
// }

// // Do processing with no buffer, just looks one value ahead
// func DoProcessing(incomingData <-chan int, outgoingData chan<- int) {
// 	defer close(outgoingData)

// 	for data := range incomingData{
// 		outgoingData <- data + 1
// 	}
// }

// // Do processing with a buffer of size bufferSize
// func DoProcessingBuffer(bufferSize int, incomingData <-chan int, outgoingData chan<- int) {
// 	defer close(outgoingData)

//  // would be more efficient with a circular buffer
// 	dataBuffer := make([]int, 0, bufferSize)
// 	chanOpen := true
// 	var data int
// 	for fillBuffer := 0; fillBuffer < bufferSize && chanOpen; fillBuffer++ {
// 		data, chanOpen = <-incomingData
// 		if chanOpen {
// 			dataBuffer = append(dataBuffer, data)
// 		}
// 	}

// 	for len(dataBuffer) > 0{
// 		total := 0
// 		for _, bufferValue := range dataBuffer {
// 			total += bufferValue;
// 			fmt.Print(bufferValue, ", ")
// 		}
// 		outgoingData <- total
// 		fmt.Println()

// 		dataBuffer = dataBuffer[1:]
// 		data, chanOpen = <-incomingData
// 		if chanOpen {
// 			dataBuffer = append(dataBuffer, data)
// 		}
// 	}
// }

// A ring buffer used to store coordinates
type CoordinateRingBuffer struct {
	data     []Coordinate // data in the buffer
	capacity int          // length of data buiffer
	start    int          // current beginning of buffer
	length   int          // number of items in buffer
}

// Create a new buffer with the given capacity
func NewCoordinateRingBuffer(capacity int) *CoordinateRingBuffer {
	return &CoordinateRingBuffer{
		data:     make([]Coordinate, capacity),
		capacity: capacity,
		start:    0,
		length:   0,
	}
}

// Add a coordinate to the end of the buffer
func (ring *CoordinateRingBuffer) Enqueue(coord Coordinate) {
	writeIndex := ring.start + ring.length
	if writeIndex >= ring.Cap() {
		writeIndex -= ring.capacity
	}

	ring.data[writeIndex] = coord

	if ring.length == ring.capacity {
		panic("Attempted to overfill buffer")
	}

	ring.length++
}

// Remove a coordinate from the beginning of the buffer
func (ring *CoordinateRingBuffer) Dequeue() Coordinate {
	if ring.length == 0 {
		panic("Attempted to dequeue an empty buffer")
	}

	result := ring.data[ring.start]

	ring.start++
	if ring.start == ring.capacity {
		ring.start = 0
	}
	ring.length--

	return result
}

// Amount of data in the buffer
func (ring *CoordinateRingBuffer) Len() int {
	return ring.length
}

// Capacity of the buffer
func (ring *CoordinateRingBuffer) Cap() int {
	return ring.capacity
}
