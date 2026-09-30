package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/unixpickle/mips32"
)

func main() {
	var littleEndian bool
	flag.BoolVar(&littleEndian, "little", false, "decode instructions as little endian")

	flag.Parse()
	if len(flag.Args()) != 2 {
		dieUsage()
	}

	inFile := flag.Args()[0]
	outFile := flag.Args()[1]

	binary, err := ioutil.ReadFile(inFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else if len(binary)&3 != 0 {
		fmt.Fprintln(os.Stderr, "file is invalid length (must be multiple of 4)")
		os.Exit(1)
	}

	instructions := []*mips32.Instruction{}
	for i := 0; i < len(binary); i += 4 {
		var instNum uint32
		if littleEndian {
			instNum = uint32(binary[i]) | (uint32(binary[i+1]) << 8) | (uint32(binary[i+2]) << 16) |
				(uint32(binary[i+3]) << 24)
		} else {
			instNum = (uint32(binary[i]) << 24) | (uint32(binary[i+1]) << 16) |
				(uint32(binary[i+2]) << 8) | uint32(binary[i+3])
		}
		inst := mips32.DecodeInstruction(instNum)
		instructions = append(instructions, inst)
	}

	output, err := os.Create(outFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer output.Close()

	for _, inst := range instructions {
		rendering, err := inst.Render()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		output.WriteString(rendering.String())
		output.WriteString("\n")
	}
}

func dieUsage() {
	fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "[flags] <in.bin> <out.s>")
	flag.PrintDefaults()
	os.Exit(1)
}
