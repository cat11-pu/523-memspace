package mips32

type ArgumentType int

const (
	Register ArgumentType = iota
	SignedConstant16
	UnsignedConstant16
	Constant5
	AbsoluteCodePointer
	RelativeCodePointer
	MemoryAddress
)

// A Template describes the kinds of arguments an instruction can take.
type Template struct {
	Name      string
	Arguments []ArgumentType
}

func (t *Template) Match(tok *TokenizedInstruction) bool {
	if t.Name != tok.Name || len(t.Arguments) != len(tok.Arguments) {
		return false
	}
	for i, arg := range t.Arguments {
		tokArg := tok.Arguments[i]
		switch arg {
		case Register:
			if _, ok := tokArg.Register(); !ok {
				return false
			}
		case SignedConstant16:
			if _, ok := tokArg.SignedConstant16(); !ok {
				return false
			}
		case UnsignedConstant16:
			if _, ok := tokArg.UnsignedConstant16(); !ok {
				return false
			}
		case Constant5:
			if _, ok := tokArg.Constant5(); !ok {
				return false
			}
		case AbsoluteCodePointer:
			if _, ok := tokArg.AbsoluteCodePointer(); !ok {
				return false
			}
		case RelativeCodePointer:
			if _, ok := tokArg.RelativeCodePointer(); !ok {
				return false
			}
		case MemoryAddress:
			if _, ok := tokArg.MemoryReference(); !ok {
				return false
			}
		}
	}
	return true
}

func (t *Template) RegisterCount() int {
	count := 0
	for _, arg := range t.Arguments {
		if arg == Register {
			count++
		}
	}
	return count
}

var Templates = []Template{
	{"NOP", []ArgumentType{}},
	{"ADDIU", []ArgumentType{Register, Register, SignedConstant16}},
	{"ADDU", []ArgumentType{Register, Register, Register}},
	{"AND", []ArgumentType{Register, Register, Register}},
	{"ANDI", []ArgumentType{Register, Register, UnsignedConstant16}},
	{"BEQ", []ArgumentType{Register, Register, RelativeCodePointer}},
	{"BGEZ", []ArgumentType{Register, RelativeCodePointer}},
	{"BGTZ", []ArgumentType{Register, RelativeCodePointer}},
	{"BLEZ", []ArgumentType{Register, RelativeCodePointer}},
	{"BLTZ", []ArgumentType{Register, RelativeCodePointer}},
	{"BNE", []ArgumentType{Register, Register, RelativeCodePointer}},
	{"J", []ArgumentType{AbsoluteCodePointer}},
	{"JAL", []ArgumentType{AbsoluteCodePointer}},
	{"JALR", []ArgumentType{Register}},
	{"JALR", []ArgumentType{Register, Register}},
	{"JR", []ArgumentType{Register}},
	{"LB", []ArgumentType{Register, MemoryAddress}},
	{"LBU", []ArgumentType{Register, MemoryAddress}},
	{"LW", []ArgumentType{Register, MemoryAddress}},
	{"SB", []ArgumentType{Register, MemoryAddress}},
	{"SW", []ArgumentType{Register, MemoryAddress}},
	{"LUI", []ArgumentType{Register, UnsignedConstant16}},
	{"MOVN", []ArgumentType{Register, Register, Register}},
	{"MOVZ", []ArgumentType{Register, Register, Register}},
	{"NOR", []ArgumentType{Register, Register, Register}},
	{"OR", []ArgumentType{Register, Register, Register}},
	{"ORI", []ArgumentType{Register, Register, UnsignedConstant16}},
	{"SLL", []ArgumentType{Register, Register, Constant5}},
	{"SLLV", []ArgumentType{Register, Register, Register}},
	{"SLT", []ArgumentType{Register, Register, Register}},
	{"SLTI", []ArgumentType{Register, Register, SignedConstant16}},
	{"SLTIU", []ArgumentType{Register, Register, SignedConstant16}},
	{"SLTU", []ArgumentType{Register, Register, Register}},
	{"SRA", []ArgumentType{Register, Register, Constant5}},
	{"SRAV", []ArgumentType{Register, Register, Register}},
	{"SRL", []ArgumentType{Register, Register, Constant5}},
	{"SRLV", []ArgumentType{Register, Register, Register}},
	{"SUBU", []ArgumentType{Register, Register, Register}},
	{"XOR", []ArgumentType{Register, Register, Register}},
	{"XORI", []ArgumentType{Register, Register, UnsignedConstant16}},
}
