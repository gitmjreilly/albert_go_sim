package memory

import (
	"fmt"
	"os"
	"runtime"
)

// Constants associated with memory mapped devices.
// Think of these as chip select lines.
// For the devices mapped beteen F000 and F0DF
// we assume the constant can be derived from the address.
// e.g F023 has const 2 == (F023 - F000) / 0x10
// RAM and ROM are special cases because of the way the hardware
// is actually constructed.  See _helper for all of the special case handling
const (
	F000  = iota
	F010  = iota
	F020  = iota
	F030  = iota
	F040  = iota
	F050  = iota
	F060  = iota
	F070  = iota
	F080  = iota
	F090  = iota
	F0A0  = iota
	F0B0  = iota
	F0C0  = iota
	F0D0  = iota
	RAMCS = iota
	RomCS = iota
)

// MEMSIZE defines the size of the entire address space
const (
	MEMSIZE = 1 * 1024 * 1024
)

// Memory Protection Permissions
const (
	CODERO   = iota
	DATARO   = iota
	DATARW   = iota
	NOACCESS = iota
)

// TMemory matches hardware
// RAM with protection plus an array of mapped devices
type TMemory struct {
	mappedDevice [16]struct {
		readData  func(address uint32) uint16
		writeData func(address uint32, value uint16)
		isMapped  bool
	}
	memory [MEMSIZE]struct {
		data       uint16
		protection uint8
	}
}

// _helper takes the address of a memory mapped device
// and returns the index into the TMemory structure
// of the mappedDevice.
// The idea is, for any given mapped, memory-like thing
// we'll call the correct read and write functions.
//
// It also returns the address for use within the device
// e.g. if address is 0xF012
// return 1 for the index and 2 for the address
//
// RAM and ROM are special cases.  Their addresses are accepted as-is.
func _helper(address uint32) (int, uint32) {
	// The order here is super important

	// Check for the most common case, RAM
	// RAM lives above ROM and excludes the memory mapped devices in F000 -> F0DF
	if address >= 0x0400 && (address < 0xF000 || address > 0xF0DF) {
		return RAMCS, address
	}

	// Check for ROM
	if address < 0x0400 {
		return RomCS, address
	}

	// If we got this far, we must be in the memory mapped device range -
	// address(es) are 0xF000 <= x <= 0xF0DF
	address -= 0xF000
	// address is now 0x0000 <= x <= 0x00FF

	deviceIndex := int(address / 0x0010)
	deviceAddress := uint32(address % 0x0010)

	return deviceIndex, deviceAddress
}

// Read takes an address and returns a value
// from the memory map.  Could be ram/rom or
// memory mapped device like a Serial Port
func (m *TMemory) Read(address uint32) uint16 {

	index, subAddress := _helper(address)

	if !m.mappedDevice[index].isMapped {
		fmt.Printf("FATAL Error in memory read.  Attempt to read non mapped memory.\n")
		fmt.Printf("address is [%08X]\n", address)
		runtime.Goexit()
	}

	value := m.mappedDevice[index].readData(subAddress)

	return value
}

// ReadCodeMemory takes an address and returns a value
// from the memory map.  Could be ram/rom or
// memory mapped device like a Serial Port
func (m *TMemory) ReadCodeMemory(address uint32) uint16 {

	index, subAddress := _helper(address)

	if !m.mappedDevice[index].isMapped {
		fmt.Printf("FATAL Error in memory CodeRead.  Attempt to read non mapped memory.\n")
		fmt.Printf("address is [%08X]\n", address)
		runtime.Goexit()
	}

	value := m.mappedDevice[index].readData(subAddress)

	return value
}

// Write takes an address and a value and writes to
// the memory map.  Could be ram or
// memory mapped device like a Serial Port
func (m *TMemory) Write(address uint32, value uint16) {

	index, subAddress := _helper(address)

	if !m.mappedDevice[index].isMapped {
		fmt.Printf("FATAL Error in memory write.  Attempt to write non mapped memory.\n")
		fmt.Printf("address is [%08X]\n", address)
		runtime.Goexit()
	}

	m.mappedDevice[index].writeData(subAddress, value)
}

// AddDevice maps a device based on an addresRange
func (m *TMemory) AddDevice(addressRange int,
	read func(address uint32) uint16,
	write func(address uint32, value uint16)) {

	if m.mappedDevice[addressRange].isMapped {
		fmt.Printf("Tried to add device to existing mem map location!")
		os.Exit(1)
	}
	fmt.Printf("Added device with CS %d\n", addressRange)

	m.mappedDevice[addressRange].readData = read
	m.mappedDevice[addressRange].writeData = write
	m.mappedDevice[addressRange].isMapped = true
}

// dump is an interactive function which lets the user
// specify an area of memory to dump
// func (m *Memory) dump() {
// 	s := cli.RawInput("Enter starting address (in hex) >")

// 	n, _ := strconv.ParseUint(s, 16, 32)
// 	startingAddress := uint32(n)

// 	size := uint32(16)

// 	for i := uint32(0); i < size; i++ {
// 		var s string
// 		workingAddress := startingAddress + i
// 		value := m.read(workingAddress)
// 		if value >= 32 && value <= 126 {
// 			s = fmt.Sprintf("%s", string(value))
// 		} else {
// 			s = "NP"
// 		}
// 		fmt.Printf("  %04X: %04X %3s\n", workingAddress, value, s)
// 	}
// }

// load403File - uses Original Pat loader format from 2006!
// It is interactive and prompts for a file name.
// TODO create a return value to show success or failure
// func load403File() {

// 	// read4 returns uint16 by reading 4 HEX digits from f
// 	read4 := func(f *os.File) uint16 {
// 		b := make([]byte, 4)
// 		f.Read(b)
// 		s := string(b)
// 		n, _ := strconv.ParseUint(s, 16, 32)
// 		w := uint16(n)
// 		return (w)

// 	}

// 	filename := cli.RawInput("Enter 403 file name >")
// 	fileInfo, err := os.Stat(filename)
// 	if err != nil {
// 		fmt.Printf("Could not stat file [%s].  Returning\n", filename)
// 		return
// 	}
// 	actualFileSize := fileInfo.Size()
// 	fmt.Printf("File Size is %d\n", actualFileSize)
// 	if actualFileSize < 12 {
// 		fmt.Printf("File is too small to be a valid 403 file; returning...\n")
// 		return
// 	}

// 	f, err := os.Open(filename)
// 	if err != nil {
// 		fmt.Printf("Could not open %s\n", filename)
// 		return
// 	}

// 	objectLength := read4(f)
// 	requiredFileSize := 4 + 4 + 4*objectLength
// 	if actualFileSize != int64(requiredFileSize) {
// 		fmt.Printf("file size mismatch; returning...\n")
// 	}
// 	startAddress := read4(f)

// 	fmt.Printf("Setting PC to [%04X]\n", startAddress)
// 	mycpu.SetPC(startAddress)

// 	memoryAddress := uint32(0x0403)
// 	for {
// 		if objectLength == 0 {
// 			break
// 		}
// 		dataWord := read4(f)
// 		ram.write(memoryAddress, dataWord)
// 		memoryAddress++
// 		objectLength--
// 	}

// }

// mycpu.ReadCodeMemory = ram.read
// mycpu.ReadDataMemory = ram.read
// mycpu.WriteDataMemory = ram.write
// mycpu.InterruptCallback = interruptController1.GetOutput
