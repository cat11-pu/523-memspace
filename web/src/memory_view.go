package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/unixpickle/mips32"
)

const memoryViewRows = 6
const memoryViewColumns = 8

type MemoryView struct {
	baseLabel    *js.Object
	baseAddress  uint32
	memoryCells  []*js.Object
	addressCells []*js.Object

	memory mips32.Memory
}

func NewMemoryView() *MemoryView {
	res := &MemoryView{
		baseLabel:    js.Global.Get("debugger-memory-base"),
		memoryCells:  []*js.Object{},
		addressCells: []*js.Object{},
	}

	document := js.Global.Get("document")
	table := js.Global.Get("debugger-memory-contents")
	for i := 0; i < memoryViewRows; i++ {
		row := document.Call("createElement", "tr")
		row.Set("className", "debugger-memory-row")
		addressCell := document.Call("createElement", "td")
		addressCell.Set("className", "debugger-memory-address")
		addressCell.Set("textContent", format32BitHex(0))
		row.Call("appendChild", addressCell)

		for j := 0; j < memoryViewColumns; j++ {
			valueCell := document.Call("createElement", "td")
			valueCell.Set("className", "debugger-memory-value")
			valueCell.Set("textContent", "00")
			row.Call("appendChild", valueCell)
			res.memoryCells = append(res.memoryCells, valueCell)

			offset := i*memoryViewColumns + j
			valueCell.Call("addEventListener", "click", func() {
				res.clickedCell(offset)
			})
		}
		res.addressCells = append(res.addressCells, addressCell)
		table.Call("appendChild", row)
	}

	res.baseLabel.Call("addEventListener", "click", func() {
		NewEntryPopup("View memory at address", func(num uint32) {
			go res.updateBase(num)
		})
	})

	res.updateAddressCells()
	return res
}

func (m *MemoryView) Update(mem mips32.Memory) {
	m.memory = mem

	addr := m.baseAddress
	idx := 0
	for i := 0; i < memoryViewRows; i++ {
		for j := 0; j < memoryViewColumns; j++ {
			val := m.memory.Get(addr)
			m.memoryCells[idx].Set("textContent", format8BitHex(uint8(val)))
			addr += 1
			idx += 1
		}
	}
}

func (m *MemoryView) clickedCell(index int) {
	addr := m.baseAddress + uint32(index)
	NewEntryPopup("Set value at "+format32BitHex(addr), func(val uint32) {
		go func() {
			m.memory.Set(addr, byte(val))
			m.Update(m.memory)
		}()
	})
}

func (m *MemoryView) updateBase(addr uint32) {
	alignment := uint32(memoryViewColumns)
	for (addr % alignment) != 0 {
		addr--
	}

	// Make sure they don't view past the end of memory.
	for addr+(memoryViewRows*memoryViewColumns) < addr {
		addr -= alignment
	}

	m.baseAddress = addr
	m.baseLabel.Set("textContent", format32BitHex(addr))
	m.updateAddressCells()
	m.Update(m.memory)
}

func (m *MemoryView) updateAddressCells() {
	addr := m.baseAddress
	for i := 0; i < memoryViewRows; i++ {
		m.addressCells[i].Set("textContent", format32BitHex(addr))
		addr += memoryViewColumns
	}
}
