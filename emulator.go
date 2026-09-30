package mips32

import (
	"errors"
	"strconv"
)

type RegisterFile [32]uint32

func (r RegisterFile) String() string {
	res := ""
	for i := 0; i < 16; i++ {
		label1 := "r" + strconv.Itoa(i)
		if len(label1) == 2 {
			label1 += " "
		}
		label2 := "r" + strconv.Itoa(i+16)
		if i != 0 {
			res += "\n"
		}
		res += label1 + " = " + eightDigitHex(r[i]) + "  " + label2 + " = " +
			eightDigitHex(r[i+16])
	}
	return res
}

type Emulator struct {
	RegisterFile   RegisterFile
	Memory         Memory
	Executable     *Executable
	ProgramCounter uint32

	LittleEndian      bool
	ForceMemAlignment bool

	// DelaySlot is set during and after an instruction in the delay slot is executed.
	DelaySlot bool

	// JumpNext is set if a jump/branch instruction was just executed and the delay slot's PC is in
	// the ProgramCounter field.
	JumpNext bool

	// JumpTarget is the target location for the jump/branch referred to by JumpNext.
	JumpTarget uint32
}

// Done returns true if the program has begun to execute NOPs past the executable code.
func (e *Emulator) Done() bool {
	if e.JumpNext {
		return false
	}
	return e.ProgramCounter >= e.Executable.End()
}

// Step performs the next instruction on the CPU.
// If the instruction fails, then this will return an error.
// In the case of an error, the program counter may still be changed as usual.
func (e *Emulator) Step() error {
	inst := e.Executable.Get(e.ProgramCounter)
	if e.JumpNext {
		e.DelaySlot = true
		e.JumpNext = false
		e.ProgramCounter = e.JumpTarget
	} else {
		e.DelaySlot = false
		e.ProgramCounter += 4
	}

	// If there is no instruction in the ROM, we assume it is a NOP.
	if inst == nil {
		return nil
	}

	switch inst.Name {
	case "NOP":
	case "BEQ", "BGEZ", "BGTZ", "BLEZ", "BLTZ", "BNE":
		return e.executeBranch(inst)
	case "J", "JR", "JAL", "JALR":
		return e.executeJump(inst)
	case "LB", "LBU", "LW", "SB", "SW":
		return e.executeMemory(inst)
	case "ADDU", "AND", "NOR", "OR", "SUBU", "XOR":
		e.executeRegisterArithmetic(inst)
	case "ADDIU", "ANDI", "ORI", "XORI":
		e.executeImmediateArithmetic(inst)
	case "LUI":
		e.executeLoadUpperImmediate(inst)
	case "SLT", "SLTI", "SLTIU", "SLTU":
		e.executeSetLessThan(inst)
	case "SLL", "SRL", "SRA":
		e.executeConstantShift(inst)
	case "SLLV", "SRLV", "SRAV":
		e.executeRegisterShift(inst)
	case "MOVN", "MOVZ":
		e.executeConditionalMove(inst)
	default:
		return errors.New("unknown instruction: " + inst.Name)
	}
	return nil
}

func (e *Emulator) executeBranch(inst *Instruction) error {
	if e.DelaySlot {
		return errors.New("branch in delay slot yields unpredictable behavior")
	}

	offset, err := instructionBranchOffset(inst, e.ProgramCounter-4, e.Executable.Symbols)
	if err != nil {
		return e.instructionError(err.Error())
	}
	e.JumpTarget = e.ProgramCounter + offset

	val1 := int32(e.RegisterFile[inst.Registers[0]])
	switch inst.Name {
	case "BEQ", "BNE":
		val2 := int32(e.RegisterFile[inst.Registers[1]])
		e.JumpNext = (val1 == val2) == (inst.Name == "BEQ")
	case "BGEZ":
		e.JumpNext = val1 >= 0
	case "BGTZ":
		e.JumpNext = val1 > 0
	case "BLEZ":
		e.JumpNext = val1 <= 0
	case "BLTZ":
		e.JumpNext = val1 < 0
	}

	return nil
}

func (e *Emulator) executeJump(inst *Instruction) error {
	if e.DelaySlot {
		return errors.New("jump in delay slot yields unpredictable behavior")
	}

	if inst.Name == "J" || inst.Name == "JAL" {
		offset, err := instructionJumpBase(inst, e.ProgramCounter-4, e.Executable.Symbols)
		if err != nil {
			return e.instructionError(err.Error())
		}
		e.JumpTarget = (e.ProgramCounter & 0xf0000000) | offset
		e.JumpNext = true

		if inst.Name == "JAL" {
			e.RegisterFile[31] = e.ProgramCounter + 4
		}
	} else {
		newAddress := e.RegisterFile[inst.Registers[len(inst.Registers)-1]]
		if (newAddress & 3) != 0 {
			return e.instructionError("misaligned address")
		}
		e.JumpTarget = newAddress
		e.JumpNext = true

		if inst.Name == "JALR" {
			destReg := 31
			if len(inst.Registers) == 2 {
				destReg = inst.Registers[0]
			}
			e.setReg(destReg, e.ProgramCounter+4)
		}
	}

	return nil
}

