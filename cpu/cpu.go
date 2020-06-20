package cpu

import (
	"albert_go_sim/intmaxmin"
	"fmt"
)

// These are constants indicating return status
// from CPU.doInstruction
const (
	Normal  = iota
	Halt    = iota
	Unknown = iota
)

// Opcode values
const (
	andOpcode         = 27 // Done
	branchOpcode      = 4  // Done
	branchFalseOpcode = 12 // Done
	diOpcode          = 37 // Done
	doLitOpcode       = 2  // Done
	dropOpcode        = 7  // Done
	dupOpcode         = 19 // Done
	eiOpcode          = 35 // Done
	equallOpcode      = 31 //  Done
	fetchOpcode       = 9  // Done
	fromROpcode       = 14 // Done
	haltOpcode        = 3  // Done
	jsrOpcode         = 10 // Done
	jsrintOpcode      = 33 // Done
	lessOpcode        = 5  // Done and Properly signed
	lvarOpcode        = 51 // Done
	mulOpcode         = 30 // Done
	negOpcode         = 26 // Properly Signed and Done
	nopOpcode         = 1  // Done
	orOpcode          = 28 // done
	overOpcode        = 22 // Done
	plusOpcode        = 24 // Done
	rFetchOpcode      = 18 // Done
	retOpcode         = 11 // Done
	retiOpcode        = 34 // Done
	sLessOpcode       = 50 // Done
	spFetchOpcode     = 20 // Done
	spStoreOpcode     = 23 // Done
	storeOpcode       = 8  // Done
	store2Opcode      = 52 // Done
	subOpcode         = 25 // Done
	swapOpcode        = 21 // Done
	toROpcode         = 13 // Done
	xorOpcode         = 29 // Done
)

const (
	true  = 0xFFFF
	false = 0x0000
)

const (
	ticksPerInstruction = 8
)

// History contains the run time history of the CPU
var History tHistory

// CPU struct matches hardware
type CPU struct {
	PC                uint16
	cs                uint16
	DS                uint16
	ES                uint16
	PSP               uint16
	RSP               uint16
	PTOS              uint16
	RTOS              uint16
	IntCtlLow         uint8
	ReadDataMemory    func(uint16) uint16
	WriteDataMemory   func(uint16, uint16)
	ReadCodeMemory    func(uint16) uint16
	history           []Status
	InterruptCallback func() bool
	tickNum           int
}

// Init sets up the cpu before the first instruction is run
func (c *CPU) Init() {
	c.PC = 0
	c.PSP = 0xFF00
	c.RSP = 0xFE00
	c.cs = 0
	c.DS = 0
	c.ES = 0

}

// ShowStatus prints the internal state of the CPU
func (c *CPU) ShowStatus() {
	fmt.Printf("CPU State :\n")
	fmt.Printf("PC    : %04X\n", c.PC)
	fmt.Printf("PTOS  : %04X  RTOS : %04X\n", c.PTOS, c.RTOS)
	fmt.Printf("PSP   : %04X  RSP  : %04X\n", c.PSP, c.RSP)
	fmt.Printf("Interrupt state %v\n", c.InterruptCallback())
	fmt.Printf("\n")
}

// SetPC allows direct setting of the the cpu's PC
func (c *CPU) SetPC(pc uint16) {
	c.PC = pc
}

// push is the cpu's internal push operation used by almost every instruction
// e.g. DO_LIT and PLUS...
func (c *CPU) push(v uint16) {
	scaledDS := uint16(c.DS << 4)

	address := c.PSP + scaledDS
	c.WriteDataMemory(address, c.PTOS)
	c.PSP++
	c.PTOS = v
}

// pop is the cpu's internal data stack pop operation, used by many instructions
// It returns the top of the parameter stack
func (c *CPU) pop() uint16 {
	scaledDS := uint16(c.DS << 4)
	v := c.PTOS
	c.PSP--
	c.PTOS = c.ReadDataMemory(c.PSP + scaledDS)
	return v
}

// rPush is the cpu's internal push to the return stack
func (c *CPU) rPush(v uint16) {
	scaledDS := uint16(c.DS << 4)
	c.WriteDataMemory(c.RSP+scaledDS, c.RTOS)
	c.RSP++
	c.RTOS = v
}

// rPop is the cpu's internal return stack pop operation
// It returns the top of the RETURN stack
func (c *CPU) rPop() uint16 {
	scaledDS := uint16(c.DS << 4)
	v := c.RTOS
	c.RSP--
	c.RTOS = c.ReadDataMemory(c.RSP + scaledDS)
	return v
}

// consumeInstructionLiteral returns literal from the instruction stream and advances PC
func (c *CPU) consumeInstructionLiteral() uint16 {
	scaledCS := uint16(c.cs << 4)
	literal := c.ReadDataMemory(c.PC + scaledCS)
	c.PC++
	return literal
}

