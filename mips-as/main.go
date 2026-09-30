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
	flag.BoolVar(&littleEndian, "little", false, "encode instructions as little endian")

	flag.Parse()
	if len(flag.Args()) != 2 {
		dieUsage()
	}

	inFile := flag.Args()[0]
	outFile := flag.Args()[1]

	source, err := ioutil.ReadFile(inFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tokenized, err := mips32.TokenizeSource(string(source))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	executable, err := mips32.ParseExecutable(tokenized)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	output, err := os.Create(outFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer output.Close()

	for addr := uint32(0); addr < executable.End(); addr += 4 {
		inst := executable.Get(addr)
		if inst == nil {
			output.Write([]byte{0, 0, 0, 0})
			continue
		}
		enc, err := inst.Encode(addr, executable.Symbols)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if littleEndian {
			output.Write([]byte{byte(enc), byte(enc >> 8), byte(enc >> 16), byte(enc >> 24)})
		} else {
			output.Write([]byte{byte(enc >> 24), byte(enc >> 16), byte(enc >> 8), byte(enc)})
		}
	}
}

func dieUsage() {
	fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "[flags] <in.s> <out.bin>")
	flag.PrintDefaults()
	os.Exit(1)
}
