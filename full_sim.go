package main

import (
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

var mycpu cpu.CPU
var ram Memory
var counter1 counter.Counter

var isKeyboardInterrupt bool = false

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

func runSimulator(mode int) {

	fmt.Printf("Running simulator\n")
	numSecondsTick := 0
	numClockTicks := 1
	humanTime := 0.0
	for {
		if isKeyboardInterrupt {
			fmt.Printf("Simulation stopped by keyboard interrupt\n")
			isKeyboardInterrupt = false
			break
		}
		numClockTicks++
		if numClockTicks == 1000000000 {
			break
		}
		numSecondsTick++

		if numSecondsTick == 10000000 {
			numSecondsTick = 0
			humanTime++

			fmt.Printf("human time %f\n", humanTime)

		}
		SerialPort1.Tick()
		counter1.Tick()
		interruptController1.Tick()

		status := mycpu.Tick()
		if status != 0 {
			fmt.Printf("Saw non zero cpu Tick status; breaking\n")
			fmt.Printf("Number of ticks since simulation started : %d\n", numClockTicks)
			break
		}

	}

}

func helpMessage() {
	fmt.Printf("HELP\n")
	fmt.Printf("   r - run the simulator\n")
	fmt.Printf("   H - display history\n")
	fmt.Printf("   q - quit the simulator\n")
}

func main() {

	Init()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf("Simulator (h for help) >")
		scanner.Scan()
		selection := scanner.Text()
		fmt.Printf("Your selection was [%s]\n", selection)

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