func (e *Emulator) executeMemory(inst *Instruction) error {
	address := e.RegisterFile[inst.MemoryReference.Register] + uint32(inst.MemoryReference.Offset)
	register := inst.Registers[0]
	registerValue := e.RegisterFile[register]

	switch inst.Name {
	case "LB":
		e.setReg(register, uint32(int8(e.Memory.Get(address))))
	case "LBU":
		e.setReg(register, uint32(e.Memory.Get(address)))
	case "LW":
		if e.ForceMemAlignment && (address&3) != 0 {
			return e.instructionError("misaligned load word: 0x" +
				strconv.FormatUint(uint64(address), 16))
		}
		if e.LittleEndian {
			e.setReg(register, (uint32(e.Memory.Get(address+3))<<24)|
				(uint32(e.Memory.Get(address+2))<<16)|(uint32(e.Memory.Get(address+1))<<8)|
				uint32(e.Memory.Get(address)))
		} else {
			e.setReg(register, (uint32(e.Memory.Get(address))<<24)|
				(uint32(e.Memory.Get(address+1))<<16)|(uint32(e.Memory.Get(address+2))<<8)|
				uint32(e.Memory.Get(address+3)))
		}
	case "SB":
		e.Memory.Set(address, byte(registerValue))
	case "SW":
		if e.ForceMemAlignment && (address&3) != 0 {
			return e.instructionError("misaligned store word: 0x" +
				strconv.FormatUint(uint64(address), 16))
		}
		if e.LittleEndian {
			e.Memory.Set(address+3, byte(registerValue>>24))
			e.Memory.Set(address+2, byte(registerValue>>16))
			e.Memory.Set(address+1, byte(registerValue>>8))
			e.Memory.Set(address, byte(registerValue))
		} else {
			e.Memory.Set(address, byte(registerValue>>24))
			e.Memory.Set(address+1, byte(registerValue>>16))
			e.Memory.Set(address+2, byte(registerValue>>8))
			e.Memory.Set(address+3, byte(registerValue))
		}
	}

	return nil
}

func (e *Emulator) executeRegisterArithmetic(inst *Instruction) {
	val1 := e.RegisterFile[inst.Registers[1]]
	val2 := e.RegisterFile[inst.Registers[2]]
	var result uint32
	switch inst.Name {
	case "ADDU":
		result = val1 + val2
	case "AND":
		result = val1 & val2
	case "OR":
		result = val1 | val2
	case "NOR":
		result = ^(val1 | val2)
	case "SUBU":
		result = val1 - val2
	case "XOR":
		result = val1 ^ val2
	}
	e.setReg(inst.Registers[0], result)
}

func (e *Emulator) executeImmediateArithmetic(inst *Instruction) {
	val1 := e.RegisterFile[inst.Registers[1]]
	var val2 uint32
	if inst.Name == "ADDIU" {
		val2 = uint32(inst.SignedConstant16)
	} else {
		val2 = uint32(inst.UnsignedConstant16)
	}

	var result uint32
	switch inst.Name {
	case "ADDIU":
		result = val1 + val2
	case "ANDI":
		result = val1 & val2
	case "ORI":
		result = val1 | val2
	case "XORI":
		result = val1 ^ val2
	}
	e.setReg(inst.Registers[0], result)
}

func (e *Emulator) executeLoadUpperImmediate(inst *Instruction) {
	val := uint32(inst.UnsignedConstant16) << 16
	e.setReg(inst.Registers[0], val)
}

func (e *Emulator) executeSetLessThan(inst *Instruction) {
	val1 := e.RegisterFile[inst.Registers[1]]
	var val2 uint32
	if inst.Name == "SLTI" || inst.Name == "SLTIU" {
		val2 = uint32(inst.SignedConstant16)
	} else {
		val2 = e.RegisterFile[inst.Registers[2]]
	}

	var res bool
	switch inst.Name {
	case "SLTI", "SLT":
		res = int32(val1) < int32(val2)
	default:
		res = val1 < val2
	}

	if res {
		e.setReg(inst.Registers[0], 1)
	} else {
		e.setReg(inst.Registers[0], 0)
	}
}

func (e *Emulator) executeConstantShift(inst *Instruction) {
	shiftAmount := inst.Constant5
	val := e.RegisterFile[inst.Registers[1]]

	var res uint32
	switch inst.Name {
	case "SLL":
		res = val << shiftAmount
	case "SRA":
		res = uint32(int32(val) >> shiftAmount)
	case "SRL":
		res = val >> shiftAmount
	}

	e.setReg(inst.Registers[0], res)
}

func (e *Emulator) executeRegisterShift(inst *Instruction) {
	shiftAmount := e.RegisterFile[inst.Registers[2]] & 0x1f
	val := e.RegisterFile[inst.Registers[1]]

	var res uint32
	switch inst.Name {
	case "SLLV":
		res = val << shiftAmount
	case "SRAV":
		res = uint32(int32(val) >> shiftAmount)
	case "SRLV":
		res = val >> shiftAmount
	}

	e.setReg(inst.Registers[0], res)
}

func (e *Emulator) executeConditionalMove(inst *Instruction) {
	condition := e.RegisterFile[inst.Registers[2]]

	var doMove bool
	switch inst.Name {
	case "MOVN":
		doMove = (condition != 0)
	case "MOVZ":
		doMove = (condition == 0)
	}

	if doMove {
		value := e.RegisterFile[inst.Registers[1]]
		e.setReg(inst.Registers[0], value)
	}
}

func (e *Emulator) instructionError(msg string) error {
	pc := e.ProgramCounter - 4
	pcStr := "0x" + strconv.FormatUint(uint64(pc), 16)
	return errors.New("error at " + pcStr + ": " + msg)
}

func (e *Emulator) setReg(r int, val uint32) {
	if r == 0 {
		return
	}
	e.RegisterFile[r] = val
}

func eightDigitHex(n uint32) string {
	s := strconv.FormatUint(uint64(n), 16)
	for len(s) < 8 {
		s = "0" + s
	}
	return "0x" + s
}
