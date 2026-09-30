package mips32

import "testing"

func TestTokenizeSource(t *testing.T) {
	source := `.text 0x50000 # this says where our program's data is located.
    FOOBAR:
    LUI $r5, 0xDEAD
    ORI $r5, $r5, 0xBEEF # 0xDEADBEEF is a cool hex string.

    SW $r5, 0x1337($t0)

    # the next line is a NOP
    .word 0x00000000

    J FOOBAR
    NOP
    `
	tokenized, err := TokenizeSource(source)
	if err != nil {
		t.Error(err)
	} else {
		expected := []TokenizedLine{
			{
				LineNumber: 1,
				Comment:    createStringPtr(" this says where our program's data is located."),
				Directive:  &TokenizedDirective{"text", 0x50000},
			},
			{
				LineNumber:   2,
				SymbolMarker: createStringPtr("FOOBAR"),
			},
			{
				LineNumber: 3,
				Instruction: &TokenizedInstruction{
					Name: "LUI",
					Arguments: []*ArgToken{
						&ArgToken{isRegister: true, register: 5},
						&ArgToken{isConstant: true, constant: 0xDEAD},
					},
				},
			},
			{
				LineNumber: 4,
				Comment:    createStringPtr(" 0xDEADBEEF is a cool hex string."),
				Instruction: &TokenizedInstruction{
					Name: "ORI",
					Arguments: []*ArgToken{
						&ArgToken{isRegister: true, register: 5},
						&ArgToken{isRegister: true, register: 5},
						&ArgToken{isConstant: true, constant: 0xBEEF},
					},
				},
			},
			{
				LineNumber: 6,
				Instruction: &TokenizedInstruction{
					Name: "SW",
					Arguments: []*ArgToken{
						&ArgToken{isRegister: true, register: 5},
						&ArgToken{isMemory: true, memOffset: 0x1337, memRegister: 8},
					},
				},
			},
			{
				LineNumber: 8,
				Comment:    createStringPtr(" the next line is a NOP"),
			},
			{
				LineNumber: 9,
				Directive:  &TokenizedDirective{"word", 0},
			},
			{
				LineNumber: 11,
				Instruction: &TokenizedInstruction{
					Name: "J",
					Arguments: []*ArgToken{
						&ArgToken{isSymbol: true, symbol: "FOOBAR"},
					},
				},
			},
			{
				LineNumber: 12,
				Instruction: &TokenizedInstruction{
					Name:      "NOP",
					Arguments: []*ArgToken{},
				},
			},
		}
		if len(expected) != len(tokenized) {
			t.Error("invalid tokenized program:", tokenized)
		} else {
			for i, line := range tokenized {
				if !line.Equal(&expected[i]) {
					t.Error("invalid line", expected[i].LineNumber, ":", line)
				}
			}
		}
	}

	invalidStrs := []string{"LUI $r5 0xDEAD", ".text foo", "Monkey Brains:", "foo_bar $r5",
		"$r5, $r4"}
	for _, str := range invalidStrs {
		if _, err := TokenizeSource(str); err == nil {
			t.Error("expected parse to fail:", str)
		}
	}
}

func BenchmarkTokenizeSource(b *testing.B) {
	code := `
		.text 0x50000 # this says where our program's data is located.
		FOOBAR:
		LUI $r5, 0xDEAD
		ORI $r5, $r5, 0xBEEF # 0xDEADBEEF is a cool hex string.

		SW $r5, 0x1337($t0)

		# the next line is a NOP
		.word 0x00000000

		J FOOBAR
		NOP
	`
	for i := 0; i < b.N; i++ {
		TokenizeSource(code)
	}
}

func createStringPtr(s string) *string {
	return &s
}
