package mips32

import "testing"

func TestEmulatorRegModifiers(t *testing.T) {
	code := `
		# Seed the program with two random numbers.
		LUI $1, 0xca6d
		ORI $1, $1, 0x8c46       # $r1 = 0xca6d8c46
		LUI $2, 0x0a93
		ORI $2, $2, 0xd70b       # $r2 = 0x0a93d70b

		ADDIU $3, $1, 0x1337
		ADDIU $3, $3, -0x10      # $r3 = 0xCA6D9F6D
		ADDU $4, $1, $2          # $r4 = 0xD5016351
		AND $5, $1, $2           # $r5 = 0xA018402
		ANDI $30, $1, 0xd70b     # $r30 = 0x8402
		MOVN $6, $1, $2          # $r6 = 0xca6d8c46
		MOVN $7, $1, $7          # $r7 = 0
		MOVZ $8, $1, $2          # $r8 = 0
		MOVZ $9, $1, $9          # $r9 = 0xca6d8c46
		NOR $10, $1, $2          # $r10 = 0x350020b0
		OR $11, $1, $2           # $r11 = 0xCAFFDF4F
		ORI $12, $1, 0xd70b      # $r12 = 0xCA6DDF4F
		SLL $13, $1, 11          # $r13 = 0x6c623000
		SLLV $14, $1, $2         # $r14 = 0x6c623000
		SLT $15, $1, $2          # $r15 = 1
		SLT $16, $2, $1          # $r16 = 0
		SLTU $17, $1, $2         # $r17 = 0
		SLTU $18, $2, $1         # $r18 = 1
		SLTI $19, $1, -1         # $r19 = 1
		SLTI $20, $2, -1         # $r20 = 0
		SLTIU $21, $1, -1        # $r21 = 1
		SLTIU $22, $2, -1        # $r22 = 1
		SRA $23, $1, 11          # $r23 = 0xFFF94DB1
		SRAV $24, $1, $2         # $r24 = 0xFFF94DB1
		SRL $25, $1, 11          # $r25 = 0x194DB1
		SRLV $26, $1, $2         # $r26 = 0x194DB1
		SUBU $27, $1, $2         # $r27 = 0xBFD9B53B
		XOR $28, $1, $2          # $r28 = 0xC0FE5B4D
		XORI $29, $1, 0xffff     # $r29 = 0xca6d73b9
	`
	emulator, err := runTestProgram(code)
	if err != nil {
		t.Fatal(err)
	}
	regFile := RegisterFile{0, 0xca6d8c46, 0x0a93d70b, 0xCA6D9F6D, 0xD5016351, 0xA018402,
		0xca6d8c46, 0, 0, 0xca6d8c46, 0x350020b0, 0xCAFFDF4F, 0xCA6DDF4F, 0x6c623000, 0x6c623000, 1,
		0, 0, 1, 1, 0, 1, 1, 0xFFF94DB1, 0xFFF94DB1, 0x194DB1, 0x194DB1, 0xBFD9B53B, 0xC0FE5B4D,
		0xca6d73b9, 0x8402, 0}
	for i := 0; i < 32; i++ {
		if regFile[i] != emulator.RegisterFile[i] {
			t.Error("bad register", i, "-", emulator.RegisterFile[i])
		}
	}
}

