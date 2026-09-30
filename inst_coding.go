package mips32

import "errors"

var twoOperandImmediateOpcodes = map[uint32]string{
	0x09: "ADDIU",
	0x0c: "ANDI",
	0x0d: "ORI",
	0x0a: "SLTI",
	0x0b: "SLTIU",
	0x0e: "XORI",
}

var branchOpcodes = map[uint32]string{
	0x04: "BEQ",
	0x01: "BLTZ",
	0x07: "BGTZ",
	0x06: "BLEZ",
	0x05: "BNE",
}

var jTypeOpcodes = map[uint32]string{
	0x02: "J",
	0x03: "JAL",
}

var memoryOpcodes = map[uint32]string{
	0x20: "LB",
	0x24: "LBU",
	0x23: "LW",
	0x28: "SB",
	0x2b: "SW",
}

var constantShiftFuncs = map[uint32]string{
	0x00: "SLL",
	0x03: "SRA",
	0x02: "SRL",
}

var variableShiftFuncs = map[uint32]string{
	0x04: "SLLV",
	0x07: "SRAV",
	0x06: "SRLV",
}

var threeRegOperandFuncs = map[uint32]string{
	0x21: "ADDU",
	0x24: "AND",
	0x0b: "MOVN",
	0x0a: "MOVZ",
	0x27: "NOR",
	0x25: "OR",
	0x2a: "SLT",
	0x2b: "SLTU",
	0x23: "SUBU",
	0x26: "XOR",
}

const luiOpcode = 0x0f
const jrFunc = 0x08
const jalrFunc = 0x09

// DecodeInstruction returns an Instruction for a 32-bit word.
// This can never fail, since invalid instructions can be treated as ".word" directives.
func DecodeInstruction(word uint32) *Instruction {
	if word == 0 {
		return &Instruction{Name: "NOP"}
	}

	opcode := (word >> 26) & 0x3f
	registerS := int((word >> 21) & 0x1f)
	registerT := int((word >> 16) & 0x1f)
	registerD := int((word >> 11) & 0x1f)
	shiftAmount := uint8((word >> 6) & 0x1f)
	funcField := word & 0x3f
	immediate := word & 0xffff

	if instName, ok := twoOperandImmediateOpcodes[opcode]; ok {
		return &Instruction{
			Name:               instName,
			Registers:          []int{registerT, registerS},
			UnsignedConstant16: uint16(immediate),
			SignedConstant16:   int16(immediate),
		}
	}

	if opcode == luiOpcode && registerS == 0 {
		return &Instruction{
			Name:               "LUI",
			Registers:          []int{registerT},
			UnsignedConstant16: uint16(immediate),
		}
	}

	if instName, ok := branchOpcodes[opcode]; ok {
		if instName == "BEQ" || instName == "BNE" {
			return &Instruction{
				Name:        instName,
				Registers:   []int{registerS, registerT},
				CodePointer: CodePointer{Constant: uint32(int16(immediate)) << 2},
			}
		}
		if instName == "BLTZ" && registerT == 1 {
			instName = "BGEZ"
			registerT = 0
		}
		if registerT == 0 {
			return &Instruction{
				Name:        instName,
				Registers:   []int{registerS},
				CodePointer: CodePointer{Constant: uint32(int16(immediate)) << 2},
			}
		}
	}

	if instName, ok := jTypeOpcodes[opcode]; ok {
		jumpAddr := (word & 0x03ffffff) << 2
		return &Instruction{
			Name:        instName,
			CodePointer: CodePointer{Absolute: true, Constant: jumpAddr},
		}
	}

	if instName, ok := memoryOpcodes[opcode]; ok {
		return &Instruction{
			Name:            instName,
			Registers:       []int{registerT},
			MemoryReference: MemoryReference{Register: registerS, Offset: int16(immediate)},
		}
	}

	if opcode == 0 {
		if instName, ok := constantShiftFuncs[funcField]; ok && registerS == 0 {
			return &Instruction{
				Name:      instName,
				Registers: []int{registerD, registerT},
				Constant5: shiftAmount,
			}
		}

		if instName, ok := variableShiftFuncs[funcField]; ok && shiftAmount == 0 {
			return &Instruction{
				Name:      instName,
				Registers: []int{registerD, registerT, registerS},
			}
		}

		if instName, ok := threeRegOperandFuncs[funcField]; ok && shiftAmount == 0 {
			return &Instruction{
				Name:      instName,
				Registers: []int{registerD, registerS, registerT},
			}
		}

		if opcode == 0 && registerT == 0 && registerD == 0 &&
			shiftAmount == 0 && funcField == jrFunc {
			return &Instruction{
				Name:      "JR",
				Registers: []int{registerS},
			}
		}

		if opcode == 0 && registerT == 0 && shiftAmount == 0 && funcField == jalrFunc {
			return &Instruction{
				Name:      "JALR",
				Registers: []int{registerD, registerS},
			}
		}
	}

	return &Instruction{
		Name:    ".word",
		RawWord: word,
	}
}

