package clock

import (
	"albert_go_sim/intmaxmin"
	"fmt"
)

// Clock is a struct which keeps track of the (simulated)
// elapsed time.
// Requires Frequency to be set in Hz e.g 50Hz == 50
// Set DoPrint to true if you want timestamps
type Clock struct {
	// numTicks grows forevever
	numTicks         uint64
	Frequency        int
	DoPrint          bool
	numTicksInSecond int
	// numSeconds grows forever
	numSeconds uint32
}

// Tick should be called on every tick of this virtual clock.
func (c *Clock) Tick() {
	c.numTicks++
	c.numTicksInSecond = intmaxmin.IncMod(c.numTicksInSecond, 1, c.Frequency)
	if c.numTicksInSecond == 0 {
		c.numSeconds++
		if c.DoPrint {
			fmt.Printf("(simulated) elapsed time (secs)%5d\n", c.numSeconds)

		}
	}

}

// Reset the internal time keeping of the clock
func (c *Clock) Reset() {
	c.numTicks = 0
	c.numSeconds = 0
}