func TestEmulatorJumps(t *testing.T) {
	code := `
		# Seed the program with two random numbers.
		LUI $1, 0xca6d
		ORI $1, $1, 0x8c46       # $r1 = 0xca6d8c46
		LUI $2, 0x0a93
		ORI $2, $2, 0xd70b       # $r2 = 0x0a93d70b

		J SYM1
		NOP
		ADDU $3, $1, $2

		SYM1:
		JAL SYM2                 # $r31 = 36
		XOR $4, $1, $2           # $r4 = 0xC0FE5B4D
		NOR $5, $1, $2

		SYM2:
		ORI $6, $0, 56           # $r6 = 56
		JR $6
		AND $7, $1, $2           # $r7 = 0xA018402
		OR $8, $1, $2

		SYM3: # 56
		ORI $9, $0, 72           # $r9 = 72
		JALR $10, $9             # $r10 = 68
		AND $11, $1, $2          # $r11 = 0xA018402
		OR $12, $1, $2

		SYM4: # 72
		LUI $13, 0xDEAD
		ORI $13, $13, 0xBEEC     # $r13 = 0xDEADBEEC
		JR $13
		SLLV $14, $1, $2         # $r14 = 0x6c623000
		SRLV $15, $1, $2

		.text 0xDEADBEE8
		LUI $16, 0x1337
		SYM5: # 0xDEADBEEC
		LUI $17, 1337            # $r17 = 0x05390000
	`
	emulator, err := runTestProgram(code)
	if err != nil {
		t.Fatal(err)
	}
	regFile := RegisterFile{1: 0xca6d8c46, 2: 0x0a93d70b, 4: 0xC0FE5B4D, 6: 56, 7: 0xA018402, 9: 72,
		10: 68, 11: 0xA018402, 13: 0xDEADBEEC, 14: 0x6c623000, 17: 0x05390000, 31: 36}
	for i := 0; i < 32; i++ {
		if regFile[i] != emulator.RegisterFile[i] {
			t.Error("bad register", i, "-", emulator.RegisterFile[i])
		}
	}
}

func TestEmulatorBranches(t *testing.T) {
	code := `
		ORI $3, $0, 0x8c46
		ORI $4, $0, 0x8c10
		BEQLOOP:
		ADDIU $4, $4, 1
		ADDIU $5, $5, 1
		BEQ $3, $4, BEQLOOPEND
		NOP
		J BEQLOOP
		BEQLOOPEND:

		# $r3 = $r4 = 0x8c46
		# $r5 = 0x36

		ORI $6, $0, 0x8c47
		ORI $7, $0, 0x8c10
		BNELOOP:
		ADDIU $6, $6, -1
		BNE $6, $7, -8
		ADDIU $8, $8, 1

		# $r6 = $r7 = 0x8c10
		# $r8 = 0x37

		ORI $9, $0, 10
		BGEZLOOP:
		ADDIU $9, $9, -1
		BGEZ $9, BGEZLOOP
		ADDIU $10, $10, 1

		# $r9 = -1
		# $r10 = 11

		ORI $11, $0, 10
		BGTZLOOP:
		ADDIU $11, $11, -1
		BGTZ $11, BGTZLOOP
		ADDIU $12, $12, 1

		# $r11 = 0
		# $r12 = 10

		ADDIU $13, $13, -10
		BLEZLOOP:
		ADDIU $13, $13, 1
		BLEZ $13, -8
		ADDIU $14, $14, 1

		# $r13 = 1
		# $r14 = 10

		ADDIU $15, $0, -10
		BLTZLOOP:
		ADDIU $15, $15, 1
		BLTZ $15, BLTZLOOP
		ADDIU $16, $16, 1

		# $r15 = 0
		# $r16 = 9
	`
	emulator, err := runTestProgram(code)
	if err != nil {
		t.Fatal(err)
	}
	regFile := RegisterFile{3: 0x8c46, 4: 0x8c46, 5: 0x36, 6: 0x8c10, 7: 0x8c10, 8: 0x37,
		9: 0xffffffff, 10: 11, 11: 0, 12: 10, 13: 1, 14: 11, 15: 0, 16: 10}
	for i := 0; i < 32; i++ {
		if regFile[i] != emulator.RegisterFile[i] {
			t.Error("bad register", i, "-", emulator.RegisterFile[i])
		}
	}
}

