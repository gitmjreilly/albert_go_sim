package serialport

import (
	"fmt"
	"net"
)

const (
	txBufferSize = 1024
	rxBufferSize = 1024
)

// SerialPort provides virtual serial port implemented with TCP
type SerialPort struct {
	receiveBuffer            [rxBufferSize]int16
	transmitBuffer           [txBufferSize]int16
	numBytesInReceiveBuffer  int
	numBytesInTransmitBuffer int
	remainingTransmitTime    int
	remaingingReceiveTime    int
	memory                   [16]int16
	serialConnection         net.Conn
}

// Init must be called before the serial port is used
func (s *SerialPort) Init() {
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

// Write takes address and value
// 0 is the data port
func (s *SerialPort) Write(address uint16, value uint16) {
	var byteSlice []byte
	b := byte(value)
	byteSlice = append(byteSlice, b)
	s.serialConnection.Write(byteSlice)
}
