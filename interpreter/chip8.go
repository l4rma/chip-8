package interpreter

import "io"

type chip8 struct {
	memory [4096]byte
}

func NewChip8() chip8 {
	return chip8{}
}

func (c *chip8) LoadRom(data io.Reader) (int, error) {
	offset := 0x200
	return data.Read(c.memory[offset:])
}
