package mips32

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	constantNumberPattern = "(-?[0-9]*|-?0x[0-9a-fA-F]*)"
	symbolNamePattern     = "([a-zA-Z0-9_]*)"
)

var (
	constantRegexp        = regexp.MustCompile("^" + constantNumberPattern + "$")
	symbolRegexp          = regexp.MustCompile("^" + symbolNamePattern + "$")
	memoryRegexp          = regexp.MustCompile("^(" + constantNumberPattern + "|)\\(.+\\)$")
	memorySubfieldsRegexp = regexp.MustCompile("^(.*)\\((.*)\\)$")
)

var registerNames = map[string]int{
	"0": 0, "zero": 0, "1": 1, "at": 1, "2": 2, "v0": 2, "3": 3, "v1": 3, "4": 4, "a0": 4, "5": 5,
	"a1": 5, "6": 6, "a2": 6, "7": 7, "a3": 7, "8": 8, "t0": 8, "9": 9, "t1": 9, "10": 10, "t2": 10,
	"11": 11, "t3": 11, "12": 12, "t4": 12, "13": 13, "t5": 13, "14": 14, "t6": 14, "15": 15,
	"t7": 15, "16": 16, "s0": 16, "17": 17, "s1": 17, "18": 18, "s2": 18, "19": 19, "s3": 19,
	"20": 20, "s4": 20, "21": 21, "s5": 21, "22": 22, "s6": 22, "23": 23, "s7": 23, "24": 24,
	"t8": 24, "25": 25, "t9": 25, "26": 26, "k0": 26, "27": 27, "k1": 27, "28": 28, "gp": 28,
	"29": 29, "sp": 29, "30": 30, "s8": 30, "fp": 30, "31": 31, "ra": 31,

	"r0": 0, "r1": 1, "r2": 2, "r3": 3, "r4": 4, "r5": 5, "r6": 6, "r7": 7, "r8": 8, "r9": 9,
	"r10": 10, "r11": 11, "r12": 12, "r13": 13, "r14": 14, "r15": 15, "r16": 16, "r17": 17,
	"r18": 18, "r19": 19, "r20": 20, "r21": 21, "r22": 22, "r23": 23, "r24": 24, "r25": 25,
	"r26": 26, "r27": 27, "r28": 28, "r29": 29, "r30": 30, "r31": 31,
}

// A CodePointer contains some kind of information indicating where a piece of code is.
type CodePointer struct {
	// Absolute is true if this is an AbsoluteCodePointer.
	Absolute bool

	// IsSymbol is true if this is represented by a symbol rather than a constant.
	IsSymbol bool

	Symbol   string
	Constant uint32
}

// A MemoryReference stores information about a memory reference.
type MemoryReference struct {
	Register int
	Offset   int16
}

// An ArgToken represents a register, a number, a symbol, or a memory location.
// For instance, the instruction "SB $5, 5($6)" contains two tokens.
//
// An ArgToken may be able to serve as multiple types of arguments.
// For example, the ArgToken for "0x5" could be a Constant5, a Constant16, or a CodePointer.
type ArgToken struct {
	isRegister bool
	register   int

	isConstant bool
	constant   uint32

	isSymbol bool
	symbol   string

	isMemory    bool
	memRegister int
	memOffset   int16
}

// ParseArgToken parses a human-readable token string.
// The string must not have any leading or trailing whitespace.
func ParseArgToken(tokenStr string) (token *ArgToken, err error) {
	if regToken, err := parseRegisterArgToken(tokenStr); err == nil {
		return regToken, nil
	} else if constantRegexp.MatchString(tokenStr) {
		return parseConstantArgToken(tokenStr)
	} else if symbolRegexp.MatchString(tokenStr) {
		return parseSymbolArgToken(tokenStr)
	} else if memoryRegexp.MatchString(tokenStr) {
		return parseMemoryArgToken(tokenStr)
	}
	return nil, errors.New("unable to parse token: " + tokenStr)
}

// Register returns the register index represented by this token.
// If this token cannot be treated as a register index, ok will be false.
func (t *ArgToken) Register() (regIndex int, ok bool) {
	return t.register, t.isRegister
}

// UnsignedConstant16 returns the 16-bit zero-extended constant represented by this token.
// If this token cannot be treated as an unsigned 16-bit constant, ok will be false.
func (t *ArgToken) UnsignedConstant16() (constant uint16, ok bool) {
	return uint16(t.constant), t.isConstant && (t.constant&0xffff0000) == 0
}

