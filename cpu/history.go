package cpu

import (
	"albert_go_sim/intmaxmin"
	"fmt"
)

// Status contains a snapshot along with disassembly
type Status struct {
	cpuStruct         CPU
	absoluteAddress   uint16
	opCode            uint16
	pStack            [4]uint16
	rStack            [4]uint16
	leftOperand       uint16
	rightOperand      uint16
	inlineOperand     uint16
	rtosOperand       uint16
	disassemblyString string
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

		absoluteAddress := h.data[index].absoluteAddress
		disassemblyString := h.data[index].disassemblyString

		psp := h.data[index].cpuStruct.PSP
		rsp := h.data[index].cpuStruct.RSP

		p0 := h.data[index].pStack[0]
		p1 := h.data[index].pStack[1]
		p2 := h.data[index].pStack[2]

		r0 := h.data[index].rStack[0]
		r1 := h.data[index].rStack[1]
		r2 := h.data[index].rStack[2]

		s := fmt.Sprintf("%08X  %25s PSP:%04X   PSTACK[%04X %04X %04X]  RSP:%04X  RSTACK[%04X %04X %04X]",
			absoluteAddress, disassemblyString, psp, p0, p1, p2,
			rsp, r0, r1, r2)
		fmt.Println(s)
		index++
		index = index % len(h.data)
	}

}
