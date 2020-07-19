package serialport

import (
	"albert_go_sim/intmaxmin"
	"fmt"
	"net"
	"runtime"
	"time"
)

const (
	transmitBufferSize = 1024
	receiverBufferSize = 1024
)

const (
	numRXTicksPerByte = 1200
	numTXTicksPerByte = 1200
)

// SerialPort provides virtual serial port implemented with TCP
type SerialPort struct {
	name                      string
	receiveFifo               fifo
	transmitFifo              fifo
	remaingingReceiveTime     int
	serialConnection          net.Conn
	inputChannel              chan uint8
	numTicksSinceReception    int
	numTicksSinceTransmission int
	isTransmitting            bool
	timeToTransmit            int
	transmitRegister          uint8
}

type fifo struct {
	data        []uint8
	in          int
	out         int
	numElements int
}

var byteNum int

// init initializes the fifo to size
func (f *fifo) init(size int) {
	f.data = make([]uint8, size)
}

// push adds an element to the fifo
// No check is made to see if the fifo is full
func (f *fifo) push(b uint8) {
	f.data[f.in] = b
	f.in = intmaxmin.IncMod(f.in, 1, len(f.data))
	f.numElements = intmaxmin.Min(f.numElements+1, len(f.data))
}

// Pop gets the next item from the Fifo
// No  check is made to see if the fifo is empty
func (f *fifo) pop() uint8 {
	v := f.data[f.out]
	f.out = intmaxmin.IncMod(f.out, 1, len(f.data))
	f.numElements = intmaxmin.Max(f.numElements-1, 0)
	return v
}

// isFull returns true when thee fifo is full
func (f *fifo) isFull() bool {
	return f.numElements == len(f.data)
}

// isEmpty returns true when the fifo is empty
func (f *fifo) isEmpty() bool {
	return f.numElements == 0
}

func (f *fifo) clear() {
	f.numElements = 0
	f.in = 0
	f.out = 0
}

//

// Init must be called before the serial port is used.
// It uses a raw tcpPortNum to simulate the connection.
// Connect to the "SerialPort" with a text TCP client.
// The name is for debugging in case the SerialPort needs
// to report an error.
func (s *SerialPort) Init(name string, tcpPortNum int) {
	fmt.Printf("Initializing serial port %s with port :%d\n", name, tcpPortNum)
	s.name = name

	s.transmitFifo.init(transmitBufferSize)
	s.receiveFifo.init(receiverBufferSize)

	portString := fmt.Sprintf(":%d", tcpPortNum)
	ln, err := net.Listen("tcp", portString)
	if err != nil {
		fmt.Printf("Fatal error could not listen for serial port")
		panic("Done.")
	}
	fmt.Printf("   Listen succeeded; connect now.\n")
	connection, err := ln.Accept()
	if err != nil {
		fmt.Printf("Fatal error could not listen for serial port")
		panic("Done.")
	}
	s.serialConnection = connection
	fmt.Printf("   Accept succeeded\n")
	time.Sleep(1 * time.Second)

	s.numTicksSinceReception = 1000000
	s.numTicksSinceTransmission = 0
	s.inputChannel = make(chan uint8, 10)

	poll := func() {
		b := make([]uint8, 1)

		for {
			s.serialConnection.Read(b)
			s.inputChannel <- b[0]
		}

	}

	go poll()

}

// Reset the serial port.  Use this when
// you want to ensure the fifo's are empty.
func (s *SerialPort) Reset() {
	s.transmitFifo.clear()
	s.receiveFifo.clear()
	s.isTransmitting = false
}

