package mips32

import (
	"fmt"
	"testing"
)

func TestParseExecutableSuccess(t *testing.T) {
	exc := `
        SLLV $r5, $r6, $r7
        .text 0x50
        ADDIU $r5, $r6, -17
        MONKEY:
        .word 0xf2345678
        .text 0x20
        MONKEY1:
        LUI $r5, 0xBEEF
        ORI $r5, $r5, 0xDEAD
        .text 0x4
        SRLV $r6, $r5, $r7
        .text 0x100
        DANGLER:
    `
	lines, err := TokenizeSource(exc)
	if err != nil {
		t.Fatal(err)
	}
	executable, err := ParseExecutable(lines)
	if err != nil {
		t.Fatal(err)
	}
	expected := &Executable{
		Segments: map[uint32][]Instruction{
			0: []Instruction{
				{Name: "SLLV", Registers: []int{5, 6, 7}},
				{Name: "SRLV", Registers: []int{6, 5, 7}},
			},
			0x50: []Instruction{
				{Name: "ADDIU", Registers: []int{5, 6}, SignedConstant16: -17},
				{Name: ".word", RawWord: 0xf2345678},
			},
			0x20: []Instruction{
				{Name: "LUI", Registers: []int{5}, UnsignedConstant16: 0xBEEF},
				{Name: "ORI", Registers: []int{5, 5}, UnsignedConstant16: 0xDEAD},
			},
		},
		Symbols: map[string]uint32{
			"MONKEY":  0x54,
			"MONKEY1": 0x20,
			"DANGLER": 0x100,
		},
	}
	if len(expected.Segments) != len(executable.Segments) {
		t.Error("unexpected number of segments:", len(executable.Segments))
	}
	if len(expected.Symbols) != len(executable.Symbols) {
		t.Error("unexpected number of symbols:", len(executable.Symbols))
	}
	for sym, val := range expected.Symbols {
		if executable.Symbols[sym] != val {
			t.Error("unexpected symbol for name "+sym+":", executable.Symbols[sym])
		}
	}
	for seg, insts := range expected.Segments {
		actualInsts, ok := executable.Segments[seg]
		if !ok {
			t.Error("missing segment:", seg)
			continue
		}
		if len(insts) != len(actualInsts) {
			t.Error("invalid number of instructions for segment", seg, "-", len(actualInsts))
			continue
		}
		for i, inst := range insts {
			if !instructionsEquivalent(&inst, &actualInsts[i]) {
				t.Error("bad instruction", i, "in segment", seg, "-", actualInsts[i])
			}
		}
	}
}

func TestParseExecutableFailure(t *testing.T) {
	failures := []string{
		"NOP\n.text 0\nNOP",
		"NOP\n.text 0x0\nSUBU $a0, $a1, $a2",
		"FOO:\nNOP\nFOO:\nNOP",
		"ORI $r1, 5",
	}
	for _, failure := range failures {
		lines, err := TokenizeSource(failure)
		if err != nil {
			t.Error(err)
			continue
		}
		if _, err := ParseExecutable(lines); err == nil {
			t.Error("expected error for:", failure)
		}
	}
}

func TestExecutableRender(t *testing.T) {
	programs := []string{
		`
            FOO:
            .text 0x20
            MOVN $r3, $r2, $r1
            NOP
            BAZ:
            .text 0x30
            BAR:
            ADDU $r7, $r15, $r31
            BUZ:
            SLLV $r5, $r6, $r7
            .text 0x50
            SYM:
            .text 0x70
            SUM:
        `,
		`
            FOO:
            NOP
            MOVN $r3, $r2, $r1
            END:
        `,
	}
	for _, program := range programs {
		lines, err := TokenizeSource(program)
		if err != nil {
			t.Error(err)
			continue
		}
		exec, err := ParseExecutable(lines)
		if err != nil {
			t.Error(err)
			continue
		}
		rendered, err := exec.Render()
		if err != nil {
			t.Error(err)
			continue
		}
		if len(rendered) != len(lines) {
			t.Error("invalid rendering size:", len(rendered))
			for i, x := range rendered {
				fmt.Println(i, x.Directive, x.Instruction, x.SymbolMarker)
			}
			continue
		}
		for i, line := range rendered {
			line.LineNumber = lines[i].LineNumber
			if !line.Equal(&lines[i]) {
				t.Error("invalid line", i, "-", line)
			}
		}
	}
}
