package cpu

import (
	"fmt"
)

const (
	andOpcode         = 27
	branchOpcode      = 4
	branchFalseOpcode = 12
	doLitOpcode       = 2
	dropOpcode        = 7
	dupOpcode         = 19
	equallOpcode      = 31
	fetchOpcode       = 9
	fromROpcode       = 14
	haltOpcode        = 3
	jsrOpcode         = 10
	lessOpcode        = 5
	mulOpcode         = 30
	negOpcode         = 26
	orOpcode          = 28
	overOpcode        = 22
	plusOpcode        = 24
	rFetchOpcode      = 18
	retOpcode         = 11
	storeOpcode       = 8
	subOpcode         = 25
	swapOpcode        = 21
	toROpcode         = 13
	xorOpcode         = 29
)

const (
	true  = 0xFFFF
	false = 0x0000
)

// Status contains a snapshot along with disassembly
type Status struct {
	cpuStruct       CPU
	absoluteAddress uint16
	pStack          [4]uint16
	rStack          [4]uint16
}

const historySize = 1024 * 1024

var historyIndex = 0
var history [historySize]Status

var snapShot Status

// CPU struct matches hardware
type CPU struct {
	PC              uint16
	cs              uint16
	DS              uint16
	ES              uint16
	PSP             uint16
	RSP             uint16
	PTOS            uint16
	RTOS            uint16
	ReadDataMemory  func(uint16) uint16
	WriteDataMemory func(uint16, uint16)
	ReadCodeMemory  func(uint16) uint16
	history         []Status
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

// SetPC allows direct setting of the the cpu's PC
func (c *CPU) SetPC(pc uint16) {
	c.PC = pc
}

// push is the cpu's internal push operation used by almost every instruction
// e.g. DO_LIT and PLUS...
func (c *CPU) push(v uint16) {
	// fmt.Printf("Entered c.push v is %04x\n", v)
	// fmt.Printf("  PTOS is %04x\n", c.PTOS)
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
func (c *CPU) Tick() int {
	var absoluteAddress uint16 = c.cs<<4 + c.PC
	opCode := c.ReadDataMemory(absoluteAddress)
	c.PC++
	status := c.doInstruction(opCode, absoluteAddress)
	return (status)
}

// DoInstruction takes an opCode and its current absoluteAddress
// It assumes the PC already points after the location where the
// this opCode is stored.
func (c *CPU) doInstruction(opCode uint16, absoluteAddress uint16) int {
	// disassemblyString := fmt.Sprintf("%08x : ", absoluteAddress)
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

	// stackString := fmt.Sprintf("PSTACK => %04X %04X %04X PTOS: %04X RSTACK => %04X %04X %04X RTOS: %04X",
	// 	pstackBuffer[3], pstackBuffer[2], pstackBuffer[1], pstackBuffer[0],
	// 	rstackBuffer[3], rstackBuffer[2], rstackBuffer[1], rstackBuffer[0])

	// Create a bunch of short cut names for uuse witth the  disasembbly
	leftOperand := c.ReadDataMemory(scaledDS + c.PSP - 1)
	rightOperand := c.PTOS
	inlineOperand := c.ReadDataMemory(c.PC + scaledCS)
	rtosOperand := c.RTOS

	snapShot.absoluteAddress = absoluteAddress
	snapShot.cpuStruct = *c
	snapShot.pStack = pstackBuffer
	snapShot.rStack = rstackBuffer

	disassemblyString := ""
	stackString := ""

	// a b AND
	if opCode == andOpcode {
		disassemblyString += fmt.Sprintf("[%04X %04X] AND | %s", leftOperand, rightOperand, stackString)
		logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a & b)
		return (0)
	}

	// BRA dst
	if opCode == branchOpcode {
		disassemblyString += fmt.Sprintf("BRA %04X | %s", inlineOperand, stackString)
		logInstruction(snapShot)

		destinationAddress := c.consumeInstructionLiteral() + scaledCS
		c.PC = destinationAddress

		return (0)
	}

	// f JMPF dst
	if opCode == branchFalseOpcode {
		disassemblyString += fmt.Sprintf("[%04X] JMPF %04X | %s", rightOperand, inlineOperand, stackString)
		logInstruction(snapShot)

		flag := c.pop()
		destinationAddress := c.consumeInstructionLiteral()
		if flag == false {
			c.PC = destinationAddress
		}

		return (0)
	}

	// DOLIT l
	if opCode == doLitOpcode {
		disassemblyString += fmt.Sprintf("DO_LIT %04X | %s", inlineOperand, stackString)
		logInstruction(snapShot)

		l := c.consumeInstructionLiteral()
		c.push(l)

		return (0)
	}

	// a DROP
	if opCode == dropOpcode {
		disassemblyString += fmt.Sprintf("[%04X] DROP | %s", rightOperand, stackString)
		logInstruction(snapShot)

		c.pop()

		return (0)
	}

	// a DUP
	if opCode == dupOpcode {
		disassemblyString += fmt.Sprintf("[%04X] DUP | %s", rightOperand, stackString)
		logInstruction(snapShot)

		a := c.pop()
		c.push(a)
		c.push(a)
		return (0)
	}

	// a b EQUAL
	if opCode == equallOpcode {
		disassemblyString += fmt.Sprintf("[%04X %04X] EQUAL | %s", leftOperand, rightOperand, stackString)
		logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		if a == b {
			c.push(true)
		} else {
			c.push(false)
		}

		return (0)
	}

	// d FETCH
	if opCode == fetchOpcode {
		disassemblyString += fmt.Sprintf("[%04X] FETCH | %s", rightOperand, stackString)
		logInstruction(snapShot)

		destinationAddress := c.pop() + scaledDS
		v := c.ReadDataMemory(destinationAddress)
		c.push(v)

		return (0)
	}

	// (RTOS) FROM_R
	if opCode == fromROpcode {
		disassemblyString += fmt.Sprintf("[RTOS: %04X] FROM_R | %s", rtosOperand, stackString)
		logInstruction(snapShot)

		a := c.rPop()
		c.push(a)

		return (0)
	}

	// JSR d
	if opCode == jsrOpcode {
		disassemblyString += fmt.Sprintf("JSR %04X | %s", inlineOperand, stackString)
		logInstruction(snapShot)

		destinationAddress := c.consumeInstructionLiteral()
		c.rPush(c.PC)
		c.PC = destinationAddress

		return (0)
	}

	// HALT
	if opCode == haltOpcode {
		disassemblyString += fmt.Sprintf("HALT")
		logInstruction(snapShot)
		return (1)
	}

	// a b LESS
	if opCode == lessOpcode {
		disassemblyString += fmt.Sprintf("[%04X %04X] LESS | %s", leftOperand, rightOperand, stackString)
		logInstruction(snapShot)

		b := int16(c.pop())
		a := int16(c.pop())
		if a < b {
			c.push(true)
		} else {
			c.push(false)
		}
		return (0)
	}

	// a b *
	if opCode == mulOpcode {
		disassemblyString += fmt.Sprintf("[%04X %04X] MUL | %s", leftOperand, rightOperand, stackString)
		logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a * b)
		return (0)
	}

	// a NEG?
	if opCode == negOpcode {
		disassemblyString += fmt.Sprintf("[%04X] NEG? | %s", rightOperand, stackString)
		logInstruction(snapShot)

		a := int16(c.pop())
		if a < 0 {
			c.push(true)
		} else {
			c.push(false)
		}
		return (0)
	}

	// a b OR
	if opCode == orOpcode {
		disassemblyString += fmt.Sprintf("[%04X %04X] OR | %s", leftOperand, rightOperand, stackString)
		logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a | b)

		return (0)
	}

	// a b +
	if opCode == plusOpcode {
		disassemblyString += fmt.Sprintf("[%04X %04X] PLUS | %s", leftOperand, rightOperand, stackString)
		logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a + b)
		return (0)
	}

	// RET
	if opCode == retOpcode {
		disassemblyString += fmt.Sprintf("RET [%04X] | %s", c.RTOS, stackString)
		logInstruction(snapShot)

		c.PC = c.rPop()

		return 0
	}

	// val addr STORE
	if opCode == storeOpcode {
		disassemblyString += fmt.Sprintf("[%04X %04X] STORE | %s", leftOperand, rightOperand, stackString)
		logInstruction(snapShot)

		destinationAddress := c.pop() + scaledDS
		val := c.pop()
		c.WriteDataMemory(destinationAddress, val)

		return 0
	}

	// a b -
	if opCode == subOpcode {
		disassemblyString += fmt.Sprintf("[%04X %04X] SUB | %s", leftOperand, rightOperand, stackString)
		logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a - b)
		return (0)

	}

	// a b SWAP
	if opCode == swapOpcode {
		disassemblyString += fmt.Sprintf("[%04X %04X] SWAP | %s", leftOperand, rightOperand, stackString)
		logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(b)
		c.push(a)

		return 0
	}

	// a TO_R
	if opCode == toROpcode {
		disassemblyString += fmt.Sprintf("[PTOS: %04X] TO_R | %s", rightOperand, stackString)
		logInstruction(snapShot)

		a := c.pop()
		c.rPush(a)

		return 0

	}

	// a b XOR
	if opCode == xorOpcode {
		disassemblyString += fmt.Sprintf("[%04X %04X] XOR | %s", leftOperand, rightOperand, stackString)
		logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a ^ b)

		return (0)
	}

	fmt.Printf("Unknown opcode [%04X] address [%08X]\n", opCode, absoluteAddress)
	return (2)
}

// logInstruction takes a disassembly string and stores in the history buffer
func logInstruction(s Status) {
	history[historyIndex] = s
	historyIndex++
	historyIndex %= historySize
	return
}
