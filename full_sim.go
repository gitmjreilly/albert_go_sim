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
	"time"
)

const (
	numTicksPerSecond = 10000000
)

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
	consoleSerialPort.Init("Console Serial Port", 5000)
	interruptController1.Callbacks[0] = counter1.CounterIsZero

	clock1.Frequency = 10000000
	clock1.DoPrint = true

	mycpu.Init()
	rom1.Init()

	mycpu.ReadCodeMemory = mem.Read
	mycpu.ReadDataMemory = mem.Read
	mycpu.WriteDataMemory = mem.Write
	mycpu.InterruptCallback = interruptController1.GetOutput

	mem.AddDevice(memory.RomCS, rom1.Read, rom1.Write)
	mem.AddDevice(memory.RAMCS, ram1.Read, ram1.Write)

	mem.AddDevice(memory.F000, consoleSerialPort.Read, consoleSerialPort.Write)

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
			runSimulator(0)
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
