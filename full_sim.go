package main

import (
	"fmt"
	"albert_go_sim/cpu"
	"net"
	"os"
	"strconv"
)

var mycpu cpu.CPU
var ram Memory

// SerialPort1 is the console
//
var SerialPort1 serialPort

type serialPort struct {
	receiveBuffer            [1024]int16
	transmitBuffer           [1024]int16
	numBytesInReceiveBuffer  int
	numBytesInTransmitBuffer int
	remainingTransmitTime    int
	remaingingReceiveTime    int
	memory                   [16]int16
	serialConnection         net.Conn
}

func (s *serialPort) Init() {
	ln, err := net.Listen("tcp", ":5000")
	if err != nil {
		fmt.Printf("Fatal error could not listen for serial port")
		panic("Done.")
	}
	fmt.Printf("   Listen succeeded\n")
	connection, err := ln.Accept()
	if err != nil {
		fmt.Printf("Fatal error could not listen for serial port")
		panic("Done.")
	}
	s.serialConnection = connection
	fmt.Printf("   Accept succeeded\n")
	fmt.Fprintf(s.serialConnection, "Hello from the simulator\n")
}

func (s *serialPort) write(address uint16, value uint16) {
	var byteSlice []byte
	b := byte(value)
	byteSlice = append(byteSlice, b)
	s.serialConnection.Write(byteSlice)
}

// Memory matches hardware
type Memory struct {
	memory [1024 * 1024 * 8]uint16
}

func (m *Memory) read(address uint16) uint16 {
	return (m.memory[address])
}

func (m *Memory) write(address uint16, value uint16) {
	if address >= 0xF000 && address <= 0xF00F {
		address -= 0xF000
		SerialPort1.write(address, value)

	}
	m.memory[address] = value
}

// (s SerialPort) func tick() {
// 	if remainingTransmitTime > 0 {
// 		remainingTransmitTime--
// 	}
// 	if remainingReceiveTime > 0 {
// 		remainingReceiveTime--
// 	}
// 	if numByptes
// }

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

func main() {

	// TODO restore serial port
	// SerialPort1.Init()

	mycpu.Init()

	mycpu.ReadCodeMemory = ram.read
	mycpu.ReadDataMemory = ram.read
	mycpu.WriteDataMemory = ram.write

	// scanner := bufio.NewScanner(os.Stdin)
	// shellPrompt := "Simulate >"

	// Fill RAM with instructions

	fmt.Printf("Loading a 403 object file...\n")
	load403File()

	// ram.write(0, 2)
	// ram.write(1, 5)

	// ram.write(2, 2)
	// ram.write(3, 9)

	// ram.write(4, 2)
	// ram.write(5, 7)

	// ram.write(6, 2)
	// ram.write(7, 66)

	// ram.write(8, 2)
	// ram.write(9, 0xF000)

	// ram.write(10, 8)

	// ram.write(11, 3)

	for {

		status := mycpu.Tick()
		if status != 0 {
			fmt.Printf("Saw non zero return status; breaking\n")
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

	// for i := 0xFF00; i < (0xFF00 + 10); i++ {
	// 	value := ram.read(uint16(i))
	// 	fmt.Printf("Address %04X Value %04X\n", uint16(i), value)
	// }

	// fmt.Printf("PTOS is : %04X\n", mycpu.PTOS)
	fmt.Printf("Simulation is finished\n")

}
