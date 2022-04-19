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
	stack      [0x10]uint16  // 16 cells of reserved memory
	display    [64 * 32]bool // 64x32 pixel display
	delayTimer byte
	soundTimer byte
}

func NewChip8() chip8 {
	return chip8{
		PC: 0x200,
		SP: 0,
	}
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
	fmt.Printf("CHIP-8 Memory[%d]: 0x%02X\n", index, c.memory[index])
}

func (c *chip8) clearDisplay() {
	for i := range c.display {
		c.display[i] = false
	}
}

func (c *chip8) Init() error {
	// TODO: Implement function
	return nil
}

func (c *chip8) Run() error {
	err := c.Init()
	if err != nil {
		return err
	}

	for {
		err := c.Step()
		if err != nil {
			return err
		}
	}
}

func (c *chip8) Step() error {
	opcode := c.FetchInstruction()
	_, err := c.ExecuteOpcode(opcode)
	if err != nil {
		return err
	}
	return nil
}

func (c *chip8) Push(addr uint16) error {
	c.SP++
	c.stack[c.SP] = addr

	return nil
}

func (c *chip8) FetchInstruction() uint16 {
	opCode := uint16(c.memory[c.PC])<<8 | uint16(c.memory[c.PC+1])
	//c.PC += 2 // TODO: Deside to inc PC here or in each opcode execution
	return opCode
}

func (c *chip8) ExecuteOpcode(op uint16) (uint16, error) {
	switch op & 0xF000 {
	case 0x0000: // 0nnn
		switch op {
		case 0x00E0: // CLS
			// Clear the display
			c.clearDisplay()
			c.PC += 2
			break
		case 0x00EE: // RET
			// Return from a subroutine.
			// The interpreter sets the program counter to the address at the
			// top of the stack, then subtracts 1 from the stack pointer.
			c.PC = c.stack[c.SP]
			c.SP--
			c.PC += 2
			break
		default:
			return op, fmt.Errorf("Unknown opcode: 0x%04X", op)
		}
	case 0x1000: // 1nnn - JP addr
		// Jump to location nnn.
		// The interpreter sets the program counter to nnn.
		c.PC = op & 0x0FFF
		break
	case 0x2000: // 2nnn - CALL addr
		// Call subroutine at nnn.
		// The interpreter increments the stack pointer, then puts the current
		// PC on the top of the stack. The PC is then set to nnn.
		c.Push(c.PC)
		c.PC = op & 0x0FFF
		break
	case 0x3000: //3xkk - SE Vx, byte
		// Skip next instruction if Vx = kk.
		// The interpreter compares register Vx to kk, and if they are equal,
		// increments the program counter by 2.
		kk := byte(op)
		x := (op & 0x0F00) >> 8

		c.PC += 2

		if kk == c.V[x] {
			c.PC += 2
			break
		}
		break
	case 0x4000: // 4xkk - SNE Vx, byte
		// Skip next instruction if Vx != kk.
		// The interpreter compares register Vx to kk, and if they are not
		// equal, increments the program counter by 2.
		kk := byte(op)
		x := (op & 0x0F00) >> 8

		c.PC += 2

		if kk != c.V[x] {
			c.PC += 2
			break
		}
		break
	case 0x5000: // 5xy0 - SE Vx, Vy
		// 	Skip next instruction if Vx = Vy.
		// The interpreter compares register Vx to register Vy, and if they are equal, increments the program counter by 2.
		x := (op & 0x0F00) >> 8
		y := (op & 0x00F0) >> 4

		c.PC += 2

		if c.V[x] == c.V[y] {
			c.PC += 2
			break
		}
		break
	default:
		return op, fmt.Errorf("Unknown opcode: 0x%04X", op)
	}

	return op, nil
}
