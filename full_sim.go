package main

import (
	"albert_go_sim/cli"
	"albert_go_sim/clock"
	"albert_go_sim/counter"
	"albert_go_sim/cpu"
	"albert_go_sim/interruptcontroller"
	"albert_go_sim/memory"
	"albert_go_sim/ram"
	"albert_go_sim/rom"
	"albert_go_sim/serialport"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

const (
	numTicksPerSecond = 10000000
)

var wg sync.WaitGroup

var mycpu cpu.CPU
var mem memory.TMemory
var counter1 counter.Counter
var clock1 clock.Clock
var rom1 rom.Rom
var ram1 ram.RAM

var isKeyboardInterrupt bool = false
var numClockTicks uint64 = 0
var numSecondsTick uint32 = 0
var humanTime uint32 = 0

var consoleSerialPort serialport.SerialPort
var diskControllerPort serialport.SerialPort
var terminalControllerPort serialport.SerialPort
var interruptController1 interruptcontroller.InterruptController

// Init initializes the global runtime for the simulator
func Init() {
	var enableControllers string

	// First we initialize all of the devices e.g. serial ports and cpu
	// Then we add them to the memory map.
	// Then we connect to the interrupt controller (via callbacks) if necessary
	// We also have to provide memory callbacks to the cpu
	counter1.Init()

	consoleSerialPort.Init("Console Serial Port", 5000)
	for {
		enableControllers = cli.RawInput("Do you want to init disk and term controllers (y/n) >")
		if (enableControllers == "y") || (enableControllers == "n") {
			break
		}
	}

	if enableControllers == "y" {
		diskControllerPort.Init("Disk Controller", 5600)
		terminalControllerPort.Init("Terminal Controller", 6000)
	}

	interruptController1.Init()

	mycpu.Init()
	// ram does not need to be initialized
	rom1.Init()

	// Add all of the devices to the memory map
	// This corresponds to the chip select glue logic in the hardware
	mem.AddDevice(memory.RomCS, rom1.Read, rom1.Write)
	mem.AddDevice(memory.RAMCS, ram1.Read, ram1.Write)
	mem.AddDevice(memory.F000, consoleSerialPort.Read, consoleSerialPort.Write)
	mem.AddDevice(memory.F010, interruptController1.Read, interruptController1.Write)
	if enableControllers == "y" {
		mem.AddDevice(memory.F030, terminalControllerPort.Read, terminalControllerPort.Write)
		mem.AddDevice(memory.F090, diskControllerPort.Read, diskControllerPort.Write)
	}

	// Connect sources to the interrupt controller.
	// The assignments are boolean callbacks
	interruptController1.Callbacks[1] = counter1.CounterIsZero

	interruptController1.Callbacks[5] = terminalControllerPort.RXIsQuarterFull

	interruptController1.Callbacks[4] = diskControllerPort.RXIsHalfFull

	clock1.Frequency = 10000000
	clock1.DoPrint = true

	mycpu.ReadCodeMemory = mem.ReadCodeMemory
	mycpu.ReadDataMemory = mem.Read
	mycpu.WriteDataMemory = mem.Write
	mycpu.InterruptCallback = interruptController1.GetOutput

	// There's a little bit of magic here.  We've created a goroutine
	// so that we can sets a global var
	// to indicate the user has pressed CTL-C
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

func resetComputer() {
	mycpu.PC = 0
	mycpu.RSP = 0xFE00
	mycpu.PSP = 0xFF00
	mycpu.CS = 0
	mycpu.DS = 0
	mycpu.ES = 0
	mycpu.IntCtlLow = 0
	cpu.History.Clear()
	ram1.Clear()
	consoleSerialPort.Reset()
	diskControllerPort.Reset()
	terminalControllerPort.Reset()
	fmt.Printf("The computer has been reset.\n")
}

// runSimulator increments the global time
// and "drives" all of the hardware by calling
// the Tick() functions.
// mode == 0 for continuous running
// mode == 1 for single stepping
func runSimulator(mode int) {
	defer wg.Done()

	fmt.Printf("Running simulator\n")
	for {
		if isKeyboardInterrupt {
			fmt.Printf("Simulation stopped by keyboard interrupt\n")
			isKeyboardInterrupt = false
			break
		}

		clock1.Tick()

		consoleSerialPort.Tick()
		diskControllerPort.Tick()
		terminalControllerPort.Tick()
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
				if status == cpu.Halt {
					fmt.Printf("\n  *** Saw Halt instruction ***\n\n")
				}
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
			fmt.Printf("Number of ticks since simulation started TOBEFIXED : %d\n", numClockTicks)
			break
		}
		if status == cpu.BreakPoint {
			fmt.Printf("Encountered breakpoint\n")
			break
		}

	}

}

func showStacks() {
	stackString := "PSTACK => "
	for i := uint32(10); i > 0; i-- {
		stackString += fmt.Sprintf("%04X ", mem.Read(uint32(mycpu.PSP)-i))
	}
	stackString += fmt.Sprintf("PTOS:%04X", mycpu.PTOS)
	fmt.Println(stackString)

	stackString = "RSTACK => "
	for i := uint32(10); i > 0; i-- {
		stackString += fmt.Sprintf("%04X ", mem.Read(uint32(mycpu.RSP)-i))
	}
	stackString += fmt.Sprintf("RTOS:%04X", mycpu.RTOS)
	fmt.Println(stackString)

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

	filename := cli.RawInput("Enter 403 file name >")
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

	memoryAddress := uint32(0x0403)
	for {
		if objectLength == 0 {
			break
		}
		dataWord := read4(f)
		mem.Write(memoryAddress, dataWord)
		memoryAddress++
		objectLength--
	}

}

// loadV4File -
// It is interactive and prompts for a file name.
// TODO create a return value to show success or failure
func loadV4File() {

	// read4 returns uint16 by reading 4 HEX digits from f
	read4 := func(f *os.File) uint16 {
		b := make([]byte, 2)
		f.Read(b)
		w := uint16(b[0])*256 + uint16(b[1])
		return (w)

	}

	filename := cli.RawInput("Enter V4 file name >")
	fileInfo, err := os.Stat(filename)
	if err != nil {
		fmt.Printf("Could not stat file [%s].  Returning\n", filename)
		return
	}
	actualFileSize := fileInfo.Size()
	fmt.Printf("File Size is %08X\n", actualFileSize)
	if actualFileSize < 14 {
		fmt.Printf("File is too small to be a valid V4 file; returning...\n")
		return
	}

	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Could not open V4 file %s\n", filename)
		return
	}

	magic1 := read4(f)
	magic2 := read4(f)
	if (magic1 != 0) || (magic2 != 4) {
		fmt.Printf("Incorrect magic %04X %04X expected 0000:0004\n", magic1, magic2)
		return
	}

	codeSize := read4(f)
	codeLoadAddress := read4(f)
	codeStartAddress := read4(f)
	dataSize := read4(f)
	dataLoadAddress := read4(f)

	fmt.Printf("Code Size [%04X]\n", codeSize)
	fmt.Printf("Code Load Address [%04X]\n", codeLoadAddress)
	fmt.Printf("Code Start Address [%04X]\n", codeStartAddress)

	fmt.Printf("Data Size [%04X]\n", dataSize)
	fmt.Printf("Data Load Address [%04X]\n", dataLoadAddress)
	requiredFileSize := 7*2 + (2 * int(codeSize)) + (2 * int(dataSize))
	fmt.Printf("Required File Size is [%08X]\n", requiredFileSize)

	if actualFileSize != int64(requiredFileSize) {
		fmt.Printf("file size mismatch; returning...\n")
	}

	fmt.Printf("Actual File Size and Calculated File Size match. Continuing...\n")

	memoryAddress := uint32(codeLoadAddress)
	for {
		if codeSize == 0 {
			break
		}
		dataWord := read4(f)
		mem.Write(memoryAddress, dataWord)
		memoryAddress++
		codeSize--
	}

	memoryAddress = uint32(dataLoadAddress)
	for {
		if dataSize == 0 {
			break
		}
		dataWord := read4(f)
		mem.Write(memoryAddress, dataWord)
		memoryAddress++
		dataSize--
	}

	mycpu.PC = codeStartAddress

}

