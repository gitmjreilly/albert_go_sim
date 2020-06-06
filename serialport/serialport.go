package serialport

import (
	"fmt"
	"net"
	"time"
)

const (
	txBufferSize = 1024
	rxBufferSize = 1024
)

// SerialPort provides virtual serial port implemented with TCP
type SerialPort struct {
	receiveBuffer            [rxBufferSize]uint8
	transmitBuffer           [txBufferSize]uint8
	numBytesInReceiveBuffer  int
	numBytesInTransmitBuffer int
	remainingTransmitTime    int
	remaingingReceiveTime    int
	memory                   [16]int16
	serialConnection         net.Conn
	statusReg                uint16
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
	time.Sleep(2 * time.Second)
	fmt.Fprintf(s.serialConnection, "Hello from the simulator\n")
	s.statusReg = 0

}

// Tick should be called on every tick off the virtual clock
func (s *SerialPort) Tick() {

	return

	// b := make([]byte, 1)
	// // var t time.Time
	// _ = s.serialConnection.SetReadDeadline(time.Now().Add(500 * time.Nanosecond))
	// numRead, _ := s.serialConnection.Read(b)
	// // fmt.Printf("in serial Tick num read is  [%d]\n", numRead)
	// if numRead == 1 {
	// 	s.receiveBuffer[0] = b[0]
	// 	s.statusReg = 0x0002
	// }

}

// Write takes address and value
// 0 is the data port
func (s *SerialPort) Write(address uint16, value uint16) {
	var byteSlice []byte
	b := byte(value)
	byteSlice = append(byteSlice, b)
	s.serialConnection.Write(byteSlice)
}

// Read takes address and returns a value
// 0 is the data port
// 1 is the status port
// 0x0002 (bit) is set when byte had been received
func (s *SerialPort) Read(address uint16) uint16 {

	fmt.Printf(" in sp read addr is %04X\n", address)
	value := uint16(0)
	if address == 0 {
		value = uint16(s.receiveBuffer[0])
		s.statusReg = 0
	} else {
		if address == 1 {
			fmt.Printf("DEBUG reading sp statuss reg\n")
			value = s.statusReg
		}
	}
	return value
}