// Encode returns the 32-bit word representation of this instruction, or an error if the instruction
// fails to meet any available templates.
//
// This method needs the address of the instruction and a symbol table in order to encode control
// instructions, since it must resolve (potentially relative) symbol addresses.
func (inst *Instruction) Encode(instAddr uint32, symbols map[string]uint32) (uint32, error) {
	if inst.Name == ".word" {
		return inst.RawWord, nil
	} else if inst.Name == "NOP" {
		if len(inst.Registers) != 0 {
			return 0, registerCountError(inst.Name)
		}
		return 0, nil
	}

	if opcode, ok := numberForInstruction(twoOperandImmediateOpcodes, inst.Name); ok {
		if len(inst.Registers) != 2 {
			return 0, registerCountError(inst.Name)
		}
		return (opcode << 26) | (uint32(inst.Registers[1]) << 21) |
			(uint32(inst.Registers[0]) << 16) | uint32(uint16(inst.SignedConstant16)) |
			uint32(inst.UnsignedConstant16), nil
	}

	if inst.Name == "LUI" {
		if len(inst.Registers) != 1 {
			return 0, registerCountError(inst.Name)
		}
		return (luiOpcode << 26) | (uint32(inst.Registers[0]) << 16) |
			uint32(inst.UnsignedConstant16), nil
	}

	branchName := inst.Name
	if branchName == "BGEZ" {
		branchName = "BLTZ"
	}
	if opcode, ok := numberForInstruction(branchOpcodes, branchName); ok {
		branchOffset, err := instructionBranchOffset(inst, instAddr, symbols)
		if err != nil {
			return 0, err
		}
		if inst.Name == "BEQ" || inst.Name == "BNE" {
			if len(inst.Registers) != 2 {
				return 0, registerCountError(inst.Name)
			}
			return (opcode << 26) | (uint32(inst.Registers[0]) << 21) |
				(uint32(inst.Registers[1]) << 16) | ((branchOffset >> 2) & 0xffff), nil
		}
		if len(inst.Registers) != 1 {
			return 0, registerCountError(inst.Name)
		}
		var regT uint32
		if inst.Name == "BGEZ" {
			regT = 1
		}
		return (opcode << 26) | (uint32(inst.Registers[0]) << 21) |
			(regT << 16) | ((branchOffset >> 2) & 0xffff), nil
	}

	if opcode, ok := numberForInstruction(jTypeOpcodes, inst.Name); ok {
		if len(inst.Registers) != 0 {
			return 0, registerCountError(inst.Name)
		}
		jumpAddr, err := instructionJumpBase(inst, instAddr, symbols)
		if err != nil {
			return 0, err
		}
		return (opcode << 26) | (jumpAddr >> 2), nil
	}

	if opcode, ok := numberForInstruction(memoryOpcodes, inst.Name); ok {
		if len(inst.Registers) != 1 {
			return 0, registerCountError(inst.Name)
		}
		return (opcode << 26) | (uint32(inst.MemoryReference.Register) << 21) |
			(uint32(inst.Registers[0]) << 16) | uint32(uint16(inst.MemoryReference.Offset)), nil
	}

	if funcField, ok := numberForInstruction(constantShiftFuncs, inst.Name); ok {
		if len(inst.Registers) != 2 {
			return 0, registerCountError(inst.Name)
		}
		return (uint32(inst.Registers[1]) << 16) | (uint32(inst.Registers[0]) << 11) |
			(uint32(inst.Constant5) << 6) | funcField, nil
	}

	if funcField, ok := numberForInstruction(variableShiftFuncs, inst.Name); ok {
		if len(inst.Registers) != 3 {
			return 0, registerCountError(inst.Name)
		}
		return (uint32(inst.Registers[2]) << 21) | (uint32(inst.Registers[1]) << 16) |
			(uint32(inst.Registers[0]) << 11) | funcField, nil
	}

	if funcField, ok := numberForInstruction(threeRegOperandFuncs, inst.Name); ok {
		if len(inst.Registers) != 3 {
			return 0, registerCountError(inst.Name)
		}
		return (uint32(inst.Registers[1]) << 21) | (uint32(inst.Registers[2]) << 16) |
			(uint32(inst.Registers[0]) << 11) | funcField, nil
	}

	if inst.Name == "JR" {
		if len(inst.Registers) != 1 {
			return 0, registerCountError(inst.Name)
		}
		return (uint32(inst.Registers[0]) << 21) | jrFunc, nil
	}

	if inst.Name == "JALR" {
		switch len(inst.Registers) {
		case 1:
			return (uint32(inst.Registers[0]) << 21) | (31 << 11) | jalrFunc, nil
		case 2:
			return (uint32(inst.Registers[1]) << 21) | (uint32(inst.Registers[0]) << 11) |
				jalrFunc, nil
		default:
			return 0, registerCountError(inst.Name)
		}
	}

	return 0, errors.New("unknown instruction: " + inst.Name)
}