// SignedConstant16 returns the 16-bit sign-extended constant represented by this token.
// If this token cannot be treated as a signed 16-bit constant, ok will be false.
func (t *ArgToken) SignedConstant16() (constant int16, ok bool) {
	constant = int16(t.constant)
	ok = t.isConstant && t.constant == uint32(constant)
	return
}

// Constant5 returns the 5-bit unsigned constant represented by this token.
// If this token cannot be treated as a 5-bit constant, ok will be false.
func (t *ArgToken) Constant5() (constant uint8, ok bool) {
	return uint8(t.constant), t.isConstant && t.constant < 0x20
}

// RelativeCodePointer returns the relative code pointer represented by this token.
// If this token cannot be treated as a relative code pointer, ok will be false.
//
// If the code pointer is a constant, the constant value will represent an 18-bit signed address
// which is signed extended to 32 bits.
func (t *ArgToken) RelativeCodePointer() (ptr CodePointer, ok bool) {
	if t.isConstant {
		constant := int16(t.constant >> 2)
		if uint32(constant)<<2 != t.constant&0xfffffffc {
			return
		}
		return CodePointer{Constant: t.constant}, true
	} else if t.isSymbol {
		return CodePointer{IsSymbol: true, Symbol: t.symbol}, true
	} else {
		return
	}
}

// AbsoluteCodePointer returns the absolute code pointer represented by this token.
// If this token cannot be treated as an absolute code pointer, ok will be false.
//
// If the code pointer is a constant, it will be an absolute jump destination.
// The destination address should include the high bits of the intended PC+4 value.
func (t *ArgToken) AbsoluteCodePointer() (ptr CodePointer, ok bool) {
	if t.isConstant {
		return CodePointer{Absolute: true, Constant: uint32(t.constant)}, true
	} else if t.isSymbol {
		return CodePointer{Absolute: true, IsSymbol: true, Symbol: t.symbol}, true
	} else {
		return
	}
}

// MemoryReference returns the MemoryReference represented by this token.
// If this token cannot be treated as a MemoryReference, ok will be false.
func (t *ArgToken) MemoryReference() (ref MemoryReference, ok bool) {
	return MemoryReference{Register: t.memRegister, Offset: t.memOffset}, t.isMemory
}

func parseRegisterArgToken(tokenStr string) (token *ArgToken, err error) {
	regNum, err := parseRegister(tokenStr)
	if err != nil {
		return nil, err
	} else {
		return &ArgToken{isRegister: true, register: regNum}, nil
	}
}

func parseConstantArgToken(tokenStr string) (token *ArgToken, err error) {
	num, err := parseConstant(tokenStr)
	if err != nil {
		return nil, err
	}
	return &ArgToken{isConstant: true, constant: num}, nil
}

func parseSymbolArgToken(tokenStr string) (token *ArgToken, err error) {
	return &ArgToken{isSymbol: true, symbol: tokenStr}, nil
}

func parseMemoryArgToken(tokenStr string) (token *ArgToken, err error) {
	pieces := memorySubfieldsRegexp.FindStringSubmatch(tokenStr)
	if pieces == nil {
		return nil, errors.New("invalid memory reference: " + tokenStr)
	}

	reg, err := parseRegister(pieces[2])
	if err != nil {
		return
	}

	var offset int16
	if len(pieces[1]) != 0 {
		if offNum, err := parseConstant(pieces[1]); err != nil {
			return nil, err
		} else if (offNum&0xffff8000) != 0xffff8000 && (offNum&0xffff8000) != 0 {
			return nil, errors.New("memory offset out of bounds: " + pieces[1])
		} else {
			offset = int16(offNum)
		}
	}

	return &ArgToken{isMemory: true, memOffset: offset, memRegister: reg}, nil
}

func parseRegister(tokenStr string) (regIndex int, err error) {
	if !strings.HasPrefix(tokenStr, "$") {
		return 0, errors.New("missing $ in register name: " + tokenStr)
	}
	rawName := tokenStr[1:]
	if regNum, ok := registerNames[rawName]; ok {
		return regNum, nil
	} else {
		return 0, errors.New("invalid register name: " + tokenStr)
	}
}

func parseConstant(tokenStr string) (constant uint32, err error) {
	resNum, err := strconv.ParseInt(tokenStr, 0, 64)
	if err != nil {
		return 0, err
	}
	if resNum > 0xffffffff || resNum < -0xffffffff {
		return 0, err
	}
	return uint32(resNum), nil
}
