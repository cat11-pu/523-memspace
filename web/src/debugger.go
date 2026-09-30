package main

import (
	"strconv"
	"sync"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/unixpickle/mips32"
)

type Debugger struct {
	lock sync.Mutex

	frequency   int
	controlChan chan debuggerCommand
	emulator    *mips32.Emulator
	stepCount   int

	registers      *Registers
	codeView       *CodeView
	memoryView     *MemoryView
	errorView      *js.Object
	stepCountLabel *js.Object
}

func NewDebugger() *Debugger {
	res := &Debugger{
		frequency:   4,
		controlChan: make(chan debuggerCommand, 0),
		emulator: &mips32.Emulator{
			Memory: mips32.NewLazyMemory(),
			Executable: &mips32.Executable{
				Segments: map[uint32][]mips32.Instruction{},
				Symbols:  map[string]uint32{},
			},
			LittleEndian: true,
		},
		registers:      NewRegisters(),
		codeView:       NewCodeView(),
		memoryView:     NewMemoryView(),
		errorView:      js.Global.Get("debugger-error"),
		stepCountLabel: js.Global.Get("debugger-step-count"),
	}

	go res.debugLoop()

	res.registers.SetCallback(func(reg int, val uint32) {
		res.lock.Lock()
		defer res.lock.Unlock()
		res.emulator.RegisterFile[reg] = val
	})

	res.registerUIEvents()
	res.updateUI()

	return res
}

// SetExecutable loads a new executable in the debugger.
// It deletes all saved state (i.e. registers and memory).
// If the given executable is nil, the current executable will be reloaded.
func (d *Debugger) SetExecutable(e *mips32.Executable) {
	d.hideError()
	d.controlChan <- stopDebugger
	d.lock.Lock()
	if e == nil {
		e = d.emulator.Executable
	}
	d.emulator = &mips32.Emulator{
		Memory:       mips32.NewLazyMemory(),
		Executable:   e,
		LittleEndian: true,
	}
	d.stepCount = 0
	d.lock.Unlock()
	d.updateUI()
}

func (d *Debugger) Show() {
	js.Global.Get("location").Set("hash", "#debugger")
}

// Get returns the byte at a given memory address in the debugger's RAM.
func (d *Debugger) Get(ptr uint32) byte {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.emulator.Memory.Get(ptr)
}

// Set changes the byte at a given memory address in the debugger's RAM.
func (d *Debugger) Set(ptr uint32, b byte) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.emulator.Memory.Set(ptr, b)
}

func (d *Debugger) debugLoop() {
	for command := range d.controlChan {
		if command == stepDebugger {
			d.stepDebugger()
		} else if command == startDebugger {
			d.runDebugger()
		}
	}
}

func (d *Debugger) runDebugger() {
	frameDuration, cyclesPerFrame := d.tickInfo()
	ticker := time.NewTicker(frameDuration)
	defer func() {
		ticker.Stop()
	}()
	defer d.updateButtonState(false)
	d.updateButtonState(true)
	d.hideError()
	for {
		select {
		case <-ticker.C:
		case msg := <-d.controlChan:
			if msg == stopDebugger {
				return
			} else if msg == updateDebuggerFreq {
				ticker.Stop()
				frameDuration, cyclesPerFrame = d.tickInfo()
				ticker = time.NewTicker(frameDuration)
			}
		}
		d.lock.Lock()
		for i := 0; i < cyclesPerFrame; i++ {
			err := d.emulator.Step()
			d.stepCount++
			if err != nil {
				d.lock.Unlock()
				d.updateUI()
				d.handleError(err)
				return
			}
			if d.emulator.Done() {
				d.lock.Unlock()
				d.updateUI()
				return
			}
		}
		d.lock.Unlock()
		d.updateUI()
	}
}

func (d *Debugger) stepDebugger() {
	d.lock.Lock()
	err := d.emulator.Step()
	d.stepCount++
	d.lock.Unlock()
	if err != nil {
		d.handleError(err)
	} else {
		d.hideError()
		d.updateUI()
	}
}

func (d *Debugger) handleError(err error) {
	d.updateButtonState(false)
	d.lock.Lock()
	defer d.lock.Unlock()
	d.errorView.Set("textContent", err.Error())
	d.errorView.Set("className", "error-view showing-error")
}

func (d *Debugger) hideError() {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.errorView.Set("className", "error-view")
}

func (d *Debugger) updateUI() {
	d.memoryView.Update(d)

	d.lock.Lock()
	defer d.lock.Unlock()

	d.registers.Update(d.emulator.RegisterFile)
	d.codeView.Update(d.emulator)
	d.stepCountLabel.Set("textContent", "Steps: "+strconv.Itoa(d.stepCount))
}

func (d *Debugger) updateButtonState(running bool) {
	d.lock.Lock()
	defer d.lock.Unlock()

	button := js.Global.Get("debugger-play")
	if running {
		button.Set("textContent", "Stop")
	} else {
		button.Set("textContent", "Play")
	}
}

func (d *Debugger) tickInfo() (framePause time.Duration, instsPerFrame int) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.frequency > 32 {
		return time.Second / 32, d.frequency / 32
	} else {
		return time.Second / time.Duration(d.frequency), 1
	}
}

func (d *Debugger) registerUIEvents() {
	js.Global.Get("debugger-step").Call("addEventListener", "click", func() {
		d.controlChan <- stepDebugger
	})

	playButton := js.Global.Get("debugger-play")
	playButton.Call("addEventListener", "click", func() {
		go func() {
			if playButton.Get("textContent").String() == "Stop" {
				d.controlChan <- stopDebugger
			} else {
				d.controlChan <- startDebugger
			}
		}()
	})

	ratePicker := js.Global.Get("debugger-rate")
	ratePicker.Call("addEventListener", "change", func() {
		go func() {
			rate := ratePicker.Get("value").String()
			num, _ := strconv.Atoi(rate)
			d.lock.Lock()
			d.frequency = num
			d.lock.Unlock()
			d.controlChan <- updateDebuggerFreq
		}()
	})

	js.Global.Get("debugger-reset").Call("addEventListener", "click", func() {
		go d.SetExecutable(nil)
	})
}

type debuggerCommand int

const (
	stopDebugger debuggerCommand = iota
	startDebugger
	stepDebugger
	updateDebuggerFreq
)
