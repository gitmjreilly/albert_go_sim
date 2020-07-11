package cpu

import (
	"albert_go_sim/cli"
	"albert_go_sim/intmaxmin"
	"fmt"
	"runtime"
	"strconv"
)

// These are constants indicating return status
// from CPU.doInstruction
const (
	Normal     = iota
	Halt       = iota
	BreakPoint = iota
	Unknown    = iota
)

// Opcode values
const (
	andOpcode         = 27  // Done - OK
	branchOpcode      = 4   // Done - OK
	branchFalseOpcode = 12  // Done - OK
	csFetchOpcode     = 43  // untested
	diOpcode          = 37  // untested
	doLitOpcode       = 2   // Done - OK
	dropOpcode        = 7   // Done - OK
	dsFetchOpcode     = 42  // untested
	dupOpcode         = 19  // Done - OK
	eiOpcode          = 35  // Done - OK
	equallOpcode      = 31  // Done - OK
	esFetchOpcode     = 41  // untested
	fetchOpcode       = 9   // Done - OK
	fromROpcode       = 14  // Done - OK
	haltOpcode        = 3   // Done - OK
	jsrOpcode         = 10  // Done - OK
	jsrintOpcode      = 33  // Done - OK
	kSpStoreOpcode    = 47  // Done
	lessOpcode        = 5   // Done - OK
	longFetchOpcode   = 44  // untested
	longStoreOpcode   = 45  // untested
	longTypeStore     = 999 // Not implemented yet (see Python version)
	lvarOpcode        = 51  // Done - OK
	mulOpcode         = 30  // Done - OK
	negOpcode         = 26  // Properly Signed and Done
	nopOpcode         = 1   // Done - OK
	orOpcode          = 28  // done - OK
	overOpcode        = 22  // untested
	plusOpcode        = 24  // Done - OK
	plusPlusOpcode    = 6   // untested
	popFOpcode        = 49  // untested
	pushFOpcode       = 48  // untested
	rFetchOpcode      = 18  // Done
	retOpcode         = 11  // Done - OK
	retiOpcode        = 34  // Done - OK
	rpFetchOpcode     = 16  // untested
	rpStoreOpcode     = 17  // untested
	sLessOpcode       = 50  // Done - OK (Same as less)
	sllOpcode         = 15  // untested
	spFetchOpcode     = 20  // Done - OK
	spStoreOpcode     = 23  // Done - OK
	sraOpcode         = 36  // untested
	srlOpcode         = 38  // untested
	storeOpcode       = 8   // Done - OK
	store2Opcode      = 52  // Done - OK
	subOpcode         = 25  // Done - OK
	swapOpcode        = 21  // Done
	sysCallOpcode     = 46  // Done - OK
	toDSOpcode        = 40  // untested
	toESOpcode        = 39  // untested
	toROpcode         = 13  // Done - OK
	umPlusOpcode      = 32  // Not implemented yet
	xorOpcode         = 29  // Done - OK
)

const (
	cpuTrue  = 0xFFFF
	cpuFalse = 0x0000
)

const (
	ticksPerInstruction = 8
)

// History contains the run time history of the CPU
var History tHistory

// CPU struct matches hardware
type CPU struct {
	PC        uint16
	CS        uint16
	DS        uint16
	ES        uint16
	PSP       uint16
	RSP       uint16
	PTOS      uint16
	RTOS      uint16
	IntCtlLow uint8
	// ReadDataMemory takes a 32bit address and returns 16 bit word
	// It should be used by instructions like FETCH.
	// Should not be used for instructions like DO_LIT
	ReadDataMemory            func(address uint32) uint16
	WriteDataMemory           func(address uint32, value uint16)
	ReadCodeMemory            func(address uint32) uint16
	history                   []Status
	InterruptCallback         func() bool
	tickNum                   int
	breakPoints               map[uint32]bool
	previousBreakPointAddress uint32
}

// Init sets up the cpu before the first instruction is run
func (c *CPU) Init() {
	c.PC = 0
	c.PSP = 0xFF00
	c.RSP = 0xFE00
	c.CS = 0
	c.DS = 0
	c.ES = 0

	c.breakPoints = make(map[uint32]bool)

}

