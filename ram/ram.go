package ram

// RAMSIZE defines only the amount of RAM
const (
	RAMSIZE = 1 * 1024 * 1024
)

// RAM is the simulated ram
type RAM [RAMSIZE]uint16

// Read takes address and returns a value
func (r *RAM) Read(address uint32) uint16 {
	return (r[address])
}

// Write takes an address and a value
func (r *RAM) Write(address uint32, value uint16) {
	r[address] = value
}

// Clear RAM to 0
func (r *RAM) Clear() {
	for i := 0; i < RAMSIZE; i++ {
		r[i] = 0
	}
}
