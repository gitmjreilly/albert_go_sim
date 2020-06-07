package serialport

import (
	"albert_go_sim/intmaxmin"
	"fmt"
	"net"
	"time"
)

const (
	txBufferSize = 1024
	rxBufferSize = 1024
)

const (
	numRXTicksPerByte = 1200
	numTXTicksPerByte = 1200
)

// SerialPort provides virtual serial port implemented with TCP
type SerialPort struct {
	receiveBuffer             [rxBufferSize]uint8
	transmitBuffer            [txBufferSize]uint8
	numBytesInReceiveBuffer   int
	numBytesInTransmitBuffer  int
	remainingTransmitTime     int
	remaingingReceiveTime     int
	memory                    [16]int16
	serialConnection          net.Conn
	statusReg                 uint16
	inputChannel              chan uint8
	numTicksSinceReception    int
	numTicksSinceTransmission int
}

func (s *SerialPort) pollInput() {

}

// Init must be called before the serial port is used
func (s *SerialPort) Init() {
	ln, err := net.Listen("tcp", ":5000")
	if err != nil {
		fmt.Printf("Fatal error could not listen for serial port")
		panic("Done.")
	}
	fmt.Printf("   Listen succeeded\n")
	fmt.Printf("   Connect your virtual terminal to TCP 5000\n")
	connection, err := ln.Accept()
	if err != nil {
		fmt.Printf("Fatal error could not listen for serial port")
		panic("Done.")
	}
	s.serialConnection = connection
	fmt.Printf("   Accept succeeded\n")
	time.Sleep(1 * time.Second)

	s.numTicksSinceReception = 1000000
	s.numTicksSinceTransmission = 1000000
	s.statusReg = 0x0001
	s.inputChannel = make(chan uint8, 0)

	poll := func() {
		b := make([]uint8, 1)

		for {
			// _ = s.serialConnection.SetReadDeadline(time.Now().Add(500 * time.Nanosecond))
			// This Read will block until a byte arrives
			// numRead, _ := s.serialConnection.Read(b)
			s.serialConnection.Read(b)
			s.inputChannel <- b[0]

		}

	}

	go poll()

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
			fmt.Printf("Got b from receive channel\n")
			s.receiveBuffer[0] = b
			s.statusReg |= 0x0002
			s.numTicksSinceReception = 0
		default:
			break
		}
	}

	s.numTicksSinceTransmission++
	s.numTicksSinceTransmission = intmaxmin.Min(s.numTicksSinceTransmission, numTXTicksPerByte)
	if s.numBytesInTransmitBuffer == 0 {
		return
	}

	// If we got this far, there's something to transmit
	// We also have to ensure enough (tick) time has passed since the last transmission
	if s.numTicksSinceTransmission < numTXTicksPerByte {
		return
	}

	b := byte(s.transmitBuffer[0])
	var byteSlice []byte
	byteSlice = append(byteSlice, b)
	s.serialConnection.Write(byteSlice)
	s.numTicksSinceTransmission = 0
	s.numBytesInTransmitBuffer = 0
	// Mark the transmitter free bit
	s.statusReg |= 0x0001

}

// Write takes address and value
// 0 is the data port
func (s *SerialPort) Write(address uint16, value uint16) {
	s.transmitBuffer[0] = uint8(value)
	s.numBytesInTransmitBuffer = 1
	// Clear the transmitter free bit
	s.statusReg &= 0xFFFE
}

// Read takes address and returns a value
// 0 is the data port
// 1 is the status port
// 0x0002 (bit) is set when byte had been received
func (s *SerialPort) Read(address uint16) uint16 {

	value := uint16(0)
	if address == 0 {
		value = uint16(s.receiveBuffer[0])
		// Note in the status register that no data is available
		s.statusReg &= 0xFFFD
	} else {
		if address == 1 {
			value = s.statusReg
		}
	}
	return value
}
