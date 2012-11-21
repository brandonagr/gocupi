package plotter

import (
	"testing"
)

// Test that buffer works as expected
func TestCoordinateRingBuffer(t *testing.T) {

	buffer := NewCoordinateRingBuffer(4)

	coords := [4]Coordinate{ Coordinate{1,0}, Coordinate{2,0}, Coordinate{3,0}, Coordinate{4,0} }

	for index := 0; index < buffer.Cap(); index++ {
		buffer.Enqueue(coords[index])
	}
	if buffer.Len() != buffer.Cap() {
		t.Error("Expected buffer length to equal capacity", buffer.Len(), buffer.Cap())
	}
	for index := 0; index < buffer.Cap(); index++ {
		value := buffer.Dequeue()
		if value != coords[index] {
			t.Error("Expected", coords[index], "and got", value)
		}
	}
	if buffer.Len() != 0 {
		t.Error("Expected buffer length to be 0 and was", buffer.Len())
	}

	buffer.Enqueue(coords[2])
	if buffer.Dequeue() != coords[2] {
		t.Error("Unexpected result")
	}
}