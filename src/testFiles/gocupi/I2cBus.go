package gocupi

import (
	"log"
	"os"
	"syscall"
	"unsafe"
)

// Type representing an bus connection
type I2cBus struct {
	fd *os.File
}

// Data that is passed to/from ioctl calls
type i2c_smbus_ioctl_data struct {
	read_write uint8
	command    uint8
	size       int
	data       *uint8
}

// Constants used by ioctl
const (
	I2C_SMBUS_READ  = 1
	I2C_SMBUS_WRITE = 0

	I2C_SMBUS_BYTE_DATA = 2

	I2C_SMBUS = 0x0720
)

// Create an I2cBus for the specified address
func CreateBus(address uint8) *I2cBus {
	file, err := os.OpenFile("/dev/i2c-0", os.O_RDWR, os.ModeExclusive)
	if err != nil {
		log.Fatal(err)
	}

	busAddr := I2C_SMBUS
	argp := uint(address)
	_, _, err = syscall.Syscall(syscall.SYS_IOCTL, uintptr(file.Fd()), uintptr(unsafe.Pointer(&busAddr)), uintptr(unsafe.Pointer(&argp)))
	if err != nil {
		log.Fatal(err)
	}

	return &I2cBus{
		fd: file,
	}
}

// Write 1 byte onto the bus
func (bus *I2cBus) Write8(command, data uint8) {

	// i2c_smbus_access(file,I2C_SMBUS_WRITE,command,I2C_SMBUS_BYTE_DATA, &data);

	busData := i2c_smbus_ioctl_data{
		read_write: I2C_SMBUS_WRITE,
		command:    command,
		size:       I2C_SMBUS_BYTE_DATA,
		data:       &data,
	}

	busAddr := I2C_SMBUS
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(bus.fd.Fd()), uintptr(unsafe.Pointer(&busAddr)), uintptr(unsafe.Pointer(&busData)))
	if err != 0 {
		log.Fatal(err)
	}
}

// Read 1 byte from the bus
func (bus *I2cBus) Read8(command uint8) uint8 {

	// i2c_smbus_access(file,I2C_SMBUS_READ,command,I2C_SMBUS_BYTE_DATA,&data)

	data := uint8(0)

	busData := i2c_smbus_ioctl_data{
		read_write: I2C_SMBUS_READ,
		command:    command,
		size:       I2C_SMBUS_BYTE_DATA,
		data:       &data,
	}

	busAddr := I2C_SMBUS
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(bus.fd.Fd()), uintptr(unsafe.Pointer(&busAddr)), uintptr(unsafe.Pointer(&busData)))
	if err != 0 {
		log.Fatal(err)
	}

	return data
}
