package counter

import (
	"albert_go_sim/intmaxmin"
	"fmt"
	"runtime"
)

const (
	ticksPerCount = 8
)

// Counter represents a memory mapped free running counter
type Counter struct {
	value   uint16
	tickNum int
}

// Init the counter
func (c *Counter) Init() {
	fmt.Printf("Initializing Counter\n")
}

// Tick should be called on every tick off the virtual clock
// The hardware counter will increment on every 8 Ticks
// to match the hardware
func (c *Counter) Tick() {
	if c.tickNum == 0 {
		c.value++
	}
	c.tickNum = intmaxmin.IncMod(c.tickNum, 1, ticksPerCount)
}

// CounterIsZero is meant to be a callback for external
// functions to check if the counter is zero
func (c *Counter) CounterIsZero() bool {
	return c.value == 0
}

// Read ignores address and returns the value of the counter
func (c *Counter) Read(address uint16) uint16 {
	return c.value
}

// Write for the counter doesn't make sense; flag as simulation WARNING
func (c *Counter) Write(address uint16, value uint16) {
	fmt.Printf("WARNING tried to write to read only counter\n")
	runtime.Goexit()
}