// ShowStatus prints the internal state of the CPU
func (c *CPU) ShowStatus() {
	fmt.Printf("CPU State :\n")
	fmt.Printf("PC    : %04X\n", c.PC)
	fmt.Printf("PTOS  : %04X  RTOS : %04X\n", c.PTOS, c.RTOS)
	fmt.Printf("PSP   : %04X  RSP  : %04X\n", c.PSP, c.RSP)
	fmt.Printf("CS    : %04x  DS   : %04X   ES :  %04X\n", c.CS, c.DS, c.ES)
	fmt.Printf("IntCtl : %04x\n", c.IntCtlLow)
	fmt.Printf("Interrupt state %v\n", c.InterruptCallback())
	fmt.Printf("\n")
}

// SetBreakPoint interactively prompts for a break point address
func (c *CPU) SetBreakPoint() {
	s := cli.RawInput("Enter PC (in hex) for breakpoint>")
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		fmt.Printf("Invalid hex string.  Breakpoint was not set.\n")
		return
	}
	c.breakPoints[uint32(n)] = true
}

// ClearBreakPoint interactively prompts for a break point address
func (c *CPU) ClearBreakPoint() {
	s := cli.RawInput("Enter PC (in hex) for breakpoint to clear>")
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		fmt.Printf("Invalid hex string.  Breakpoint was not set.\n")
		return
	}
	delete(c.breakPoints, uint32(n))
}

// ShowBreakPoints prints all user defined break points
func (c *CPU) ShowBreakPoints() {
	for breakPoint := range c.breakPoints {
		fmt.Printf("%08X\n", breakPoint)
	}
}

// SetPC allows direct setting of the the cpu's PC
func (c *CPU) SetPC(pc uint16) {
	c.PC = pc
}

// push is the cpu's internal push operation used by almost every instruction
// e.g. DO_LIT and PLUS...
func (c *CPU) push(v uint16) {
	scaledDS := uint32(c.DS) << 4
	address := uint32(c.PSP) + scaledDS
	c.WriteDataMemory(address, c.PTOS)
	c.PSP++
	c.PTOS = v
}

// pop is the cpu's internal data stack pop operation, used by many instructions
// It returns the top of the parameter stack
func (c *CPU) pop() uint16 {
	v := c.PTOS
	c.PSP--
	scaledDS := uint32(c.DS) << 4
	address := uint32(c.PSP) + scaledDS
	c.PTOS = c.ReadDataMemory(address)
	return v
}

// rPush is the cpu's internal push to the return stack
func (c *CPU) rPush(v uint16) {
	scaledDS := uint32(c.DS) << 4
	address := uint32(c.RSP) + scaledDS
	c.WriteDataMemory(address, c.RTOS)
	c.RSP++
	c.RTOS = v
}

// rPop is the cpu's internal return stack pop operation
// It returns the top of the RETURN stack
func (c *CPU) rPop() uint16 {
	v := c.RTOS
	c.RSP--
	scaledDS := uint32(c.DS) << 4
	address := uint32(c.RSP) + scaledDS
	c.RTOS = c.ReadDataMemory(address)
	return v
}

// consumeInstructionLiteral returns literal from the instruction stream and advances PC
func (c *CPU) consumeInstructionLiteral() uint16 {
	scaledCS := uint32(c.CS) << 4
	address := uint32(c.PC) + scaledCS
	literal := c.ReadCodeMemory(address)
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

	// var absoluteAddress uint32 = uint32(c.CS<<4 + c.PC)
	absoluteAddress := uint32(c.CS)<<4 + uint32(c.PC)

	if c.InterruptCallback() && ((c.IntCtlLow & 0x01) == 1) {
		// Notice the PC has not been incremented.
		// This is because the JSR should return to the PC location
		// that was interrupted
		status := c.doInstruction(jsrintOpcode, absoluteAddress)
		return status
	}

	// Check for a breakpoint.  The tricky part is we have to remember the
	// previous address where a break point occurred.  This is because
	// we'll keep breaking at the same breakpoint after we reach it for the
	// first time.
	if c.breakPoints[absoluteAddress] && absoluteAddress != c.previousBreakPointAddress {
		fmt.Printf("Break point encountered at %08X\n", absoluteAddress)
		c.previousBreakPointAddress = absoluteAddress
		return (BreakPoint)
	}

	opCode := c.ReadCodeMemory(absoluteAddress)
	c.PC++
	status := c.doInstruction(opCode, absoluteAddress)
	return (status)
}

