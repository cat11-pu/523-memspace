package mips32

import "testing"

func TestParseTokenizedInstruction(t *testing.T) {
	cases := map[string]*Instruction{
		"ADDIU $r5, $r6, -17": &Instruction{
			Name:             "ADDIU",
			Registers:        []int{5, 6},
			SignedConstant16: -17,
		},
		"J FOO": &Instruction{
			Name: "J",
			CodePointer: CodePointer{
				Absolute: true,
				IsSymbol: true,
				Symbol:   "FOO",
			},
		},
	}
	for code, expected := range cases {
		tokenized, err := TokenizeSource(code)
		if err != nil {
			t.Error(code, "-", err)
			continue
		} else if len(tokenized) != 1 || tokenized[0].Instruction == nil {
			t.Error(code, "- bad tokens")
			continue
		}
		inst, err := ParseTokenizedInstruction(tokenized[0].Instruction)
		if err != nil {
			t.Error(code, "-", err)
			continue
		}
		if !instructionsEquivalent(inst, expected) {
			t.Error(code, "-", inst)
		}
	}

	broken := []string{"ADDIU $r5, $r6, $7", "J $r7"}
	for _, brokenCode := range broken {
		tokenized, err := TokenizeSource(brokenCode)
		if err != nil {
			t.Error(brokenCode, "-", err)
			continue
		} else if len(tokenized) != 1 || tokenized[0].Instruction == nil {
			t.Error(brokenCode, "- bad tokens")
			continue
		}
		if _, err := ParseTokenizedInstruction(tokenized[0].Instruction); err == nil {
			t.Error("expected error for", brokenCode)
		}
	}
}

func TestInstructionRender(t *testing.T) {
	code := `
		NOP
		SLL $r5, $r2, 15
		SLLV $r5, $r6, $r7
		ADDIU $r5, $r6, -17
		LUI $r5, 0xBEEF
		ORI $r5, $r5, 0xDEAD
		SRLV $r6, $r5, $r7
		J FOOBAR
		BEQ $r5, $r31, TEST
		BEQ $r5, $r31, 0xf000
		JAL 0xDEADBEEF
		SB $r5, 15($r3)
	`
	lines, err := TokenizeSource(code)
	if err != nil {
		t.Fatal(err)
	}
	for i, line := range lines {
		if line.Instruction == nil {
			t.Error("missing instruction for line", i)
			continue
		}
		inst, err := ParseTokenizedInstruction(line.Instruction)
		if err != nil {
			t.Error(err)
			continue
		}
		rendered, err := inst.Render()
		if err != nil {
			t.Error(err)
			continue
		}
		rendered.LineNumber = line.LineNumber
		if !rendered.Equal(&line) {
			t.Error("failed to render line", line, "got", rendered)
		}
	}

	inst := &Instruction{Name: ".word", RawWord: 0xf2345678}
	if rendered, err := inst.Render(); err != nil {
		t.Error(err)
	} else if rendered.Directive == nil || rendered.Directive.Name != "word" ||
		rendered.Directive.Constant != 0xf2345678 {
		t.Error("bad result:", rendered.Directive)
	}
}
