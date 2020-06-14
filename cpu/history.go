package cpu

import (
	"albert_go_sim/intmaxmin"
	"fmt"
)

// Status contains a snapshot along with disassembly
type Status struct {
	cpuStruct       CPU
	absoluteAddress uint16
	opCode          uint16
	pStack          [4]uint16
	rStack          [4]uint16
	leftOperand     uint16
	rightOperand    uint16
	inlineOperand   uint16
	rtosOperand     uint16
	pspOperand      uint16
	ptosOperand     uint16
	// disassemblyString string
}

const historySize = 1024 * 1024

type tHistory struct {
	data       [historySize]Status
	numEntries int
	nextIn     int
}

func (h *tHistory) logInstruction(s Status) {
	h.data[h.nextIn] = s
	h.nextIn = (h.nextIn + 1) % historySize
	if h.numEntries < historySize {
		h.numEntries++
	}
}

// Display dumps numInstructions of the cpu history
func (h *tHistory) Display(numInstructions int) {
	numInstructions = intmaxmin.Constrain(numInstructions, 0, h.numEntries)

	start := h.nextIn - numInstructions
	if start < 0 {
		start = len(h.data) - start
	}

	index := start
	for i := 0; i < numInstructions; i++ {

		// absoluteAddress := h.data[index].absoluteAddress
		// disassemblyString := h.data[index].disassemblyString

		// psp := h.data[index].cpuStruct.PSP
		// rsp := h.data[index].cpuStruct.RSP

		// p0 := h.data[index].pStack[0]
		// p1 := h.data[index].pStack[1]
		// p2 := h.data[index].pStack[2]

		// r0 := h.data[index].rStack[0]
		// r1 := h.data[index].rStack[1]
		// r2 := h.data[index].rStack[2]

		// s := fmt.Sprintf("%08X  %25s PSP:%04X   PSTACK[%04X %04X %04X]  RSP:%04X  RSTACK[%04X %04X %04X]",
		// 	absoluteAddress, disassemblyString, psp, p0, p1, p2,
		// 	rsp, r0, r1, r2)

		disassemblyString := createDisassemblyString(h.data[index])
		fmt.Println(disassemblyString)
		index++
		index = index % len(h.data)
	}

}

