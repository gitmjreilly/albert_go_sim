package main

import (
	"albert_go_sim/cli"
	"albert_go_sim/clock"
	"albert_go_sim/counter"
	"albert_go_sim/cpu"
	"albert_go_sim/interruptcontroller"
	"albert_go_sim/serialport"
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"
)

const (
	numTicksPerSecond = 10000000
)

var mycpu cpu.CPU
var ram Memory
var counter1 counter.Counter
var clock1 clock.Clock

var isKeyboardInterrupt bool = false
var numClockTicks uint64 = 0
var numSecondsTick uint32 = 0
var humanTime uint32 = 0

// SerialPort1 is the console
//
var SerialPort1 serialport.SerialPort
var interruptController1 interruptcontroller.InterruptController

// Memory matches hardware
type Memory struct {
	memory [1024 * 1024 * 8]uint16
}

func (m *Memory) read(address uint16) uint16 {
	if address >= 0xF000 && address <= 0xF00F {
		address -= 0xF000
		return SerialPort1.Read(address)
	}
	if address >= 0xF010 && address <= 0xF01F {
		address -= 0xF010
		return interruptController1.Read(address)
	}

	return (m.memory[address])
}

func (m *Memory) write(address uint16, value uint16) {
	if address >= 0xF000 && address <= 0xF00F {
		address -= 0xF000
		SerialPort1.Write(address, value)

	}
	if address >= 0xF010 && address <= 0xF01F {
		address -= 0xF010
		interruptController1.Write(address, value)
	}

	m.memory[address] = value
}

// dump is an interactive function which lets the user
// specify an area of memory to dump
func (m *Memory) dump() {
	s := cli.RawInput("Enter starting address (in hex) >")

	n, _ := strconv.ParseUint(s, 16, 32)
	startingAddress := uint16(n)

	size := uint16(16)

	for i := uint16(0); i < size; i++ {
		var s string
		workingAddress := startingAddress + i
		value := m.read(workingAddress)
		if value >= 32 && value <= 126 {
			s = fmt.Sprintf("%s", string(value))
		} else {
			s = "NP"
		}
		fmt.Printf("  %04X: %04X %3s\n", workingAddress, value, s)
	}
}

func loadPatsLoader() {
	fmt.Printf("Entered loadPatsLoader\n")
	f, err := os.Open("loader_from_zero.txt")
	if err != nil {
		fmt.Printf("Could not open loader file\n")
		fmt.Printf("Error is %v\n", err)
		return
	}
	scanner := bufio.NewScanner(f)
	address := uint16(0)
	for scanner.Scan() {
		s := scanner.Text()
		n, _ := strconv.ParseUint(s, 16, 32)
		w := uint16(n)
		ram.write(address, w)
		address++
	}
	f.Close()

}

// load403File - uses Original Pat loader format from 2006!
// It is interactive and prompts for a file name.
// TODO create a return value to show success or failure
func load403File() {

	// read4 returns uint16 by reading 4 HEX digits from f
	read4 := func(f *os.File) uint16 {
		b := make([]byte, 4)
		f.Read(b)
		s := string(b)
		n, _ := strconv.ParseUint(s, 16, 32)
		w := uint16(n)
		return (w)

	}

	fmt.Printf("Enter 403 filename>")
	// scanner := bufio.NewScanner(os.Stdin)

	// scanner.Scan()
	// filename := scanner.Text()
	filename := "obj"
	fileInfo, err := os.Stat(filename)
	if err != nil {
		fmt.Printf("Could not stat file [%s].  Returning\n", filename)
		return
	}
	actualFileSize := fileInfo.Size()
	fmt.Printf("File Size is %d\n", actualFileSize)
	if actualFileSize < 12 {
		fmt.Printf("File is too small to be a valid 403 file; returning...\n")
		return
	}

	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Could not open %s\n", filename)
		return
	}

	objectLength := read4(f)
	requiredFileSize := 4 + 4 + 4*objectLength
	if actualFileSize != int64(requiredFileSize) {
		fmt.Printf("file size mismatch; returning...\n")
	}
	startAddress := read4(f)

	fmt.Printf("Setting PC to [%04X]\n", startAddress)
	mycpu.SetPC(startAddress)

	memoryAddress := uint16(0x0403)
	for {
		if objectLength == 0 {
			break
		}
		dataWord := read4(f)
		ram.write(memoryAddress, dataWord)
		memoryAddress++
		objectLength--
	}

}