// DoInstruction takes an opCode and its current absoluteAddress
// It assumes the PC already points after the location where the
// this opCode is stored.
// return int const Normal, Halt or Unknown
// return 1 for HALT
func (c *CPU) doInstruction(opCode uint16, absoluteAddress uint32) int {
	var snapShot Status

	scaledCS := uint32(c.CS) << 4
	scaledDS := uint32(c.DS) << 4
	scaledES := uint32(c.ES) << 4

	var pstackBuffer [4]uint16
	for i := uint32(3); i > 0; i-- {
		address := scaledDS + uint32(c.PSP) - i
		pstackBuffer[i] = c.ReadDataMemory(address)
	}
	pstackBuffer[0] = c.PTOS

	var rstackBuffer [4]uint16
	for i := uint32(3); i > 0; i-- {
		address := scaledDS + uint32(c.RSP) - i
		rstackBuffer[i] = c.ReadDataMemory(address)
	}
	rstackBuffer[0] = c.RTOS

	// Create a bunch of short cut names for uuse witth the  disasembbly
	leftOperand := c.ReadDataMemory(scaledDS + uint32(c.PSP) - 1)
	rightOperand := c.PTOS
	inlineOperand := c.ReadDataMemory(scaledCS + uint32(c.PC))

	snapShot.absoluteAddress = absoluteAddress
	snapShot.cpuStruct = *c
	snapShot.pStack = pstackBuffer
	snapShot.rStack = rstackBuffer
	snapShot.opCode = opCode
	snapShot.leftOperand = leftOperand
	snapShot.rightOperand = rightOperand
	snapShot.inlineOperand = inlineOperand
	snapShot.rtosOperand = c.RTOS
	snapShot.pspOperand = c.PSP
	snapShot.rspOperand = c.RSP
	snapShot.csOperand = c.CS
	snapShot.dsOperand = c.DS
	snapShot.esOperand = c.ES

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

		destinationAddress := c.consumeInstructionLiteral()
		c.PC = destinationAddress

		return Normal
	}

	// f JMPF dst
	if opCode == branchFalseOpcode {
		History.logInstruction(snapShot)

		flag := c.pop()
		destinationAddress := c.consumeInstructionLiteral()
		if flag == cpuFalse {
			c.PC = destinationAddress
		}

		return Normal
	}

	// CS_FETCH
	if opCode == csFetchOpcode {
		History.logInstruction(snapShot)

		c.push(c.CS)

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

	// DS_FETCH
	if opCode == dsFetchOpcode {
		History.logInstruction(snapShot)

		c.push(c.DS)

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
			c.push(cpuTrue)
		} else {
			c.push(cpuFalse)
		}

		return Normal
	}

	// EI
	if opCode == eiOpcode {
		History.logInstruction(snapShot)

		c.IntCtlLow |= 0x0001

		return Normal
	}

	// ES_FETCH
	if opCode == esFetchOpcode {
		History.logInstruction(snapShot)

		c.push(c.ES)

		return Normal
	}

	// d FETCH
	if opCode == fetchOpcode {
		History.logInstruction(snapShot)

		destinationAddress := uint32(c.pop()) + scaledDS
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

	// K_SP_STORE
	if opCode == kSpStoreOpcode {
		History.logInstruction(snapShot)

		c.DS = 0x0000
		c.PSP = c.PTOS

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
		tmpRTOS := c.RTOS
		c.rPush(c.DS)
		c.rPush(c.CS)
		c.rPush(c.ES)
		c.rPush(c.PSP)
		c.rPush(c.PTOS)
		c.rPush(c.PC)
		c.rPush(uint16(c.IntCtlLow))
		c.rPush(tmpRSP)
		c.rPush(tmpRTOS)

		c.IntCtlLow = c.IntCtlLow & 0xFE
		c.PC = 0xFD00
		c.CS = 0x0000

		return Normal
	}

	// RETI
	// rPop sequence should match rPush sequene in JSRINT
	// Sequence from TOP down is RTOS, RSP, Flags, PC, PTOS, PSP, ES, CS, DS
	if opCode == retiOpcode {
		History.logInstruction(snapShot)

		tmpRTOS := c.rPop()
		tmpRSP := c.rPop()
		c.IntCtlLow = uint8(c.rPop())
		c.PC = c.rPop()
		c.PTOS = c.rPop()
		c.PSP = c.rPop()
		c.ES = c.rPop()
		c.CS = c.rPop()
		c.DS = c.rPop()
		c.RSP = tmpRSP
		c.RTOS = tmpRTOS

		fmt.Printf("DEBUG in RETI 9 values were popped:\n")
		c.ShowStatus()

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
			c.push(cpuTrue)
		} else {
			c.push(cpuFalse)
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

	// d LONG_FETCH
	if opCode == longFetchOpcode {
		History.logInstruction(snapShot)

		destinationAddress := uint32(c.pop()) + scaledES
		v := c.ReadDataMemory(destinationAddress)
		c.push(v)

		return Normal
	}

	// val addr LONG_STORE
	if opCode == longStoreOpcode {
		History.logInstruction(snapShot)

		unscaledAddress := uint32(c.pop())
		destinationAddress := unscaledAddress + scaledES
		val := c.pop()
		c.WriteDataMemory(destinationAddress, val)

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
			c.push(cpuTrue)
		} else {
			c.push(cpuFalse)
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

	// address PLUS_PLUS
	if opCode == plusPlusOpcode {
		History.logInstruction(snapShot)

		address := c.pop()
		c.push(address)
		value := c.ReadDataMemory(uint32(address) + scaledDS)
		value++
		c.WriteDataMemory(uint32(address)+scaledDS, value)

		return Normal
	}

	// POPF
	// Restore the flags register
	if opCode == popFOpcode {
		History.logInstruction(snapShot)

		flags := c.pop()
		c.IntCtlLow = uint8(flags)

		return Normal
	}

	// PUSHF
	// Save the flags register
	if opCode == pushFOpcode {
		History.logInstruction(snapShot)

		c.push(uint16(c.IntCtlLow))

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

	// RP_FETCH
	if opCode == rpFetchOpcode {
		History.logInstruction(snapShot)

		c.push(c.RSP)

		return Normal
	}

	// RP_STORE
	if opCode == rpStoreOpcode {
		History.logInstruction(snapShot)

		c.RSP = c.pop()

		return Normal
	}

	// SLL
	if opCode == sllOpcode {
		History.logInstruction(snapShot)

		c.PTOS = c.PTOS << 1

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

	// SRA
	if opCode == sraOpcode {
		History.logInstruction(snapShot)

		signBit := c.PTOS & 0x8000
		c.PTOS = signBit | (c.PTOS >> 1)

		return Normal
	}

	// SRL
	if opCode == srlOpcode {
		History.logInstruction(snapShot)

		c.PTOS = c.PTOS >> 1

		return Normal
	}

	// val addr STORE
	if opCode == storeOpcode {
		History.logInstruction(snapShot)

		destinationAddress := uint32(c.pop()) + scaledDS
		val := c.pop()
		c.WriteDataMemory(destinationAddress, val)

		return Normal
	}

	// addr val STORE2
	if opCode == store2Opcode {
		History.logInstruction(snapShot)

		val := c.pop()
		destinationAddress := uint32(c.pop()) + scaledDS
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

	// SYSCALL
	// rPush sequence should match rPop sequene in RETI
	if opCode == sysCallOpcode {
		History.logInstruction(snapShot)

		tmpRSP := c.RSP
		tmpRTOS := c.RTOS
		c.rPush(c.DS)
		c.rPush(c.CS)
		c.rPush(c.ES)
		c.rPush(c.PSP)
		c.rPush(c.PTOS)
		c.rPush(c.PC)
		c.rPush(uint16(c.IntCtlLow))
		c.rPush(tmpRSP)
		c.rPush(tmpRTOS)

		c.IntCtlLow = c.IntCtlLow & 0xFE
		c.PC = 0xFD02
		c.CS = 0x0000

		return Normal
	}

	// a TO_DS
	if opCode == toDSOpcode {
		History.logInstruction(snapShot)

		c.DS = c.pop()

		return Normal
	}

	// a TO_ES
	if opCode == toESOpcode {
		History.logInstruction(snapShot)

		c.ES = c.pop()

		return Normal
	}

	// a TO_R
	if opCode == toROpcode {
		History.logInstruction(snapShot)

		a := c.pop()
		c.rPush(a)

		return Normal
	}

	// UM+
	if opCode == umPlusOpcode {
		a := uint32(c.pop())
		b := uint32(c.pop())
		sum := (a + b) & 0xFFFF
		c.push(uint16(sum))
		carry := ((a + b) & 0x10000) >> 16
		c.push(uint16(carry))
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
	runtime.Goexit()
	return 0 // Will never be reached
}