// Tick should be called for each "tick" of the virtual clock.
// Return value indicates normal operation, halt seen or other TBD.
// return values  0 = cpu stepped normally, 100 = tick only
func (c *CPU) Tick() int {
	c.tickNum = intmaxmin.IncMod(c.tickNum, 1, ticksPerInstruction)
	if c.tickNum != 0 {
		return (100)
	}

	var absoluteAddress uint16 = c.cs<<4 + c.PC

	if c.InterruptCallback() && ((c.IntCtlLow & 0x01) == 1) {
		// Notice the PC has not been incremented.
		// This is because the JSR should return to the PC location
		// that was interrupted
		status := c.doInstruction(jsrintOpcode, absoluteAddress)
		return status
	}

	opCode := c.ReadDataMemory(absoluteAddress)
	c.PC++
	status := c.doInstruction(opCode, absoluteAddress)
	return (status)
}

// DoInstruction takes an opCode and its current absoluteAddress
// It assumes the PC already points after the location where the
// this opCode is stored.
// return int const Normal, Halt or Unknown
// return 1 for HALT
func (c *CPU) doInstruction(opCode uint16, absoluteAddress uint16) int {
	var snapShot Status

	scaledCS := uint16(c.cs << 4)
	scaledDS := uint16(c.DS << 4)
	// scaledES := c.ES << 4

	var pstackBuffer [4]uint16
	for i := 3; i > 0; i-- {
		pstackBuffer[i] = c.ReadDataMemory(scaledDS + c.PSP - uint16(i))
	}
	pstackBuffer[0] = c.PTOS

	var rstackBuffer [4]uint16
	for i := 3; i > 0; i-- {
		rstackBuffer[i] = c.ReadDataMemory(scaledDS + c.RSP - uint16(i))
	}
	rstackBuffer[0] = c.RTOS

	// Create a bunch of short cut names for uuse witth the  disasembbly
	leftOperand := c.ReadDataMemory(scaledDS + c.PSP - 1)
	rightOperand := c.PTOS
	inlineOperand := c.ReadDataMemory(c.PC + scaledCS)

	snapShot.absoluteAddress = absoluteAddress
	snapShot.cpuStruct = *c
	snapShot.pStack = pstackBuffer
	snapShot.rStack = rstackBuffer
	snapShot.opCode = opCode
	snapShot.leftOperand = leftOperand
	snapShot.rightOperand = rightOperand
	snapShot.inlineOperand = inlineOperand

	// a b AND
	if opCode == andOpcode {
		History.logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a & b)

		return (Normal)
	}

	// BRA dst
	if opCode == branchOpcode {
		History.logInstruction(snapShot)

		destinationAddress := c.consumeInstructionLiteral() + scaledCS
		c.PC = destinationAddress

		return Normal
	}

	// f JMPF dst
	if opCode == branchFalseOpcode {
		History.logInstruction(snapShot)

		flag := c.pop()
		destinationAddress := c.consumeInstructionLiteral()
		if flag == false {
			c.PC = destinationAddress
		}

		return Normal
	}

	// DI
	if opCode == diOpcode {
		History.logInstruction(snapShot)

		c.IntCtlLow = c.IntCtlLow & 0xFE

		return Normal
	}

	// DOLIT l
	if opCode == doLitOpcode {
		History.logInstruction(snapShot)

		l := c.consumeInstructionLiteral()
		c.push(l)

		return Normal
	}

	// a DROP
	if opCode == dropOpcode {
		History.logInstruction(snapShot)

		c.pop()

		return Normal
	}

	// a DUP
	if opCode == dupOpcode {
		History.logInstruction(snapShot)

		a := c.pop()
		c.push(a)
		c.push(a)

		return Normal
	}

	// a b EQUAL
	if opCode == equallOpcode {
		History.logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		if a == b {
			c.push(true)
		} else {
			c.push(false)
		}

		return Normal
	}

	// EI
	if opCode == eiOpcode {
		History.logInstruction(snapShot)

		c.IntCtlLow |= 0x0001

		return Normal
	}

	// d FETCH
	if opCode == fetchOpcode {
		History.logInstruction(snapShot)

		destinationAddress := c.pop() + scaledDS
		v := c.ReadDataMemory(destinationAddress)
		c.push(v)

		return Normal
	}

	// (RTOS) FROM_R
	if opCode == fromROpcode {
		History.logInstruction(snapShot)

		a := c.rPop()
		c.push(a)

		return Normal
	}

	// JSR d
	if opCode == jsrOpcode {
		History.logInstruction(snapShot)

		destinationAddress := c.consumeInstructionLiteral()
		c.rPush(c.PC)
		c.PC = destinationAddress

		return Normal
	}

	// JSRINT
	// rPush sequence should match rPop sequene in RETI
	if opCode == jsrintOpcode {
		tmpRSP := c.RSP
		c.rPush(c.DS)
		c.rPush(c.cs)
		c.rPush(c.ES)
		c.rPush(c.PSP)
		c.rPush(c.PTOS)
		c.rPush(c.PC)
		c.rPush(uint16(c.IntCtlLow))
		c.rPush(tmpRSP)

		c.IntCtlLow = c.IntCtlLow & 0xFE
		c.PC = 0xFD00
		c.cs = 0x0000

		return Normal
	}

	// RETI
	// rPop sequence should match rPush sequene in JSRINT
	if opCode == retiOpcode {
		History.logInstruction(snapShot)

		tmpRSP := c.rPop()
		c.IntCtlLow = uint8(c.rPop())
		c.PC = c.rPop()
		c.PTOS = c.rPop()
		c.PSP = c.rPop()
		c.ES = c.rPop()
		c.cs = c.rPop()
		c.DS = c.rPop()
		c.RSP = tmpRSP

		return Normal
	}

	// HALT
	if opCode == haltOpcode {
		History.logInstruction(snapShot)

		return Halt
	}

	// a b LESS
	// a b S_LESS
	// (accidentally implemented signed less twice!)
	// Notice we have to cast stack  values
	// to int16 because all 16 bit values in
	// simulation are considered signed
	if opCode == lessOpcode || opCode == sLessOpcode {
		History.logInstruction(snapShot)

		b := int16(c.pop())
		a := int16(c.pop())
		if a < b {
			c.push(true)
		} else {
			c.push(false)
		}

		return Normal
	}

	// L_VAR n
	if opCode == lvarOpcode {
		History.logInstruction(snapShot)

		offset := c.consumeInstructionLiteral()
		c.push(offset + c.RTOS)

		return Normal
	}

	// a b *
	if opCode == mulOpcode {
		History.logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a * b)

		return Normal
	}

	// a NEG?
	// Notice we have to cast stack  value
	// to int16 because all 16 bit values in
	// simulation are considered signed
	if opCode == negOpcode {
		History.logInstruction(snapShot)

		a := int16(c.pop())
		if a < 0 {
			c.push(true)
		} else {
			c.push(false)
		}

		return Normal
	}

	// NOP
	if opCode == nopOpcode {
		History.logInstruction(snapShot)

		// Do nothing

		return Normal
	}

	// a b OR
	if opCode == orOpcode {
		History.logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a | b)

		return Normal
	}

	//  OVER
	//  BEFORE   AFTER
	//           x
	//  n        n
	//  x        x
	if opCode == overOpcode {
		History.logInstruction(snapShot)

		n := c.pop()
		x := c.pop()
		c.push(x)
		c.push(n)
		c.push(x)

		return Normal
	}

	// a b +
	if opCode == plusOpcode {
		History.logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a + b)

		return Normal
	}

	// RET
	if opCode == retOpcode {
		History.logInstruction(snapShot)

		c.PC = c.rPop()

		return Normal
	}

	// R_FETCH
	if opCode == rFetchOpcode {
		History.logInstruction(snapShot)

		c.push(c.RTOS)

		return Normal
	}

	// SP_FETCH
	if opCode == spFetchOpcode {
		History.logInstruction(snapShot)

		c.push(c.PSP)

		return Normal
	}

	// SP_STORE
	if opCode == spStoreOpcode {
		History.logInstruction(snapShot)

		c.PSP = c.PTOS
		return Normal
	}

	// val addr STORE
	if opCode == storeOpcode {
		History.logInstruction(snapShot)

		destinationAddress := c.pop() + scaledDS
		val := c.pop()
		c.WriteDataMemory(destinationAddress, val)

		return Normal
	}

	// addr val STORE2
	if opCode == store2Opcode {
		History.logInstruction(snapShot)

		val := c.pop()
		destinationAddress := c.pop() + scaledDS
		c.WriteDataMemory(destinationAddress, val)

		return Normal
	}

	// a b -
	if opCode == subOpcode {
		History.logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a - b)

		return Normal
	}

	// a b SWAP
	if opCode == swapOpcode {
		History.logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(b)
		c.push(a)

		return Normal
	}

	// a TO_R
	if opCode == toROpcode {
		History.logInstruction(snapShot)

		a := c.pop()
		c.rPush(a)

		return Normal
	}

	// a b XOR
	if opCode == xorOpcode {
		History.logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a ^ b)

		return Normal
	}

	fmt.Printf("Unknown opcode [%04X] address [%08X]\n", opCode, absoluteAddress)
	return Unknown
}