func instructionBranchOffset(inst *Instruction, instAddr uint32,
	symbols map[string]uint32) (uint32, error) {
	if inst.CodePointer.Absolute {
		return 0, errors.New("expecting relative code pointer for " + inst.Name)
	} else if inst.CodePointer.IsSymbol {
		if addr, ok := symbols[inst.CodePointer.Symbol]; !ok {
			return 0, unknownSymbolError(inst.CodePointer.Symbol)
		} else {
			diff := (int32(addr) - int32(instAddr+4)) >> 2
			if diff >= 0x8000 || diff < -0x8000 {
				return 0, errors.New("branch offset out of bounds")
			}
			return uint32(int16(diff)) << 2, nil
		}
	} else {
		if inst.CodePointer.Constant&3 != 0 {
			return 0, errors.New("misaligned address")
		}
		return inst.CodePointer.Constant, nil
	}
}

func instructionJumpBase(inst *Instruction, instAddr uint32,
	symbols map[string]uint32) (uint32, error) {
	if !inst.CodePointer.Absolute {
		return 0, errors.New("expecting absolute code pointer for " + inst.Name)
	} else if inst.CodePointer.IsSymbol {
		if addr, ok := symbols[inst.CodePointer.Symbol]; !ok {
			return 0, unknownSymbolError(inst.CodePointer.Symbol)
		} else if (addr & 0xf0000000) != ((instAddr + 4) & 0xf0000000) {
			return 0, errors.New("jump address overflows 26 bits")
		} else {
			return (addr & 0x0fffffff), nil
		}
	} else {
		addr := inst.CodePointer.Constant
		if (addr & 0xf0000000) != ((instAddr + 4) & 0xf0000000) {
			return 0, errors.New("cannot encode jump address in 26 bits")
		} else if (addr & 3) != 0 {
			return 0, errors.New("misaligned address")
		} else {
			return (addr & 0x0fffffff), nil
		}
	}
}

func numberForInstruction(m map[uint32]string, inst string) (uint32, bool) {
	for number, name := range m {
		if name == inst {
			return number, true
		}
	}
	return 0, false
}

func registerCountError(instName string) error {
	return errors.New("invalid number of registers for " + instName)
}

func unknownSymbolError(symbol string) error {
	return errors.New("unknown symbol: " + symbol)
}
