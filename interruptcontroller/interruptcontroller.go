package interruptcontroller

import "fmt"

const (
	statusAddress = 0
	maskAddress   = 1
	clearAddress  = 2
)

type interruptCallbackFunction = func() bool

// InterruptController provides a simulated int ctlr
// It requires for each interrupt it should check and
// It provides a callback to note if its simulated
// interrupt output line is asserted (high)
type InterruptController struct {
	status          uint16
	mask            uint16
	clear           uint16
	Callbacks       [16]interruptCallbackFunction
	outputIsBlocked bool
}

// limitAddress captures just the lower two bits
// of an address because the interrupt controller
// only has 3 addresses
func limitAddress(a uint32) uint32 {
	return a & 0x0003
}

// Init the interrupt controller
func (i *InterruptController) Init() {
	fmt.Printf("Initializing interrupt controller\n")
}

// Tick should be called on every tick off the virtual clock
// The hardware interrupt controller polls all of the interrupt
// sources on every clock tick.
func (i *InterruptController) Tick() {

	for interruptNum := 0; interruptNum < 15; interruptNum++ {
		// Was a callback defined?
		if i.Callbacks[interruptNum] == nil {
			continue
		}
		// Does callback indicate an interrupt occurred?
		if !(i.Callbacks[interruptNum]()) {
			continue
		}
		// Does mask allow capturing this interrupt
		mask := (1 << interruptNum) & i.mask
		if mask == 0 {
			continue
		}
		// If we got this far, we update the status to indicate an int occurred
		i.status |= mask
	}
}

// GetOutput gets the single bit (implemented as bool)
// output of the interrupt controller.
// It is meant to be used as a callback by any code
// which needs to know if an interrupt has occurred.
func (i *InterruptController) GetOutput() bool {
	return i.status != 0
}

// Read takes address and returns a value
// Addresses are defined above
func (i *InterruptController) Read(address uint32) uint16 {

	address &= 0x0003
	if address == statusAddress {
		return i.status
	}
	if address == maskAddress {
		return i.mask
	}
	if address == clearAddress {
		return i.clear
	}

	return 0
}

// Write takes an address and a value
func (i *InterruptController) Write(address uint32, value uint16) {
	address = limitAddress(address)
	if address == clearAddress {
		i.clear = value
		i.status = i.status &^ i.clear
		return
	}
	if address == maskAddress {
		i.mask = value
	}
}