// createDisassemblyString
func createDisassemblyString(s Status) string {
	// var disassemblyString string

	var pstackBuffer [4]uint16 = s.pStack
	var rstackBuffer [4]uint16 = s.rStack

	stackString := fmt.Sprintf("PSTACK => %04X %04X %04X PTOS:%04X    RSTACK => %04X %04X %04X RTOS:%04X",
		pstackBuffer[3], pstackBuffer[2], pstackBuffer[1], pstackBuffer[0],
		rstackBuffer[3], rstackBuffer[2], rstackBuffer[1], rstackBuffer[0])

	// Create a bunch of short cut names for use with the  disasembly
	absoluteAddress := s.absoluteAddress
	opCode := s.opCode
	leftOperand := s.leftOperand
	rightOperand := s.rightOperand
	inlineOperand := s.inlineOperand
	pspOperand := s.pspOperand
	rtosOperand := s.rtosOperand

	// a b AND
	if opCode == andOpcode {
		instructionString := fmt.Sprintf("[%04X %04X] AND", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// BRA dst
	if opCode == branchOpcode {
		instructionString := fmt.Sprintf("BRA %04X", inlineOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// f JMPF dst
	if opCode == branchFalseOpcode {
		// disassemblyString := fmt.Sprintf("%08X  [%04X] JMPF %04X | %s", absoluteAddress, rightOperand, inlineOperand, stackString)

		instructionString := fmt.Sprintf("[%04X] JMPF %04X", rightOperand, inlineOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// DI
	if opCode == diOpcode {
		disassemblyString := fmt.Sprintf("%08X  DI | %s", absoluteAddress, stackString)
		return disassemblyString
	}

	// DOLIT l
	if opCode == doLitOpcode {
		instructionString := fmt.Sprintf("DO_LIT %04X", inlineOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// a DROP
	if opCode == dropOpcode {
		disassemblyString := fmt.Sprintf("%08X  [%04X] DROP | %s", absoluteAddress, rightOperand, stackString)
		return disassemblyString
	}

	// a DUP
	if opCode == dupOpcode {
		instructionString := fmt.Sprintf("[%04X] DUP", rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// a b EQUAL
	if opCode == equallOpcode {
		// disassemblyString := fmt.Sprintf("%08X  [%04X %04X] EQUAL | %s", absoluteAddress, leftOperand, rightOperand, stackString)
		instructionString := fmt.Sprintf("[%04X %04X] EQUAL", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// EI
	if opCode == eiOpcode {
		disassemblyString := fmt.Sprintf("%08X  EI | %s", absoluteAddress, stackString)
		return disassemblyString
	}

	// d FETCH
	if opCode == fetchOpcode {
		instructionString := fmt.Sprintf("[%04X] FETCH", rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// (RTOS) FROM_R
	if opCode == fromROpcode {
		disassemblyString := fmt.Sprintf("%08X  [RTOS: %04X] FROM_R | %s", absoluteAddress, rtosOperand, stackString)
		return disassemblyString
	}

	// JSR d
	if opCode == jsrOpcode {
		instructionString := fmt.Sprintf("JSR %04X", inlineOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// JSRINT
	if opCode == jsrintOpcode {
		disassemblyString := fmt.Sprintf("JSRINT | %s", stackString)
		return disassemblyString
	}

	// RETI
	if opCode == retiOpcode {
		disassemblyString := fmt.Sprintf("%08X  RETI | %s", absoluteAddress, stackString)
		return disassemblyString
	}

	// HALT
	if opCode == haltOpcode {
		disassemblyString := fmt.Sprintf("%08X  HALT | %s", absoluteAddress, stackString)
		return disassemblyString
	}

	// a b LESS
	if opCode == lessOpcode || opCode == sLessOpcode {
		// disassemblyString := fmt.Sprintf("%08X  [%04X %04X] LESS | %s", absoluteAddress, leftOperand, rightOperand, stackString)

		instructionString := fmt.Sprintf("[%04X %04X] LESS", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// L_VAR n
	if opCode == lvarOpcode {
		instructionString := fmt.Sprintf("L_VAR %04X", inlineOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString

	}

	// a b *
	if opCode == mulOpcode {
		// disassemblyString := fmt.Sprintf("%08X  [%04X %04X] MUL | %s", absoluteAddress, leftOperand, rightOperand, stackString)

		instructionString := fmt.Sprintf("[%04X %04X] MUL", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// a NEG?
	if opCode == negOpcode {
		disassemblyString := fmt.Sprintf("%08X  [%04X] NEG? | %s", absoluteAddress, rightOperand, stackString)
		return disassemblyString
	}

	// NOP
	if opCode == nopOpcode {
		disassemblyString := fmt.Sprintf("%08X  NOP | %s", absoluteAddress, stackString)
		return disassemblyString
	}

	// a b OR
	if opCode == orOpcode {
		// disassemblyString := fmt.Sprintf("%08X  [%04X %04X] OR | %s", absoluteAddress, leftOperand, rightOperand, stackString)

		instructionString := fmt.Sprintf("[%04X %04X] OR", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// a b +
	if opCode == plusOpcode {
		// disassemblyString := fmt.Sprintf("%08X  [%04X %04X] PLUS | %s", absoluteAddress, leftOperand, rightOperand, stackString)

		instructionString := fmt.Sprintf("[%04X %04X] PLUS", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// RET
	if opCode == retOpcode {
		instructionString := fmt.Sprintf("RET [%04X]", rtosOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// R_FETCH
	if opCode == rFetchOpcode {
		disassemblyString := fmt.Sprintf("%08X  [RTOS: %04X] R_FETCH | %s", absoluteAddress, rtosOperand, stackString)
		return disassemblyString
	}

	// SP_FETCH
	if opCode == spFetchOpcode {
		disassemblyString := fmt.Sprintf("%08X  [PSP: %04X] SP_FETCH | %s", absoluteAddress, pspOperand, stackString)
		return disassemblyString
	}

	// SP_STORE
	if opCode == spStoreOpcode {
		disassemblyString := fmt.Sprintf("%08X  [PTOS: %04X] SP_STORE | %s", absoluteAddress, rightOperand, stackString)
		return disassemblyString
	}

	// val addr STORE
	if opCode == storeOpcode {
		// disassemblyString := fmt.Sprintf("%08X  [%04X %04X] STORE | %s", absoluteAddress, leftOperand, rightOperand, stackString)

		instructionString := fmt.Sprintf("[%04X %04X] STORE", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// addr val STORE2
	if opCode == store2Opcode {
		instructionString := fmt.Sprintf("[%04X %04X] STORE2", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// a b -
	if opCode == subOpcode {
		// disassemblyString := fmt.Sprintf("%08X  [%04X %04X] SUB | %s", absoluteAddress, leftOperand, rightOperand, stackString)

		instructionString := fmt.Sprintf("[%04X %04X] MINUS", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString

	}

	// a b SWAP
	if opCode == swapOpcode {
		// disassemblyString := fmt.Sprintf("%08X  [%04X %04X] SWAP | %s", absoluteAddress, leftOperand, rightOperand, stackString)

		instructionString := fmt.Sprintf("[%04X %04X] SWAP", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// a TO_R
	if opCode == toROpcode {
		disassemblyString := fmt.Sprintf("%08X  [PTOS: %04X] TO_R | %s", absoluteAddress, rightOperand, stackString)
		return disassemblyString

	}

	// a b XOR
	if opCode == xorOpcode {
		// disassemblyString := fmt.Sprintf("%08X  [%04X %04X] XOR | %s", absoluteAddress, leftOperand, rightOperand, stackString)

		instructionString := fmt.Sprintf("[%04X %04X] XOR", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	fmt.Printf("Unknown opcode [%04X] address [%08X]\n", opCode, absoluteAddress)
	return "Unknown Opcode"
}
