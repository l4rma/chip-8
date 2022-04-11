package interpreter

import (
	"bytes"
	"fmt"
	"io"
)

type chip8 struct {
	memory     [0x1000]byte  // 4096 bytes internal memory
	V          [0x10]byte    // 16 8-bit virtual registers (V0-VF)
	I          uint16        // Address register
	PC         uint16        // Program Counter (starts at 0x200)
	SP         byte          // Stack Pointer
	stack      [16]uint16    // 16 cells of reserved memory
	display    [64 * 32]bool // 64x32 pixel display
	delayTimer byte
	soundTimer byte
}

func NewChip8() chip8 {
	return chip8{}
}

func (c *chip8) LoadRom(data io.Reader) (int, error) {
	offset := 0x200
	return c.load(offset, data)
}

func (c *chip8) load(offset int, r io.Reader) (int, error) {
	return r.Read(c.memory[offset:])
}

func (c *chip8) LoadBytes(o int, b []byte) (int, error) {
	return c.load(o, bytes.NewReader(b))
}

func (c *chip8) PrintMemory(index int) {
	fmt.Printf("CHIP-8 Memory[%02X]: 0x%04X\n", index, c.memory[index])
}

func (c *chip8) clearDisplay() {
	for i := range c.display {
		c.display[i] = false
	}
}

func (c *chip8) Dispatch(op uint16) error {
	switch op & 0xF000 {
	case 0x0000:
		switch op {
		// 00E0 - CLS
		case 0x00E0:
			c.clearDisplay()
			c.PC += 2
			break
		case 0x00EE:
			// TODO
			break
		default:
			return fmt.Errorf("Unknown opcode: 0x%04X", op)
		}
	}
	return nil
}
