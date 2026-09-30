package mips32

import (
	"strconv"
	"testing"
)

func TestParseArgToken(t *testing.T) {
	token, err := ParseArgToken("5")
	if err != nil {
		t.Error(err)
	} else {
		if val, ok := token.Constant5(); !ok {
			t.Error("5 is a Constant5")
		} else if val != 5 {
			t.Error("invalid Constant5 for 5:", val)
		}
		if _, ok := token.MemoryReference(); ok {
			t.Error("5 is not a MemoryReference")
		}
		if _, ok := token.Register(); ok {
			t.Error("5 is not a Register")
		}
		if ptr, ok := token.AbsoluteCodePointer(); !ok {
			t.Error("5 is an absolute code pointer")
		} else if !ptr.Absolute {
			t.Error("absolute flag should be set")
		} else if ptr.IsSymbol {
			t.Error("5 is not a symbol")
		} else if ptr.Constant != 5 {
			t.Error("invalid AbsoluteCodePointer for 5:", ptr.Constant)
		}
		if ptr, ok := token.RelativeCodePointer(); !ok {
			t.Error("5 is a relative code pointer")
		} else if ptr.Absolute {
			t.Error("absolute flag should not be set")
		} else if ptr.IsSymbol {
			t.Error("5 is not a symbol")
		} else if ptr.Constant != 5 {
			t.Error("invalid RelativeCodePointer for 5:", ptr.Constant)
		}
		if val, ok := token.SignedConstant16(); !ok {
			t.Error("5 is a SignedConstant16")
		} else if val != 5 {
			t.Error("invalid SignedConstant16 for 5:", val)
		}
		if val, ok := token.UnsignedConstant16(); !ok {
			t.Error("5 is an UnsignedConstant16")
		} else if val != 5 {
			t.Error("invalid UnsignedConstant16 for 5:", val)
		}
	}

	token, err = ParseArgToken("0x50")
	if err != nil {
		t.Error(err)
	} else {
		if _, ok := token.Constant5(); ok {
			t.Error("0x50 is not a Constant5")
		}
		if _, ok := token.MemoryReference(); ok {
			t.Error("0x50 is not a MemoryReference")
		}
		if _, ok := token.Register(); ok {
			t.Error("0x50 is not a Register")
		}
		if ptr, ok := token.AbsoluteCodePointer(); !ok {
			t.Error("0x50 is an absolute code pointer")
		} else if !ptr.Absolute {
			t.Error("absolute flag should be set")
		} else if ptr.IsSymbol {
			t.Error("0x50 is not a symbol")
		} else if ptr.Constant != 0x50 {
			t.Error("invalid AbsoluteCodePointer for 0x50:", ptr.Constant)
		}
		if ptr, ok := token.RelativeCodePointer(); !ok {
			t.Error("0x50 is a relative code pointer")
		} else if ptr.Absolute {
			t.Error("absolute flag should not be set")
		} else if ptr.IsSymbol {
			t.Error("0x50 is not a symbol")
		} else if ptr.Constant != 0x50 {
			t.Error("invalid RelativeCodePointer for 0x50:", ptr.Constant)
		}
		if val, ok := token.SignedConstant16(); !ok {
			t.Error("0x50 is a SignedConstant16")
		} else if val != 0x50 {
			t.Error("invalid SignedConstant16 for 0x50:", val)
		}
		if val, ok := token.UnsignedConstant16(); !ok {
			t.Error("0x50 is an UnsignedConstant16")
		} else if val != 0x50 {
			t.Error("invalid UnsignedConstant16 for 0x50:", val)
		}
	}

	token, err = ParseArgToken("-0x50")
	if err != nil {
		t.Error(err)
	} else {
		if _, ok := token.Constant5(); ok {
			t.Error("-0x50 is not a Constant5")
		}
		if _, ok := token.MemoryReference(); ok {
			t.Error("-0x50 is not a MemoryReference")
		}
		if _, ok := token.Register(); ok {
			t.Error("-0x50 is not a Register")
		}
		if ptr, ok := token.AbsoluteCodePointer(); !ok {
			t.Error("-0x50 is an absolute code pointer")
		} else if !ptr.Absolute {
			t.Error("absolute flag should be set")
		} else if ptr.IsSymbol {
			t.Error("-0x50 is not a symbol")
		} else if int32(ptr.Constant) != -0x50 {
			t.Error("invalid AbsoluteCodePointer for -0x50:", int32(ptr.Constant))
		}
		if ptr, ok := token.RelativeCodePointer(); !ok {
			t.Error("-0x50 is a relative code pointer")
		} else if ptr.Absolute {
			t.Error("absolute flag should not be set")
		} else if ptr.IsSymbol {
			t.Error("-0x50 is not a symbol")
		} else if int32(ptr.Constant) != -0x50 {
			t.Error("invalid RelativeCodePointer for -0x50:", int32(ptr.Constant))
		}
		if val, ok := token.SignedConstant16(); !ok {
			t.Error("-0x50 is a SignedConstant16")
		} else if val != -0x50 {
			t.Error("invalid SignedConstant16 for -0x50:", val)
		}
		if _, ok := token.UnsignedConstant16(); ok {
			t.Error("-0x50 is not an UnsignedConstant16")
		}
	}

	regTests := map[string]int{"$t0": 8, "$8": 8, "$r8": 8, "$a0": 4, "$4": 4, "$r15": 15,
		"$r31": 31}
	for regName, regNum := range regTests {
		token, err = ParseArgToken(regName)
		if err != nil {
			t.Error(err)
			continue
		}
		if n, ok := token.Register(); !ok {
			t.Error(regName, "is a register")
		} else if n != regNum {
			t.Error("invalid register number for", regName, n, "expected", regNum)
		}
		if _, ok := token.AbsoluteCodePointer(); ok {
			t.Error(regName, "is not an absolute code pointer")
		}
		if _, ok := token.RelativeCodePointer(); ok {
			t.Error(regName, "is not a relative code pointer")
		}
		if _, ok := token.Constant5(); ok {
			t.Error(regName, "is not a Constant5")
		}
		if _, ok := token.SignedConstant16(); ok {
			t.Error(regName, "is not a SignedConstant16")
		}
		if _, ok := token.UnsignedConstant16(); ok {
			t.Error(regName, "is not a UnsignedConstant16")
		}
		if _, ok := token.MemoryReference(); ok {
			t.Error(regName, "is not a MemoryReference")
		}
	}

	token, err = ParseArgToken("Monkey5")
	if err != nil {
		t.Error(err)
	} else {
		if _, ok := token.Register(); ok {
			t.Error("Monkey5 is not a register")
		}
		if ptr, ok := token.AbsoluteCodePointer(); !ok {
			t.Error("Monkey5 is an absolute code pointer")
		} else if !ptr.Absolute {
			t.Error("absolute flag should be set")
		} else if !ptr.IsSymbol {
			t.Error("symbol flag should be set")
		} else if ptr.Symbol != "Monkey5" {
			t.Error("invalid symbol for Monkey5:", ptr.Symbol)
		}
		if ptr, ok := token.RelativeCodePointer(); !ok {
			t.Error("Monkey5 is a relative code pointer")
		} else if ptr.Absolute {
			t.Error("absolute flag should not be set")
		} else if !ptr.IsSymbol {
			t.Error("symbol flag should be set")
		} else if ptr.Symbol != "Monkey5" {
			t.Error("invalid symbol for Monkey5:", ptr.Symbol)
		}
		if _, ok := token.Constant5(); ok {
			t.Error("Monkey5 is not a Constant5")
		}
		if _, ok := token.SignedConstant16(); ok {
			t.Error("Monkey5 is not a SignedConstant16")
		}
		if _, ok := token.UnsignedConstant16(); ok {
			t.Error("Monkey5 is not a UnsignedConstant16")
		}
		if _, ok := token.MemoryReference(); ok {
			t.Error("Monkey5 is not a MemoryReference")
		}
	}

	token, err = ParseArgToken("-0x50($r5)")
	if err != nil {
		t.Error(err)
	} else {
		if _, ok := token.Register(); ok {
			t.Error("-0x50($r5) is not a register")
		}
		if _, ok := token.AbsoluteCodePointer(); ok {
			t.Error("-0x50($r5) is not an absolute code pointer")
		}
		if _, ok := token.RelativeCodePointer(); ok {
			t.Error("-0x50($r5) is not a relative code pointer")
		}
		if _, ok := token.Constant5(); ok {
			t.Error("-0x50($r5) is not a Constant5")
		}
		if _, ok := token.SignedConstant16(); ok {
			t.Error("-0x50($r5) is not a SignedConstant16")
		}
		if _, ok := token.UnsignedConstant16(); ok {
			t.Error("-0x50($r5) is not a UnsignedConstant16")
		}
		if ref, ok := token.MemoryReference(); !ok {
			t.Error("-0x50($r5) is a MemoryReference")
		} else if ref.Offset != -0x50 {
			t.Error("invalid offset for -0x50($r5):", ref.Offset)
		} else if ref.Register != 5 {
			t.Error("invalid register for -0x50($r5):", ref.Offset)
		}
	}

	token, err = ParseArgToken("($r31)")
	if err != nil {
		t.Error(err)
	} else {
		if _, ok := token.Register(); ok {
			t.Error("($31) is not a register")
		}
		if _, ok := token.AbsoluteCodePointer(); ok {
			t.Error("($r31) is not an absolute code pointer")
		}
		if _, ok := token.RelativeCodePointer(); ok {
			t.Error("($r31) is not a relative code pointer")
		}
		if _, ok := token.Constant5(); ok {
			t.Error("($r31) is not a Constant5")
		}
		if _, ok := token.SignedConstant16(); ok {
			t.Error("($r31) is not a SignedConstant16")
		}
		if _, ok := token.UnsignedConstant16(); ok {
			t.Error("($r31) is not a UnsignedConstant16")
		}
		if ref, ok := token.MemoryReference(); !ok {
			t.Error("($r31) is a MemoryReference")
		} else if ref.Offset != 0 {
			t.Error("invalid offset for ($r31):", ref.Offset)
		} else if ref.Register != 31 {
			t.Error("invalid register for ($r31):", ref.Offset)
		}
	}

	validTokens := []string{"$r0", "$r31", "$r15", "0x7fff($r1)", "-0x8000($r1)"}
	for _, tok := range validTokens {
		if _, err := ParseArgToken(tok); err != nil {
			t.Error("failed to parse "+tok+":", err)
		}
	}

	badTokens := []string{"Monkey Brain", "$r32", "$r-1", "$32", "0x8000($r1)", "-0x8001($r1)"}
	for _, tok := range badTokens {
		if _, err := ParseArgToken(tok); err == nil {
			t.Error("parsed invalid token:", tok)
		}
	}
}