// Init initializes the global runtime for the simulator
func Init() {
	SerialPort1.Init()
	interruptController1.Callbacks[0] = counter1.CounterIsZero

	clock1.Frequency = 10000000
	clock1.DoPrint = true

	mycpu.Init()

	mycpu.ReadCodeMemory = ram.read
	mycpu.ReadDataMemory = ram.read
	mycpu.WriteDataMemory = ram.write
	mycpu.InterruptCallback = interruptController1.GetOutput

	loadPatsLoader()

	go func() {
		signalChannel := make(chan os.Signal, 2)
		// When SIGINT occurs send signal to signalChannel
		signal.Notify(signalChannel, os.Interrupt)
		for {
			<-signalChannel
			isKeyboardInterrupt = true
			time.Sleep(1 * time.Second)
		}
	}()
}

// runSimulator increments the global time
// and "drives" all of the hardware by calling
// the Tick() functions.
// mode == 0 for continuous running
// mode == 1 for single stepping
func runSimulator(mode int) {

	fmt.Printf("Running simulator\n")
	for {
		if isKeyboardInterrupt {
			fmt.Printf("Simulation stopped by keyboard interrupt\n")
			isKeyboardInterrupt = false
			break
		}

		clock1.Tick()

		// numClockTicks++
		// numSecondsTick++

		// if numSecondsTick == numTicksPerSecond {
		// 	numSecondsTick = 0
		// 	humanTime++
		// 	fmt.Printf("(simulated) elapsed time (secs)%5d\n", humanTime)

		// }

		SerialPort1.Tick()
		counter1.Tick()
		interruptController1.Tick()

		// Check to see if caller only wants to single step
		// because, if so, we may have to call Tick() multiple
		// times before the cpu actually steps
		if mode == 1 {
			numTicks := 0
			for {
				status := mycpu.Tick()
				numTicks++
				// Keep ticking until status is NOT "tick only" (100)
				if status != 100 {
					break
				}
			}
			fmt.Printf("Single Stepped. NumTicks was %d\n", numTicks)
			return
		}

		// If we got this far, we are in "continuous running mode"
		status := mycpu.Tick()
		if status == cpu.Halt {
			fmt.Printf("Saw cpu Tick status == 1 indicating a HALT; breaking\n")
			fmt.Printf("Number of ticks since simulation started : %d\n", numClockTicks)
			break
		}

	}

}

func helpMessage() {
	fmt.Printf("HELP\n")
	fmt.Printf("   r - run the simulator\n")
	fmt.Printf("   s - STEP simulator")
	fmt.Printf("   m - dump memory\n")
	fmt.Printf("   H - display history\n")
	fmt.Printf("   q - quit the simulator\n")
}

func main() {

	Init()

	for {
		selection := cli.RawInput("Simulator (h for help) >")

		if selection == "h" {
			helpMessage()
			continue
		}

		if selection == "H" {
			cpu.History.Display(1000)
			continue
		}

		if selection == "r" {
			runSimulator(0)
			continue
		}

		if selection == "s" {
			runSimulator(1)
			mycpu.ShowStatus()
			continue
		}

		if selection == "m" {
			ram.dump()
		}

		if selection == "q" {
			break
		}

	}

	fmt.Printf("PC is %04X\n", mycpu.PC)
	fmt.Printf("PSP is %04X\n", mycpu.PSP)
	fmt.Printf("RSP is %04X\n", mycpu.RSP)
	fmt.Printf("Dumping stack\n")

	stackString := "PSTACK => "
	for i := 5; i > 0; i-- {
		stackString += fmt.Sprintf("%04X ", ram.read(mycpu.PSP-uint16(i)))
	}
	stackString += fmt.Sprintf("PTOS:%04X", mycpu.PTOS)
	fmt.Println(stackString)

	stackString = "RSTACK => "
	for i := 5; i > 0; i-- {
		stackString += fmt.Sprintf("%04X ", ram.read(mycpu.RSP-uint16(i)))
	}
	stackString += fmt.Sprintf("RTOS:%04X", mycpu.RTOS)
	fmt.Println(stackString)

	fmt.Printf("Simulation is finished\n")

}
