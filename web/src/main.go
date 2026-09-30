package main

import "github.com/gopherjs/gopherjs/js"

var GlobalDebugger *Debugger
var GlobalAssembler *Assembler
var GlobalDisassembler *Disassembler

var defaultProgram = `# Put your code here, then hit Assemble.
# Large programs may take a moment or two to assemble.

# Set $r1 to 0xDEADBEEF.
LUI $r1, 0xDEAD
ORI $r1, $r1, 0xBEEF`

func main() {
	js.Global.Get("window").Call("addEventListener", "load", func() {
		go func() {
			GlobalDebugger = NewDebugger()
			GlobalAssembler = NewAssembler()
			GlobalDisassembler = NewDisassembler()
			GlobalAssembler.SetCode(defaultProgram)
			GlobalAssembler.Assemble()
		}()
	})
}