// Tick should be called on every tick off the virtual clock
func (s *SerialPort) Tick() {

	// On each Tick (of the clock) check for incoming bytes from port
	// and for bytes in the transmission buffer (as a result of a cpu write)
	s.numTicksSinceReception++
	s.numTicksSinceReception = intmaxmin.Min(s.numTicksSinceReception, numRXTicksPerByte)
	if s.numTicksSinceReception >= numRXTicksPerByte {
		select {
		case b := <-s.inputChannel:
			if s.receiveFifo.isFull() {
				fmt.Printf("WARNING receiver buffer is full.  Data overrun will occur.\n")
			}

			// byteNum++
			s.receiveFifo.push(b)

			s.numTicksSinceReception = 0
		default:
			break
		}
	}

	// Work on the transmit side

	// Were we in the middle of a transmission?
	if s.isTransmitting {
		s.timeToTransmit--
		s.timeToTransmit = intmaxmin.Max(s.timeToTransmit, 0)
		if s.timeToTransmit == 0 {
			// If we got this far, we had a pending transmission, let's actually
			// transmit.  The idea the "transmission" started a while ago, but the
			// full byte has finally been transmitted.
			// We don't transmit a bit at a time; we transmit the whole byte at the end.
			var byteSlice []byte
			byteSlice = append(byteSlice, s.transmitRegister)
			s.serialConnection.Write(byteSlice)
			s.numTicksSinceTransmission = 0
			s.isTransmitting = false
		}
		return
	}

	// If we got this far, should we begin a new transmission?
	if s.transmitFifo.isEmpty() {
		return
	}

	// If we got this far, the transmit fifo has more to send
	// and we know the simulated transmitter has been idle
	// Let's "begin" the transmission.
	s.transmitRegister = s.transmitFifo.pop()
	s.timeToTransmit = numTXTicksPerByte
	s.isTransmitting = true

}

// Write takes address and value.
// 0 is the data port.
// no other address is valid.
func (s *SerialPort) Write(address uint32, value uint16) {
	if address != 0 {
		fmt.Printf("WARNING Tried to write to serial port address %04x != 0; returning\n", address)
		return
	}

	if s.transmitFifo.isFull() {
		fmt.Printf("WARNING you are writing to an already full serial transmit buffer; this will result in overrun\n")
	}

	s.transmitFifo.push(uint8(value))

}

// Read takes address and returns a value
// 0 is the data port
// 1 is the status port
// 0x0002 (bit) is set when byte had been received
func (s *SerialPort) Read(address uint32) uint16 {

	value := uint16(0)
	if address == 0 {
		// User wants to read received serial data.
		// Do some sanity checks along the way
		if s.receiveFifo.isEmpty() {
			fmt.Printf("WARNING trying to read from empty serial receive buffer; returning\n")
			return 0
		}

		// Consume from receiver buffer, including updating index and
		// decrementing number
		value = uint16(s.receiveFifo.pop())

		return value
	}
	// address 1 is the compatibility address for original
	// serial port from 2006
	// 0x0001 == transmitter is available for transmission
	// 0x0002 == receiver has a byte to be read
	if address == 1 {
		if !s.transmitFifo.isFull() {
			value |= 0x0001
		}
		if !s.receiveFifo.isEmpty() {
			value |= 0x0002
		}
		return value
	}

	if address == 2 {
		if s.transmitFifo.isEmpty() {
			value = 0x0001
		}
		return value
	}

	if address == 3 {
		if s.transmitFifo.numElements < transmitBufferSize/2 {
			value = 0x0001
		}
		return value
	}

	if address == 4 {
		if s.transmitFifo.numElements == transmitBufferSize {
			value = 0x0001
		}
		return value
	}

	if address == 5 {
		if s.transmitFifo.isFull() {
			value = 0x0001
		}
		return value
	}

	if address == 6 { // OK
		if s.receiveFifo.isEmpty() {
			value = 0x0001
		}
		return value

	}

	if address == 7 {
		if s.receiveFifo.numElements >= receiverBufferSize/2 {
			value = 0x0001
		}
		return value
	}

	if address == 8 {
		if s.receiveFifo.numElements >= receiverBufferSize/4 {
			value = 0x0001
		}
		return value
	}

	if address == 9 {
		if s.receiveFifo.numElements == receiverBufferSize {
			value = 0x0001
		}
		return value
	}

	if address == 0xE {
		return uint16(s.receiveFifo.numElements)
	}

	if address == 0xF {
		return uint16(s.transmitFifo.numElements)
	}

	fmt.Printf("FATAL - tried to read from unmapped serial port address %02X in [%s]\n", address, s.name)
	runtime.Goexit()

	return 0
}

// RXIsHalfFull is a callback meant for use by an interrupt controller
func (s *SerialPort) RXIsHalfFull() bool {
	return s.receiveFifo.numElements >= receiverBufferSize/2
}

// RXIsQuarterFull is a callback meant for use by an interrupt controller
func (s *SerialPort) RXIsQuarterFull() bool {
	return s.receiveFifo.numElements >= receiverBufferSize/4
}
