package mips32

import (
	"errors"
	"sort"
	"strconv"
)

// An Executable stores chunks of instructions (called segments) and a symbol table.
type Executable struct {
	// Segments maps chunks of instructions to various parts of the address space.
	Segments map[uint32][]Instruction

	// Symbols maps symbol names to their addresses.
	Symbols map[string]uint32
}

// ParseExecutable turns a tokenized source file into an executable blob.
//
// If the executable cannot be parsed for any reason, this will fail.
// Overlapping .text sections, invalid instructions, and repeated symbols will all cause errors.
func ParseExecutable(lines []TokenizedLine) (*Executable, error) {
	var segmentStart uint32
	var instructionAddr uint32
	res := &Executable{
		Segments: map[uint32][]Instruction{},
		Symbols:  map[string]uint32{},
	}
	for _, line := range lines {
		if line.Instruction != nil {
			parsed, err := ParseTokenizedInstruction(line.Instruction)
			if err != nil {
				return nil, errors.New("line " + strconv.Itoa(line.LineNumber) + ": " + err.Error())
			}
			if res.addressInUse(instructionAddr) {
				return nil, addressInUseError(line.LineNumber, instructionAddr)
			}
			res.Segments[segmentStart] = append(res.Segments[segmentStart], *parsed)
			instructionAddr += 4
		} else if line.Directive != nil {
			dir := line.Directive
			if dir.Name == "word" {
				if res.addressInUse(instructionAddr) {
					return nil, addressInUseError(line.LineNumber, instructionAddr)
				}
				nextInst := DecodeInstruction(dir.Constant)
				res.Segments[segmentStart] = append(res.Segments[segmentStart], *nextInst)
				instructionAddr += 4
			} else if dir.Name == "text" {
				if dir.Constant&3 != 0 {
					return nil, errors.New("line " + strconv.Itoa(line.LineNumber) +
						": misaligned segment")
				}
				segmentStart = dir.Constant
				instructionAddr = dir.Constant
			} else {
				return nil, errors.New("line " + strconv.Itoa(line.LineNumber) +
					": unknown directive: " + dir.Name)
			}
		} else if line.SymbolMarker != nil {
			sym := *line.SymbolMarker
			if _, ok := res.Symbols[sym]; ok {
				return nil, errors.New("line " + strconv.Itoa(line.LineNumber) +
					": repeated symbol declaration: " + sym)
			}
			res.Symbols[sym] = instructionAddr
		}
	}
	res.joinContiguousSegments()
	// TODO: make sure no jump offsets are invalid.
	return res, nil
}

// Render generates a tokenized source file that corresponds to the given executable.
// If any the instructions are invalid, this will return an error.
func (e *Executable) Render() (list []TokenizedLine, err error) {
	sortedSegments := e.sortedSegmentAddresses()
	sortedSymbols := e.sortedSymbolAddrPairs()

	if len(sortedSegments) == 0 {
		sortedSegments = uint32List{0}
	}

	var symbolIdx int
	var currentAddress uint32

	for _, segment := range sortedSegments {
		for symbolIdx < len(sortedSymbols) && sortedSymbols[symbolIdx].Address < segment {
			sym := sortedSymbols[symbolIdx]
			if sym.Address != currentAddress {
				currentAddress = sym.Address
				list = append(list, TokenizedLine{
					Directive: &TokenizedDirective{
						Name:     "text",
						Constant: sym.Address,
					},
				})
			}
			list = append(list, TokenizedLine{SymbolMarker: &sym.Symbol})
			symbolIdx++
		}
		if segment != 0 {
			list = append(list, TokenizedLine{
				Directive: &TokenizedDirective{
					Name:     "text",
					Constant: segment,
				},
			})
		}
		currentAddress = segment
		for _, inst := range e.Segments[segment] {
			for symbolIdx < len(sortedSymbols) &&
				sortedSymbols[symbolIdx].Address == currentAddress {
				sym := sortedSymbols[symbolIdx]
				list = append(list, TokenizedLine{SymbolMarker: &sym.Symbol})
				symbolIdx++
			}
			rendered, err := inst.Render()
			if err != nil {
				hexStr := "0x" + strconv.FormatUint(uint64(currentAddress), 16)
				return nil, errors.New("failed to render instruction at " + hexStr + ": " +
					err.Error())
			}
			list = append(list, *rendered)
			currentAddress += 4
		}
	}

	for symbolIdx < len(sortedSymbols) {
		sym := sortedSymbols[symbolIdx]
		if sym.Address != currentAddress {
			currentAddress = sym.Address
			list = append(list, TokenizedLine{
				Directive: &TokenizedDirective{
					Name:     "text",
					Constant: sym.Address,
				},
			})
		}
		list = append(list, TokenizedLine{SymbolMarker: &sym.Symbol})
		symbolIdx++
	}

	return
}

