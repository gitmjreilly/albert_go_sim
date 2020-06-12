package cpu

import (
	"albert_go_sim/intmaxmin"
	"fmt"
)

// Opcode values
const (
	andOpcode         = 27
	branchOpcode      = 4
	branchFalseOpcode = 12
	diOpcode          = 37
	doLitOpcode       = 2
	dropOpcode        = 7
	dupOpcode         = 19
	eiOpcode          = 35
	equallOpcode      = 31
	fetchOpcode       = 9
	fromROpcode       = 14
	haltOpcode        = 3
	jsrOpcode         = 10
	jsrintOpcode      = 33
	lessOpcode        = 5
	lvarOpcode        = 51
	mulOpcode         = 30
	negOpcode         = 26
	nopOpcode         = 1
	orOpcode          = 28
	overOpcode        = 22
	plusOpcode        = 24
	rFetchOpcode      = 18
	retOpcode         = 11
	retiOpcode        = 34
	sLessOpcode       = 50
	spFetchOpcode     = 20
	spStoreOpcode     = 23
	storeOpcode       = 8
	store2Opcode      = 52
	subOpcode         = 25
	swapOpcode        = 21
	toROpcode         = 13
	xorOpcode         = 29
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
	c.tickNum = intmaxmin.IncMod(c.tickNum, 1, ticksPerInstruction)
	if c.tickNum != 0 {
		return (0)
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
func (c *CPU) doInstruction(opCode uint16, absoluteAddress uint16) int {
	var snapShot Status

	// snapShot.disassemblyString := fmt.Sprintf("%08x : ", absoluteAddress)
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
	pspOperand := c.PSP

	snapShot.absoluteAddress = absoluteAddress
	snapShot.cpuStruct = *c
	snapShot.pStack = pstackBuffer
	snapShot.rStack = rstackBuffer
	snapShot.opCode = opCode
	stackString := ""

	// a b AND
	if opCode == andOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[%04X %04X] AND | %s", leftOperand, rightOperand, stackString)
		History.logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a & b)
		return (0)
	}

	// BRA dst
	if opCode == branchOpcode {
		snapShot.disassemblyString = fmt.Sprintf("BRA %04X | %s", inlineOperand, stackString)
		History.logInstruction(snapShot)

		destinationAddress := c.consumeInstructionLiteral() + scaledCS
		c.PC = destinationAddress

		return (0)
	}

	// f JMPF dst
	if opCode == branchFalseOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[%04X] JMPF %04X | %s", rightOperand, inlineOperand, stackString)
		History.logInstruction(snapShot)

		flag := c.pop()
		destinationAddress := c.consumeInstructionLiteral()
		if flag == false {
			c.PC = destinationAddress
		}

		return (0)
	}

	// DI
	if opCode == diOpcode {
		snapShot.disassemblyString = fmt.Sprintf("DI | %s", stackString)
		History.logInstruction(snapShot)
		c.IntCtlLow = c.IntCtlLow & 0xFE
		return 0
	}

	// DOLIT l
	if opCode == doLitOpcode {
		snapShot.disassemblyString = fmt.Sprintf("DO_LIT %04X | %s", inlineOperand, stackString)
		History.logInstruction(snapShot)

		l := c.consumeInstructionLiteral()
		c.push(l)

		return (0)
	}

	// a DROP
	if opCode == dropOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[%04X] DROP | %s", rightOperand, stackString)
		History.logInstruction(snapShot)

		c.pop()

		return (0)
	}

	// a DUP
	if opCode == dupOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[%04X] DUP | %s", rightOperand, stackString)
		History.logInstruction(snapShot)

		a := c.pop()
		c.push(a)
		c.push(a)
		return (0)
	}

	// a b EQUAL
	if opCode == equallOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[%04X %04X] EQUAL | %s", leftOperand, rightOperand, stackString)
		History.logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		if a == b {
			c.push(true)
		} else {
			c.push(false)
		}

		return (0)
	}

	// EI
	if opCode == eiOpcode {
		snapShot.disassemblyString = fmt.Sprintf("EI | %s", stackString)
		History.logInstruction(snapShot)
		fmt.Printf("Enabling Interrupts")
		c.IntCtlLow |= 0x0001
		return 0
	}

	// d FETCH
	if opCode == fetchOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[%04X] FETCH | %s", rightOperand, stackString)
		History.logInstruction(snapShot)

		destinationAddress := c.pop() + scaledDS
		v := c.ReadDataMemory(destinationAddress)
		c.push(v)

		return (0)
	}

	// (RTOS) FROM_R
	if opCode == fromROpcode {
		snapShot.disassemblyString = fmt.Sprintf("[RTOS: %04X] FROM_R | %s", rtosOperand, stackString)
		History.logInstruction(snapShot)

		a := c.rPop()
		c.push(a)

		return (0)
	}

	// JSR d
	if opCode == jsrOpcode {
		snapShot.disassemblyString = fmt.Sprintf("JSR %04X | %s", inlineOperand, stackString)
		History.logInstruction(snapShot)

		destinationAddress := c.consumeInstructionLiteral()
		c.rPush(c.PC)
		c.PC = destinationAddress

		return (0)
	}

	// JSRINT
	if opCode == jsrintOpcode {
		// fmt.Printf("Entered JSRINT\n")
		tmpRSP := c.RSP
		c.rPush(c.DS)
		c.rPush(c.cs)
		c.rPush(c.ES)
		c.rPush(c.PSP)
		c.rPush(c.PTOS)
		// fmt.Printf("  saving PC %4X\n", c.PC)
		c.rPush(c.PC)
		c.rPush(uint16(c.IntCtlLow))
		c.rPush(tmpRSP)

		c.IntCtlLow = c.IntCtlLow & 0xFE
		// fmt.Printf("intctl low is now %04X\n", c.IntCtlLow)
		c.PC = 0xFD00
		c.cs = 0x0000
		return 0
	}

	// RETI
	if opCode == retiOpcode {
		snapShot.disassemblyString = fmt.Sprintf("RETI | %s", stackString)
		History.logInstruction(snapShot)

		// fmt.Printf("entered RETI\n")
		tmpRSP := c.rPop()
		c.IntCtlLow = uint8(c.rPop())
		c.PC = c.rPop()
		c.PTOS = c.rPop()
		c.PSP = c.rPop()
		c.ES = c.rPop()
		c.cs = c.rPop()
		c.DS = c.rPop()
		c.RSP = tmpRSP
		return 0
	}

	// HALT
	if opCode == haltOpcode {
		snapShot.disassemblyString = fmt.Sprintf("HALT | %s", stackString)
		History.logInstruction(snapShot)
		return (1)
	}

	// a b LESS
	if opCode == lessOpcode || opCode == sLessOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[%04X %04X] LESS | %s", leftOperand, rightOperand, stackString)
		History.logInstruction(snapShot)

		b := int16(c.pop())
		a := int16(c.pop())
		if a < b {
			c.push(true)
		} else {
			c.push(false)
		}
		return (0)
	}

	// L_VAR n
	if opCode == lvarOpcode {
		snapShot.disassemblyString = fmt.Sprintf("L_VAR %004X | %s", inlineOperand, stackString)
		History.logInstruction(snapShot)

		offset := c.consumeInstructionLiteral()
		// tempRTOS := c.rPop()
		// c.rPush(tempRTOS)
		c.push(offset + c.RTOS)
		return (0)

	}

	// a b *
	if opCode == mulOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[%04X %04X] MUL | %s", leftOperand, rightOperand, stackString)
		History.logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a * b)
		return (0)
	}

	// a NEG?
	if opCode == negOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[%04X] NEG? | %s", rightOperand, stackString)
		History.logInstruction(snapShot)

		a := int16(c.pop())
		if a < 0 {
			c.push(true)
		} else {
			c.push(false)
		}
		return (0)
	}

	// NOP
	if opCode == nopOpcode {
		snapShot.disassemblyString = fmt.Sprintf("NOP | %s", stackString)
		History.logInstruction(snapShot)

		return (0)

	}

	// a b OR
	if opCode == orOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[%04X %04X] OR | %s", leftOperand, rightOperand, stackString)
		History.logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a | b)

		return (0)
	}

	// a b +
	if opCode == plusOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[%04X %04X] PLUS | %s", leftOperand, rightOperand, stackString)
		History.logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a + b)
		return (0)
	}

	// RET
	if opCode == retOpcode {
		snapShot.disassemblyString = fmt.Sprintf("RET [%04X] | %s", c.RTOS, stackString)
		History.logInstruction(snapShot)

		c.PC = c.rPop()

		return 0
	}

	// R_FETCH
	if opCode == rFetchOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[RTOS: %04X] R_FETCH | %s", rtosOperand, stackString)
		History.logInstruction(snapShot)

		c.push(c.RTOS)
		return 0
	}

	// SP_FETCH
	if opCode == spFetchOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[PSP: %04X] SP_FETCH | %s", pspOperand, stackString)
		History.logInstruction(snapShot)
		c.push(c.PSP)
		return 0
	}

	// SP_STORE
	if opCode == spStoreOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[PTOS: %04X] SP_STORE | %s", rightOperand, stackString)
		History.logInstruction(snapShot)

		c.PSP = c.PTOS
		return 0
	}

	// val addr STORE
	if opCode == storeOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[%04X %04X] STORE | %s", leftOperand, rightOperand, stackString)
		History.logInstruction(snapShot)

		destinationAddress := c.pop() + scaledDS
		val := c.pop()
		c.WriteDataMemory(destinationAddress, val)

		return 0
	}

	// addr val STORE2
	if opCode == store2Opcode {
		snapShot.disassemblyString = fmt.Sprintf("[%04X %04X] STORE | %s", leftOperand, rightOperand, stackString)
		History.logInstruction(snapShot)

		val := c.pop()
		destinationAddress := c.pop() + scaledDS
		c.WriteDataMemory(destinationAddress, val)

		return 0
	}

	// a b -
	if opCode == subOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[%04X %04X] SUB | %s", leftOperand, rightOperand, stackString)
		History.logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a - b)
		return (0)

	}

	// a b SWAP
	if opCode == swapOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[%04X %04X] SWAP | %s", leftOperand, rightOperand, stackString)
		History.logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(b)
		c.push(a)

		return 0
	}

	// a TO_R
	if opCode == toROpcode {
		snapShot.disassemblyString = fmt.Sprintf("[PTOS: %04X] TO_R | %s", rightOperand, stackString)
		History.logInstruction(snapShot)

		a := c.pop()
		c.rPush(a)

		return 0

	}

	// a b XOR
	if opCode == xorOpcode {
		snapShot.disassemblyString = fmt.Sprintf("[%04X %04X] XOR | %s", leftOperand, rightOperand, stackString)
		History.logInstruction(snapShot)

		b := c.pop()
		a := c.pop()
		c.push(a ^ b)

		return (0)
	}

	fmt.Printf("Unknown opcode [%04X] address [%08X]\n", opCode, absoluteAddress)
	return (2)
}