func TestEmulatorMemory(t *testing.T) {
	code := `
		# Seed the program with two random numbers.
		LUI $1, 0xca46
		ORI $1, $1, 0x8c6d       # $r1 = 0xca468c6d
		LUI $2, 0x0a0b
		ORI $2, $2, 0xd793       # $r2 = 0x0a0bd793

		SB $1, ($0)
		SW $2, 4($0)
		SB $2, ($1)
		SW $1, 1($2)

		LB $3, ($0)              # $r3 = 0x6d
		LW $4, 4($0)             # $r4 = 0x0a0bd793
		LB $5, ($1)              # $r5 = 0xffffff93
		LW $6, 1($2)             # $r6 = 0xca468c6d
		LBU $7, ($1)             # $r7 = 0x93
		LBU $8, ($0)             # $r8 = 0x6d
		LBU $9, 4($0)            # $r9 = {BE: 0x0a, LE: 0x93}
		LBU $10, 5($0)           # $r10 = {BE: 0x0b, LE: 0xd7}
		LBU $11, 6($0)           # $r11 = {BE: 0xd7, LE: 0x0b}
		LBU $12, 7($0)           # $r12 = {BE: 0x93, LE: 0x0a}
	`
	results := map[bool]RegisterFile{
		true: RegisterFile{1: 0xca468c6d, 2: 0x0a0bd793, 3: 0x6d, 4: 0x0a0bd793, 5: 0xffffff93,
			6: 0xca468c6d, 7: 0x93, 8: 0x6d, 9: 0x93, 10: 0xd7, 11: 0x0b, 12: 0x0a},
		false: RegisterFile{1: 0xca468c6d, 2: 0x0a0bd793, 3: 0x6d, 4: 0x0a0bd793, 5: 0xffffff93,
			6: 0xca468c6d, 7: 0x93, 8: 0x6d, 9: 0x0a, 10: 0x0b, 11: 0xd7, 12: 0x93},
	}
	for _, littleEndian := range []bool{false, true} {
		emulator, err := runTestProgramEndianness(code, littleEndian)
		if err != nil {
			t.Error(littleEndian, "-", err)
			continue
		}
		regFile := results[littleEndian]
		for i := 0; i < 32; i++ {
			if regFile[i] != emulator.RegisterFile[i] {
				t.Error(littleEndian, "- bad register", i, "-", emulator.RegisterFile[i])
			}
		}
	}
}

func TestEmulatorErrors(t *testing.T) {
	programs := []string{
		"ORI $r1, $r0, 3\nJR $r1",
		"J SYM\nJ SYM1\nNOP\nSYM:\nSYM1:",
		"ORI $r1, $r0, 3\nSW $r1, ($r1)",
		"ORI $r1, $r0, 3\nSW $r1, 2($r0)",
	}
ProgramLoop:
	for i, code := range programs {
		lines, err := TokenizeSource(code)
		if err != nil {
			t.Error(i, "-", err)
			continue
		}
		program, err := ParseExecutable(lines)
		if err != nil {
			t.Error(i, "-", err)
			continue
		}
		memory := NewLazyMemory()
		emulator := &Emulator{
			Memory:            memory,
			Executable:        program,
			ForceMemAlignment: true,
		}
		for !emulator.Done() {
			if err := emulator.Step(); err != nil {
				continue ProgramLoop
			}
		}
		t.Error("program", i, "did not fail")
	}
}

func TestEmulatorWriteZero(t *testing.T) {
	code := "NOP\nORI $1, $0, 0x10\nJALR $0, $1\nNOP\n" +
		"SW $1, ($0)\nLW $0, ($0)"
	lines, err := TokenizeSource(code)
	if err != nil {
		t.Fatal(err)
	}
	program, err := ParseExecutable(lines)
	if err != nil {
		t.Fatal(err)
	}
	memory := NewLazyMemory()
	emulator := &Emulator{
		Memory:            memory,
		Executable:        program,
		ForceMemAlignment: true,
	}
	for !emulator.Done() {
		if err := emulator.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if emulator.RegisterFile[0] != 0 {
		t.Error("wrote to register 0.")
	}
}

func runTestProgram(code string) (*Emulator, error) {
	return runTestProgramEndianness(code, false)
}

func runTestProgramEndianness(code string, little bool) (*Emulator, error) {
	lines, err := TokenizeSource(code)
	if err != nil {
		return nil, err
	}
	program, err := ParseExecutable(lines)
	if err != nil {
		return nil, err
	}
	memory := NewLazyMemory()
	emulator := &Emulator{
		Memory:       memory,
		Executable:   program,
		LittleEndian: little,
	}
	for !emulator.Done() {
		if err := emulator.Step(); err != nil {
			return nil, err
		}
	}
	return emulator, nil
}
