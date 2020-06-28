package rom

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

type Rom [0x400]uint16

// loadPatsLoader takes the original hex "binary" and places it in ram
func (r *Rom) loadPatsLoader() {
	fmt.Printf("Entered loadPatsLoader\n")
	f, err := os.Open("loader_from_zero.txt")
	if err != nil {
		fmt.Printf("Could not open loader file\n")
		fmt.Printf("Error is %v\n", err)
		os.Exit(1)
		return
	}
	scanner := bufio.NewScanner(f)
	address := uint32(0)
	// Notice we write directly to r; we do NOT
	// use the Write method because that is for the
	// public and writing is NOT permitted to ROM.
	for scanner.Scan() {
		s := scanner.Text()
		n, _ := strconv.ParseUint(s, 16, 32)
		w := uint16(n)
		// r.Write(address, w)
		r[address] = w
		address++
	}
	f.Close()

}

// Init sets up memory including ROM settings
// and protection
func (r *Rom) Init() {
	r.loadPatsLoader()
}

// Read takes address and returns a value
func (r *Rom) Read(address uint32) uint16 {
	return (r[address])
}

// Write takes an address and a value
func (r *Rom) Write(address uint32, value uint16) {
	fmt.Printf("FATAL - tried to write to ROM address %08X\n", address)
	os.Exit(1)
}
