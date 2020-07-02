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
	interruptController1.Callbacks[0] = counter1.CounterIsZero

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

	}

}

func helpMessage() {
	fmt.Printf("HELP\n")
	fmt.Printf("   r - run the simulator\n")
	fmt.Printf("   s - STEP simulator\n")
	fmt.Printf("   l - load a 403 file\n")
	fmt.Printf("   m - dump memory\n")
	fmt.Printf("   S - show CPU status\n")
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

		if selection == "S" {
			mycpu.ShowStatus()
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

		// if selection == "l" {
		// 	load403File()
		// 	continue
		// }

		if selection == "s" {
			runSimulator(1)
			mycpu.ShowStatus()
			continue
		}

		if selection == "m" {
			// ram.dump()
			continue
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
	for i := uint32(5); i > 0; i-- {
		stackString += fmt.Sprintf("%04X ", mem.Read(uint32(mycpu.PSP)-i))
	}
	stackString += fmt.Sprintf("PTOS:%04X", mycpu.PTOS)
	fmt.Println(stackString)

	stackString = "RSTACK => "
	for i := uint32(5); i > 0; i-- {
		stackString += fmt.Sprintf("%04X ", mem.Read(uint32(mycpu.RSP)-i))
	}
	stackString += fmt.Sprintf("RTOS:%04X", mycpu.RTOS)
	fmt.Println(stackString)

	fmt.Printf("Simulation is finished\n")

}