func setPC() {
	s := cli.RawInput("Enter PC (in hex) >")

	n, _ := strconv.ParseUint(s, 16, 32)
	mycpu.PC = uint16(n)
}

func helpMessage() {
	fmt.Printf("HELP\n")
	fmt.Printf("   r - run the simulator\n")
	fmt.Printf("   s - step simulator\n")
	fmt.Printf("   S - Show stacks\n")
	fmt.Printf("   b - Set break point\n")
	fmt.Printf("   B - Show Break points\n")
	// fmt.Printf("   l - load a 403 file\n")
	fmt.Printf("   L - Load V4 file (for Bilal!)\n")
	fmt.Printf("   m - dump memory\n")
	fmt.Printf("   d - display CPU status\n")
	fmt.Printf("   c - clear break point\n")
	fmt.Printf("   H - display History\n")
	fmt.Printf("   p - Set PC\n")
	fmt.Printf("   R - reset computer\n")
	fmt.Printf("   q - quit the simulator\n")
}

func main() {

	Init()

	for {
		selection := cli.RawInput("Enter menu choice >")

		if selection == "h" {
			helpMessage()
			continue
		}

		if selection == "H" {
			cpu.History.Display(1000)
			continue
		}

		if selection == "b" {
			mycpu.SetBreakPoint()
			continue
		}

		if selection == "L" {
			loadV4File()
			continue
		}

		if selection == "p" {
			setPC()
			continue
		}

		if selection == "c" {
			mycpu.ClearBreakPoint()
			continue
		}

		if selection == "B" {
			mycpu.ShowBreakPoints()
		}

		if selection == "d" {
			mycpu.ShowStatus()
			interruptController1.ShowStatus()
			continue
		}

		if selection == "r" {
			wg.Add(1)
			go runSimulator(0)
			fmt.Printf("Started waiting in full sim\n")
			wg.Wait()
			fmt.Printf("Finished waiting in full sim\n")
			continue
		}

		if selection == "R" {
			resetComputer()
			continue
		}

		// if selection == "l" {
		// 	load403File()
		// 	continue
		// }

		if selection == "s" {
			wg.Add(1)
			go runSimulator(1)
			wg.Wait()
			mycpu.ShowStatus()
			continue
		}

		if selection == "m" {
			mem.Dump()
			continue
		}

		if selection == "S" {
			showStacks()
		}

		if selection == "q" {
			break
		}

	}

}
