package main

import (
	"bufio"
	"fmt"
	"gocupi"
	"log"
	"os"
	"time"
)

const (
	REG_MODE1      = 0x00
	REG_PRESCALE   = 0xFE
	REG_LED0_ON_L  = 0x06
	REG_LED0_ON_H  = 0x07
	REG_LED0_OFF_L = 0x08
	REG_LED0_OFF_H = 0x09

	REG_LED1_ON_L  = 0x0A
	REG_LED1_ON_H  = 0x0B
	REG_LED1_OFF_L = 0x0C
	REG_LED1_OFF_H = 0x0D
)

func main() {

	fmt.Println("hit enter")

	in := bufio.NewReader(os.Stdin)
	_, err := in.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	bus := gocupi.CreateBus(0x40)
	defer bus.Close()

	bus.Write8(REG_MODE1, 0x00) // reset

	oldMode := bus.Read8(REG_MODE1)
	newMode := (oldMode & 0x7F) | 0x10

	bus.Write8(REG_MODE1, newMode)
	bus.Write8(REG_PRESCALE, 101)
	bus.Write8(REG_MODE1, oldMode)
	time.Sleep(1 * time.Millisecond)
	bus.Write8(REG_MODE1, oldMode|0x80)

	for {
		bus.Write8(REG_LED1_ON_L, 0)
		bus.Write8(REG_LED1_ON_H, 0)
		bus.Write8(REG_LED1_OFF_L, 0xC8&0xFF)
		bus.Write8(REG_LED1_OFF_H, 0xC8>>8)

		time.Sleep(5 * time.Second)

		bus.Write8(REG_LED1_ON_L, 0)
		bus.Write8(REG_LED1_ON_H, 0)
		bus.Write8(REG_LED1_OFF_L, 0x1F4&0xFF)
		bus.Write8(REG_LED1_OFF_H, 0x1F4>>8)

		time.Sleep(5 * time.Second)
	}
}
