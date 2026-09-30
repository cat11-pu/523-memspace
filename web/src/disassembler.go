package main

import (
	"encoding/hex"
	"errors"
	"strconv"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/unixpickle/mips32"
)

type Disassembler struct {
	textarea  *js.Object
	errorView *js.Object
}

func NewDisassembler() *Disassembler {
	res := &Disassembler{
		textarea:  js.Global.Get("disassembler-data"),
		errorView: js.Global.Get("disassembler-error"),
	}
	js.Global.Get("disassembler-button").Call("addEventListener", "click", res.disassemble)
	return res
}

func (d *Disassembler) SetData(data []uint32) {
	spacedValues := make([]string, len(data))
	for i, d := range data {
		numStr := strconv.FormatUint(uint64(d), 16)
		for len(numStr) < 8 {
			numStr = "0" + numStr
		}
		spacedValues[i] = numStr
	}

	d.textarea.Set("value", strings.Join(spacedValues, " "))
}

func (d *Disassembler) disassemble() {
	text := d.textarea.Get("value").String()
	text = strings.Replace(text, " ", "", -1)
	text = strings.Replace(text, "\n", "", -1)
	text = strings.Replace(text, "\t", "", -1)

	data, err := hex.DecodeString(text)
	if err != nil {
		d.showError(err)
		return
	} else if len(data)%4 != 0 {
		d.showError(errors.New("binary data's size must be a multiple of 32-bits"))
		return
	}

	instructions := make([]*mips32.Instruction, len(data)/4)
	instStrs := make([]string, len(instructions))
	for i := 0; i < len(data); i += 4 {
		word := (uint32(data[i+0]) << 24) | (uint32(data[i+1]) << 16) | (uint32(data[i+2]) << 8) |
			uint32(data[i+3])
		instructions[i/4] = mips32.DecodeInstruction(word)
		rendering, err := instructions[i/4].Render()
		if err != nil {
			d.showError(err)
			return
		}
		instStrs[i/4] = rendering.String()
	}

	d.hideError()
	GlobalAssembler.SetCode(strings.Join(instStrs, "\n"))
	GlobalAssembler.Show()
}

func (d *Disassembler) hideError() {
	d.errorView.Set("className", "error-view")
}

func (d *Disassembler) showError(err error) {
	d.errorView.Set("className", "error-view showing-error")
	d.errorView.Set("textContent", err.Error())
}