func TestRegisterNameParsing(t *testing.T) {
	mapping := map[string]int{
		"$zero": 0,
		"$at":   1,
		"$gp":   28,
		"$sp":   29,
		"$fp":   30,
		"$ra":   31,
		"$v0":   2,
		"$v1":   3,
		"$t8":   24,
		"$t9":   25,
		"$s8":   30,
		"$k0":   26,
		"$k1":   27,
	}
	for i := 0; i < 32; i++ {
		iName := strconv.Itoa(i)
		mapping["$"+iName] = i
		mapping["$r"+iName] = i
	}
	for i := 0; i <= 7; i++ {
		name := "$t" + strconv.Itoa(i)
		mapping[name] = i + 8
	}
	for i := 0; i <= 7; i++ {
		name := "$s" + strconv.Itoa(i)
		mapping[name] = i + 16
	}
	for i := 0; i <= 3; i++ {
		name := "$a" + strconv.Itoa(i)
		mapping[name] = i + 4
	}
	for name, expected := range mapping {
		if tok, err := ParseArgToken(name); err != nil {
			t.Error(name, err)
		} else if idx, ok := tok.Register(); !ok || idx != expected {
			t.Error(name, idx, ok)
		}
	}
}

func BenchmarkParseArgTokenReg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseArgToken("$r15")
	}
}

func BenchmarkParseArgTokenMem(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseArgToken("0x15($r15)")
	}
}
