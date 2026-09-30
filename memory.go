package mips32

// Memory defines an interface for storing binary data.
type Memory interface {
	Get(ptr uint32) byte
	Set(ptr uint32, b byte)
}

// LazyMemory is a dynamic, sparse Memory implementation.
// LazyMemory allows you to write to various parts of the address space without allocating a bunch
// of room inbetween sparse addresses.
type LazyMemory struct {
	pages map[uint32][]byte
}

func NewLazyMemory() *LazyMemory {
	return &LazyMemory{pages: map[uint32][]byte{}}
}

func (l *LazyMemory) Get(ptr uint32) byte {
	page := ptr & 0xfffff000
	if data := l.pages[page]; data != nil {
		return data[ptr&0xfff]
	} else {
		return 0
	}
}

func (l *LazyMemory) Set(ptr uint32, b byte) {
	page := ptr & 0xfffff000
	if data := l.pages[page]; data != nil {
		data[ptr&0xfff] = b
	} else {
		l.pages[page] = make([]byte, 0x1000)
		l.pages[page][ptr&0xfff] = b
	}
}
