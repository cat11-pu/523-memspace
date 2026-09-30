# Overview

The root of this repository contains a Go package for manipulating, running, assembling, and disassembling MIPS code. This library, called **mips32**, is useful for random code generation, systematic code manipulation, and much more.

This repository also contains various tools that depend on the **mips32** package. These tools are as follows:

 * mips-run - run MIPS programs from the command line and see their resulting registers.
 * mips-as - assembly a MIPS program to binary
 * mips-disas - disassemble MIPS binary into MIPS assembly code.

# Usage

You must have the Go programming language installed and configured. This includes having `git` installed, a `GOPATH` configured, etc. It is also highly recommended that you add `$GOPATH/bin` to your `PATH` environment variable.

Once you have Go all setup, run the following command:

    $ go get github.com/unixpickle/mips32

The above command downloads the **mips32** package. To install its subcommands, do the following:

    $ go install github.com/unixpickle/mips32/mips-run
    $ go install github.com/unixpickle/mips32/mips-as
    $ go install github.com/unixpickle/mips32/mips-disas

Assuming you added `$GOPATH/bin` to your `PATH`, you should now be able to run these tools from the command line. For example:

    $ echo "LUI \$r2, 0x1337" >file.s
    $ mips-run file.s
    Register file:
    r0  = 0x00000000  r16 = 0x00000000
    r1  = 0x00000000  r17 = 0x00000000
    r2  = 0x13370000  r18 = 0x00000000
    r3  = 0x00000000  r19 = 0x00000000
    ...

# Supported instructions

This supports the following instructions:

 * NOP - do nothing
 * ADDIU - add a register to an immediate
 * ADDU - add two registers
 * AND - AND two registers
 * ANDI - AND a register and an immediate
 * BEQ - branch if two registers are equal
 * BGEZ - branch if a register is greater than or equal to zero
 * BGTZ - branch if a register is greater than zero
 * BLEZ - branch if a register is less than or equal to zero
 * BLTZ - branch if a register is less than zero
 * BNE - branch if two registers are not equal
 * J - jump to a symbol or hard-coded address
 * JAL - jump to a symbol or a hard-coded address, saving PC+8 in $r31
 * JALR - jump to a register, saving PC+8 to $r31 or an (optional) destination register
 * JR - jump to a register
 * LB - load a signed byte from memory
 * LBU - load an unsigned byte from memory
 * LW - load a word from memory
 * SB - store a byte to memory
 * SW - store a word to memory
 * LUI - set a register to an immediate, shifted left by 16 bits
 * MOVN - move one register into another if a third register is non-zero
 * MOVZ - move one register into another if a third register is zero
 * NOR - OR two registers, then negate the result
 * OR - OR two registers
 * ORI - OR a register and an immediate
 * SLL - shift left logical by a constant amount
 * SLLV - shift left logical by a variable amount
 * SLT - set a register to 1 or 0 depending on if another register is less than yet another one
 * SLTI - set a register to 1 or 0 depending on if another register is less than a signed immediate
 * SLTIU - set a register to 1 or 0 depending on if another register is less than a signed immediate
 * SLTU - set a register to 1 or 0 depending on if another register is less than yet another one
 * SRA - shift right arithmetic by a constant amount
 * SRAV - shift right arithmetic by a variable amount
 * SRL - shift right logical by a constant amount
 * SRLV - shift right logical by a variable amount
 * SUBU - subtract a register from another register
 * XOR - XOR one register with another one
 * XORI - XOR a register with an immediate

# Directives

You can use the `.text` directive to place code at an arbitrary address (which must be aligned by 4). For example, see this program:

```assembly
J 0x1338
NOP

.text 0x1338
ADDIU $r5, $r4, 5
```

You can use the `.word` directive to insert a raw 32-bit value into the program. For example, the above program could be converted to:

```assembly
.word 0x080004ce
.word 0

.text 0x1338
.word 0x24850005
```

# Memory

By default, word-based memory operations are big endian. If you wish to make them little endian, you can pass a `-little` flag to the `mips-run` program.

The emulator uses a lazy memory implementation, so you can access distant regions of memory without consuming too much of the host system's memory. This is good for emulating systems with 4GB of RAM when the host system doesn't have 4GB of RAM to spare.
