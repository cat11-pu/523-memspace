package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/unixpickle/mips32"
)

const MemoryDumpColumns = 16

func main() {
	var littleEndian bool
	flag.BoolVar(&littleEndian, "little", false, "use little endian memory")

	var relaxAlignment bool
	flag.BoolVar(&relaxAlignment, "misaligned", false, "allow misaligned memory access")

	var memoryDumpSize uint64
	flag.Uint64Var(&memoryDumpSize, "dumpsize", 0, "size (in bytes) for memory dump")

	var memoryDumpStart uint64
	flag.Uint64Var(&memoryDumpStart, "dumpstart", 0, "base address for memory dump")

	flag.Parse()
	if len(flag.Args()) != 1 {
		dieUsage()
	}

	file := flag.Args()[0]
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tokens, err := mips32.TokenizeSource(string(contents))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	exc, err := mips32.ParseExecutable(tokens)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	emu := &mips32.Emulator{
		Memory:            mips32.NewLazyMemory(),
		Executable:        exc,
		LittleEndian:      littleEndian,
		ForceMemAlignment: !relaxAlignment,
	}
	for !emu.Done() {
		if err := emu.Step(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	fmt.Println("Register file:")
	fmt.Println(emu.RegisterFile.String())

	if memoryDumpSize > 0 {
		dumpMemory(emu.Memory, uint32(memoryDumpStart), uint32(memoryDumpSize))
	}
}

func dieUsage() {
	fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "[flags] <file.s>")
	flag.PrintDefaults()
	os.Exit(1)
}

func dumpMemory(mem mips32.Memory, start, size uint32) {
	fmt.Println("Memory dump:")

	// The uint64 conversions deal with the case when start+size would overflow.
	column := 0
	for i := uint64(start); i < uint64(start)+uint64(size); i++ {
		b := mem.Get(uint32(i))
		hexValue := strconv.FormatUint(uint64(b), 16)
		for len(hexValue) < 2 {
			hexValue = "0" + hexValue
		}
		if column == 0 {
			hexAddr := strconv.FormatUint(i, 16)
			for len(hexAddr) < 8 {
				hexAddr = "0" + hexAddr
			}
			fmt.Print(hexAddr + "  ")
		}
		if column == MemoryDumpColumns-1 {
			fmt.Println(hexValue)
			column = 0
		} else {
			fmt.Print(hexValue + " ")
			column++
		}
	}
	if column != 0 {
		fmt.Println("")
	}
}
