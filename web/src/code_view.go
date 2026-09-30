package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/unixpickle/mips32"
)

const PreviewLineCount = 7

type CodeView struct {
	element *js.Object
}

func NewCodeView() *CodeView {
	return &CodeView{
		element: js.Global.Get("debugger-code-view"),
	}
}

func (c *CodeView) Update(e *mips32.Emulator) {
	var startAddress uint32
	if (e.ProgramCounter / 4) > (PreviewLineCount / 2) {
		startAddress = e.ProgramCounter - (PreviewLineCount/2)*4
	}
	c.element.Set("innerHTML", "<tr><td>Addr</td><td>Assembly</td><td>Code</td></tr>")
	for i := 0; i < PreviewLineCount; i++ {
		addr := startAddress + uint32(i*4)
		row := createCodeViewLine(e, addr)
		if addr == e.ProgramCounter {
			row.Set("className", row.Get("className").String()+" debugger-code-view-current")
		}
		c.element.Call("appendChild", row)
	}
}

func createCodeViewLine(e *mips32.Emulator, addr uint32) *js.Object {
	document := js.Global.Get("document")
	row := document.Call("createElement", "tr")

	addrColumn := document.Call("createElement", "td")
	addrColumn.Set("textContent", format32BitHex(addr))
	addrColumn.Set("className", "debugger-code-view-addr")
	row.Call("appendChild", addrColumn)

	codeColumn := document.Call("createElement", "td")
	codeColumn.Set("className", "debugger-code-view-code")
	if inst := e.Executable.Get(addr); inst != nil {
		rendering, err := inst.Render()
		if err != nil {
			codeColumn.Set("textContent", "(Unknown)")
		} else {
			codeColumn.Set("textContent", rendering.String())
		}
	} else {
		codeColumn.Set("textContent", "NOP")
	}
	row.Call("appendChild", codeColumn)

	opcodeColumn := document.Call("createElement", "td")
	opcodeColumn.Set("className", "debugger-code-view-opcode")
	if inst := e.Executable.Get(addr); inst != nil {
		opcode, err := inst.Encode(addr, e.Executable.Symbols)
		if err != nil {
			opcodeColumn.Set("textContent", "(invalid)")
		} else {
			opcodeColumn.Set("textContent", format32BitHex(opcode))
		}
	} else {
		opcodeColumn.Set("textContent", format32BitHex(0))
	}
	row.Call("appendChild", opcodeColumn)

	return row
}
