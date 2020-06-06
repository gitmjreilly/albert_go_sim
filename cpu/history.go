package cpu

import (
	"albert_go_sim/intmaxmin"
)

// Status contains a snapshot along with disassembly
type Status struct {
	cpuStruct       CPU
	absoluteAddress uint16
	opCode          uint16
	pStack          [4]uint16
	rStack          [4]uint16
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

}