// End returns the pointer to the first byte that is completely past any instruction data.
// Once a program starts executing instructions at or past End(), no more instructions will be seen.
func (e *Executable) End() uint32 {
	var lastAddr uint32
	for segStart, insts := range e.Segments {
		end := segStart + uint32(len(insts)*4)
		if end > lastAddr {
			lastAddr = end
		}
	}
	return lastAddr
}

// Get returns the instruction at a given pointer, or nil if no instruction exists at that pointer.
func (e *Executable) Get(addr uint32) *Instruction {
	for segStart, insts := range e.Segments {
		end := segStart + uint32(len(insts)*4)
		if segStart <= addr && end > addr {
			return &insts[(addr-segStart)>>2]
		}
	}
	return nil
}

// addressInUse reports of a word-aligned address is being used by one of the segments.
func (e *Executable) addressInUse(addr uint32) bool {
	for segment, insts := range e.Segments {
		if segment <= addr && segment+uint32(len(insts)*4) > addr {
			return true
		}
	}
	return false
}

// joinContiguousSegments joins contiguous segments.
func (e *Executable) joinContiguousSegments() {
	l := e.sortedSegmentAddresses()
	for i := 0; i < len(l)-1; i++ {
		segStart := l[i]
		size := uint32(len(e.Segments[segStart]) * 4)
		if segStart+size == l[i+1] {
			nextInstructions := e.Segments[l[i+1]]
			e.Segments[segStart] = append(e.Segments[segStart], nextInstructions...)
			delete(e.Segments, l[i+1])
			copy(l[i:], l[i+1:])
			l = l[:len(l)-1]
			i--
		}
	}
}

func (e *Executable) sortedSegmentAddresses() uint32List {
	l := make(uint32List, 0, len(e.Segments))
	for seg := range e.Segments {
		l = append(l, seg)
	}
	sort.Sort(l)
	return l
}

func (e *Executable) sortedSymbolAddrPairs() symbolAddrPairList {
	l := make(symbolAddrPairList, 0, len(e.Symbols))
	for sym, addr := range e.Symbols {
		l = append(l, symbolAddrPair{Symbol: sym, Address: addr})
	}
	sort.Sort(l)
	return l
}

func addressInUseError(line int, addr uint32) error {
	hexStr := "0x" + strconv.FormatUint(uint64(addr), 16)
	return errors.New("line " + strconv.Itoa(line) + ": overwriting address " + hexStr)
}

type uint32List []uint32

func (u uint32List) Len() int {
	return len(u)
}

func (u uint32List) Less(i, j int) bool {
	return u[i] < u[j]
}

func (u uint32List) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

type symbolAddrPair struct {
	Symbol  string
	Address uint32
}

type symbolAddrPairList []symbolAddrPair

func (s symbolAddrPairList) Len() int {
	return len(s)
}

func (s symbolAddrPairList) Less(i, j int) bool {
	return s[i].Address < s[j].Address
}

func (s symbolAddrPairList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
