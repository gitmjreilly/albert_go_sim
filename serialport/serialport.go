package serialport

import (
	"albert_go_sim/intmaxmin"
	"fmt"
	"net"
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
	receiveBuffer             [receiverBufferSize]uint8
	receiveBufferIn           int
	receiveBufferOut          int
	transmitBuffer            [transmitBufferSize]uint8
	transmitBufferIn          int
	transmitBufferOut         int
	numBytesInReceiveBuffer   int
	numBytesInTransmitBuffer  int
	remainingTransmitTime     int
	remaingingReceiveTime     int
	memory                    [16]int16
	serialConnection          net.Conn
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
	s.numTicksSinceTransmission = 0
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
			if s.numBytesInReceiveBuffer == len(s.receiveBuffer) {
				fmt.Printf("WARNING serial receiver buffer over run")
			}
			s.receiveBuffer[s.receiveBufferIn] = b
			s.receiveBufferIn = intmaxmin.IncMod(s.receiveBufferIn, 1, len(s.receiveBuffer))
			s.numBytesInReceiveBuffer = intmaxmin.Min(s.numBytesInReceiveBuffer+1, len(s.receiveBuffer))
			s.numTicksSinceReception = 0
		default:
			break
		}
	}

	// Since another clock tick has occurred, Note how much time has passed
	// since the last byte was transmitted, capping the time at the
	// the number of ticks since the last transmission so
	// numTicksSinceTransmission doesn't grow unbounded
	s.numTicksSinceTransmission = intmaxmin.Min(s.numTicksSinceTransmission+1, numTXTicksPerByte)

	// Is there any data to transmit?  If not there's nothing more to do
	if s.numBytesInTransmitBuffer == 0 {
		return
	}

	// If we got this far, there's something to transmit, but
	// we also have to ensure enough (tick) time has passed since the last transmission
	if s.numTicksSinceTransmission < numTXTicksPerByte {
		return
	}

	b := byte(s.transmitBuffer[s.transmitBufferOut])
	s.transmitBufferOut = intmaxmin.IncMod(s.transmitBufferOut, 1, len(s.transmitBuffer))
	s.numBytesInTransmitBuffer = intmaxmin.Max(s.numBytesInTransmitBuffer-1, 0)

	var byteSlice []byte
	byteSlice = append(byteSlice, b)
	s.serialConnection.Write(byteSlice)
	s.numTicksSinceTransmission = 0
}

// Write takes address and value.
// 0 is the data port.
// no other address is valid.
func (s *SerialPort) Write(address uint16, value uint16) {
	fmt.Printf("\nEntered serial port write\n")
	if address != 0 {
		fmt.Printf("WARNING Tried to write to serial port address %04x != 0; returning\n", address)
		return
	}
	if s.numBytesInTransmitBuffer == len(s.transmitBuffer) {
		fmt.Printf("WARNING you are writing to an already full serial transmit buffer; proceeding\n")
	}
	fmt.Printf("Debug about to write %d to txbufferIn %d\n", value, s.transmitBufferIn)
	fmt.Printf("num byte in tx buffer prior to write is %d\n", s.numBytesInTransmitBuffer)
	s.transmitBuffer[s.transmitBufferIn] = byte(value)
	s.numBytesInTransmitBuffer = intmaxmin.Min(s.numBytesInTransmitBuffer+1, len(s.transmitBuffer))
	s.transmitBufferIn = (s.transmitBufferIn + 1) % len(s.transmitBuffer)
}

// Read takes address and returns a value
// 0 is the data port
// 1 is the status port
// 0x0002 (bit) is set when byte had been received
func (s *SerialPort) Read(address uint16) uint16 {

	value := uint16(0)
	if address == 0 {
		// Users wants to read received serial data.
		// Do some sanity checks along the way
		if s.numBytesInReceiveBuffer == 0 {
			fmt.Printf("WARNING trying to read from empty serial receive buffer; returning\n")
			return 0
		}

		// Consume from reciever buffer, including updating index and
		// decrementing number
		value = uint16(s.receiveBuffer[s.receiveBufferOut])
		s.receiveBufferOut = intmaxmin.IncMod(s.receiveBufferOut, 1, len(s.receiveBuffer))
		s.numBytesInReceiveBuffer = intmaxmin.Max(0, s.numBytesInReceiveBuffer-1)
		return value
	}
	// address 1 is the compatibility address for original
	// serial port from 2006
	// 0x0001 == transmitter is available for transmission
	// 0x0002 == receiver has a byte to be read
	if address == 1 {
		if s.numBytesInTransmitBuffer < len(s.transmitBuffer) {
			value |= 0x0001
		}
		if s.numBytesInReceiveBuffer > 0 {
			value |= 0x0002
		}
		return value
	}

	if address == 2 {
		if s.numBytesInTransmitBuffer == 0 {
			value = 0x0001
		}
		return value
	}

	return 0
}
