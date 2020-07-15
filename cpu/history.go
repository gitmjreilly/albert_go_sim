package cpu

import (
	"albert_go_sim/intmaxmin"
	"fmt"
)

// Status contains a snapshot along with disassembly
type Status struct {
	cpuStruct       CPU
	absoluteAddress uint32
	opCode          uint16
	pStack          [4]uint16
	rStack          [4]uint16
	leftOperand     uint16
	rightOperand    uint16
	inlineOperand   uint16
	rtosOperand     uint16
	pspOperand      uint16
	rspOperand      uint16
	ptosOperand     uint16
	csOperand       uint16
	dsOperand       uint16
	esOperand       uint16
	flagsOperand    uint8
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

// Clear wipes out history
func (h *tHistory) Clear() {
	h.numEntries = 0
	h.nextIn = 0
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
	rspOperand := s.rspOperand
	rtosOperand := s.rtosOperand
	csOperand := s.csOperand
	dsOperand := s.dsOperand
	esOperand := s.esOperand
	flagsOperand := s.flagsOperand

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
		instructionString := fmt.Sprintf("[%04X] JMPF %04X", rightOperand, inlineOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// CS_FETCH
	if opCode == csOperand {
		instructionString := fmt.Sprintf("[%04X] CS_FETCH", csOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// DI
	if opCode == diOpcode {
		instructionString := fmt.Sprintf("DI")
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// DOLIT l
	if opCode == doLitOpcode {
		instructionString := fmt.Sprintf("DO_LIT %04X", inlineOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// DS_FETCH
	if opCode == dsOperand {
		instructionString := fmt.Sprintf("[%04X] DS_FETCH", dsOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// ES_FETCH
	if opCode == esOperand {
		instructionString := fmt.Sprintf("[%04X] ES_FETCH", esOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// a DROP
	if opCode == dropOpcode {
		instructionString := fmt.Sprintf("[%04X] DROP", rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
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
		instructionString := fmt.Sprintf("[%04X %04X] EQUAL", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// EI
	if opCode == eiOpcode {
		instructionString := fmt.Sprintf("EI")
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
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
		instructionString := fmt.Sprintf("[RTOS: %04X] FROM_R", rtosOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

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
		instructionString := fmt.Sprintf("RETI")
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// HALT
	if opCode == haltOpcode {
		instructionString := fmt.Sprintf("HALT")
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// K_SP_STORE
	if opCode == kSpStoreOpcode {
		instructionString := fmt.Sprintf("[PTOS: %04X] K_SP_STORE", rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// a b LESS
	if opCode == lessOpcode || opCode == sLessOpcode {
		instructionString := fmt.Sprintf("[%04X %04X] LESS", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// d LONG_FETCH
	if opCode == longFetchOpcode {
		instructionString := fmt.Sprintf("[%04X:%04X] LONG_FETCH", esOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// val addr LONG_STORE
	if opCode == longStoreOpcode {
		instructionString := fmt.Sprintf("[%04X %04X:%04X] LONG_STORE", leftOperand, esOperand, rightOperand)
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
		instructionString := fmt.Sprintf("[%04X %04X] MUL", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// a NEG?
	if opCode == negOpcode {
		instructionString := fmt.Sprintf("[%04X] NEG?", rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// NOP
	if opCode == nopOpcode {
		instructionString := fmt.Sprintf("NOP")
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// a b OR
	if opCode == orOpcode {
		instructionString := fmt.Sprintf("[%04X %04X] OR", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// a b +
	if opCode == plusOpcode {
		instructionString := fmt.Sprintf("[%04X %04X] PLUS", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// POPF
	if opCode == popFOpcode {
		instructionString := fmt.Sprintf("[PTOS: %04X] POPF", rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString

	}

	// PUSHF
	if opCode == pushFOpcode {
		instructionString := fmt.Sprintf("[Flags: %02X] PUSHF", flagsOperand)
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
		instructionString := fmt.Sprintf("[RTOS: %04X] R_FETCH", rtosOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// RP_FETCH
	if opCode == rpFetchOpcode {
		instructionString := fmt.Sprintf("[RSP: %04X] RP_FETCH (to be fixed)", rspOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString

	}

	// RP_STORE
	if opCode == rpStoreOpcode {
		instructionString := fmt.Sprintf("[PTOS: %04X] RP_STORE", rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString

	}

	// SP_FETCH
	if opCode == spFetchOpcode {
		instructionString := fmt.Sprintf("[PSP: %04X] SP_FETCH", pspOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// SP_STORE
	if opCode == spStoreOpcode {
		instructionString := fmt.Sprintf("[PTOS: %04X] SP_STORE", rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// val addr STORE
	if opCode == storeOpcode {
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
		instructionString := fmt.Sprintf("[%04X %04X] MINUS", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// a b SWAP
	if opCode == swapOpcode {
		instructionString := fmt.Sprintf("[%04X %04X] SWAP", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// SYSCALL
	if opCode == sysCallOpcode {
		instructionString := fmt.Sprintf("SYSCALL")
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	// a TO_DS
	if opCode == toDSOpcode {
		instructionString := fmt.Sprintf("[PTOS: %04X] TO_DS", rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// a TO_ES
	if opCode == toESOpcode {
		instructionString := fmt.Sprintf("[PTOS: %04X] TO_ES", rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// a TO_R
	if opCode == toROpcode {
		instructionString := fmt.Sprintf("[PTOS: %04X] TO_R", rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)
		return disassemblyString
	}

	// a b XOR
	if opCode == xorOpcode {
		instructionString := fmt.Sprintf("[%04X %04X] XOR", leftOperand, rightOperand)
		disassemblyString := fmt.Sprintf("%08X  %25s | %s", absoluteAddress, instructionString, stackString)

		return disassemblyString
	}

	fmt.Printf("Unknown opcode [%04X] address [%08X]\n", opCode, absoluteAddress)
	return "Unknown Opcode"
}
