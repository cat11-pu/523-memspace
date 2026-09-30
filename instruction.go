package mips32

import "errors"

// Instruction stores all of the information about an instruction in distinct fields.
// This makes it possible to execute an instruction and see exactly what its operands are.
type Instruction struct {
	// Name is the instruction's name.
	//
	// If this instruction is from a ".word" directive which does not correspond to a valid
	// instruction, then the Name field is ".word" and the RawWord field is set.
	Name string

	// Registers is the list of register indices passed to this instruction.
	// This list is in the same order as the instruction's operands in assembly.
	Registers []int

	UnsignedConstant16 uint16
	SignedConstant16   int16
	Constant5          uint8
	CodePointer        CodePointer
	MemoryReference    MemoryReference

	// RawWord is only used for instructions which cannot be decoded.
	// This is only used when Name is set to ".word"
	RawWord uint32
}

// ParseTokenizedInstruction generates an Instruction which represents a TokenizedInstruction.
// This may fail if the instruction is invalid, in which case an error is returned.
func ParseTokenizedInstruction(t *TokenizedInstruction) (*Instruction, error) {
	validName := false
	for _, template := range Templates {
		if template.Name == t.Name {
			validName = true
		}
		if template.Match(t) {
			res := &Instruction{Name: t.Name}
			for i, arg := range template.Arguments {
				tokArg := t.Arguments[i]
				switch arg {
				case Register:
					reg, _ := tokArg.Register()
					res.Registers = append(res.Registers, reg)
				case SignedConstant16:
					res.SignedConstant16, _ = tokArg.SignedConstant16()
				case UnsignedConstant16:
					res.UnsignedConstant16, _ = tokArg.UnsignedConstant16()
				case Constant5:
					res.Constant5, _ = tokArg.Constant5()
				case AbsoluteCodePointer:
					res.CodePointer, _ = tokArg.AbsoluteCodePointer()
				case RelativeCodePointer:
					res.CodePointer, _ = tokArg.RelativeCodePointer()
				case MemoryAddress:
					res.MemoryReference, _ = tokArg.MemoryReference()
				}
			}
			return res, nil
		}
	}
	if validName {
		return nil, errors.New("bad instruction usage for " + t.Name)
	} else {
		return nil, errors.New("unknown instruction: " + t.Name)
	}
}

// Render generates a *TokenizedLine that represents this instruction.
//
// If this succeeds, the result will normally contain a TokenizedInstruction.
// However, if this is a ".word" instruction, then the result will contain a TokenizedDirective.
//
// This will fail if the instruction's arguments are invalid.
func (i *Instruction) Render() (*TokenizedLine, error) {
	if i.Name == ".word" {
		return &TokenizedLine{
			Directive: &TokenizedDirective{
				Name:     "word",
				Constant: i.RawWord,
			},
		}, nil
	}
	found := false

TemplateLoop:
	for _, template := range Templates {
		if template.Name != i.Name {
			continue
		}
		found = true
		if template.RegisterCount() != len(i.Registers) {
			continue
		}
		res := &TokenizedInstruction{
			Name:      i.Name,
			Arguments: make([]*ArgToken, len(template.Arguments)),
		}
		regIndex := 0
		for argIndex, arg := range template.Arguments {
			switch arg {
			case Register:
				res.Arguments[argIndex] = &ArgToken{
					isRegister: true,
					register:   i.Registers[regIndex],
				}
				regIndex++
			case SignedConstant16:
				res.Arguments[argIndex] = &ArgToken{
					isConstant: true,
					constant:   uint32(i.SignedConstant16),
				}
			case UnsignedConstant16:
				res.Arguments[argIndex] = &ArgToken{
					isConstant: true,
					constant:   uint32(i.UnsignedConstant16),
				}
			case Constant5:
				res.Arguments[argIndex] = &ArgToken{
					isConstant: true,
					constant:   uint32(i.Constant5),
				}
			case AbsoluteCodePointer, RelativeCodePointer:
				if i.CodePointer.Absolute != (arg == AbsoluteCodePointer) {
					continue TemplateLoop
				}
				if i.CodePointer.IsSymbol {
					res.Arguments[argIndex] = &ArgToken{
						isSymbol: true,
						symbol:   i.CodePointer.Symbol,
					}
				} else {
					res.Arguments[argIndex] = &ArgToken{
						isConstant: true,
						constant:   i.CodePointer.Constant,
					}
				}
			case MemoryAddress:
				res.Arguments[argIndex] = &ArgToken{
					isMemory:    true,
					memOffset:   i.MemoryReference.Offset,
					memRegister: i.MemoryReference.Register,
				}
			}
		}
		return &TokenizedLine{Instruction: res}, nil
	}
	if found {
		return nil, errors.New("invalid arguments for " + i.Name)
	} else {
		return nil, errors.New("no such instruction: " + i.Name)
	}
}
