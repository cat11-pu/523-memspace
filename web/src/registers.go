package main

import (
	"strconv"

	"github.com/gopherjs/gopherjs/js"
	"github.com/unixpickle/mips32"
)

type Registers struct {
	regCells [32]*js.Object
	callback func(reg int, val uint32)
}

func NewRegisters() *Registers {
	res := &Registers{}
	regTable := js.Global.Get("debugger-registers")
	document := js.Global.Get("document")
	for i := 0; i < 16; i++ {
		row := document.Call("createElement", "tr")
		tds := make([]*js.Object, 4)
		for j := range tds {
			tds[j] = document.Call("createElement", "td")
			if j%2 == 0 {
				tds[j].Set("className", "register-name")
			} else {
				tds[j].Set("className", "register-value")
			}
			row.Call("appendChild", tds[j])
		}
		tds[0].Set("textContent", "$r"+strconv.Itoa(i))
		tds[2].Set("textContent", "$r"+strconv.Itoa(i+16))
		tds[1].Set("textContent", format32BitHex(0))
		tds[3].Set("textContent", format32BitHex(0))
		res.regCells[i] = tds[1]
		res.regCells[i+16] = tds[3]
		regTable.Call("appendChild", row)
	}
	for i := 0; i < 32; i++ {
		func(num int) {
			res.regCells[num].Call("addEventListener", "click", func() {
				res.editRegister(num)
			})
		}(i)
	}
	return res
}

func (r *Registers) Update(file mips32.RegisterFile) {
	for i := 0; i < 32; i++ {
		r.regCells[i].Set("textContent", format32BitHex(file[i]))
	}
}

func (r *Registers) SetCallback(f func(reg int, val uint32)) {
	r.callback = f
}

func (r *Registers) editRegister(i int) {
	NewEntryPopup("Enter value for $r"+strconv.Itoa(i), func(v uint32) {
		r.regCells[i].Set("textContent", format32BitHex(v))
		if r.callback != nil {
			r.callback(i, v)
		}
	})
}
