package interpreter

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"time"
)

var (
	GraphicsWidth  uint16 = 0
	GraphicsHeight uint16 = 0
	ClockSpeed            = time.Duration(60) // 60Hz
)

type chip8 struct {
	memory     [0x1000]byte         // 4096 bytes internal memory
	V          [0x10]byte           // 16 8-bit virtual registers (V0-VF)
	I          uint16               // Address register
	PC         uint16               // Program Counter (starts at 0x200)
	SP         byte                 // Stack Pointer
	stack      [0x10]uint16         // 16 cells of reserved memory
	display    [64 * 2][32 * 2]byte // 64x32 pixel display
	keypad     [16]byte             // Keypad with 16 keys
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

func (c *chip8) MemoryDump(opcode uint16) {
	log.Printf("=== MEMORY DUMP ===")
	var i int16
	for i = 0; i < 16; i++ {
		log.Printf("Register V[%d]: %02X", i, c.V[i])
	}
	log.Printf("Register I: %04X", c.I)
	log.Printf("Current opcode: %04X", opcode)
}

func (c *chip8) clearDisplay() {
	for i := range c.display {
		for j := range c.display[i] {
			c.display[i][j] = 0
		}
	}
}

func (c *chip8) loadKeys() {
	var i uint8
	for i = 0; i < 16; i++ {
		c.keypad[i] = 0x00
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
		time.Sleep(time.Second / ClockSpeed)
	}
}

func (c *chip8) Step() error {
	opcode := c.FetchInstruction()
	_, err := c.ExecuteOpcode(opcode)
	if err != nil {
		log.Printf("Exec opcode error: %s", err)
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
	log.Printf("%04X", op)
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
		// TODO: add default throwing an error if any of the last 4 bits are high
		x := (op & 0x0F00) >> 8
		y := (op & 0x00F0) >> 4

		c.PC += 2

		if c.V[x] == c.V[y] {
			c.PC += 2
			break
		}
		break
	case 0x6000: // 6xkk - LD Vx, byte
		// Set Vx = kk.
		// The interpreter puts the value kk into register Vx.
		kk := byte(op)
		x := (op & 0x0F00) >> 8

		c.V[x] = kk

		c.PC += 2
		break
	case 0x7000: // 7xkk - ADD Vx, byte
		// Set Vx = Vx + kk.
		// Adds the value kk to the value of register Vx, then stores the result in Vx.
		kk := byte(op)
		x := (op & 0x0F00) >> 8

		c.V[x] += kk

		c.PC += 2
		break
	case 0x8000: // 8xyn
		x := (op & 0x0F00) >> 8
		y := (op & 0x00F0) >> 4
		switch op & 0x000F {
		case 0x0000: // 8xy0 - LD Vx, Vy
			// Set Vx = Vy.
			// Stores the value of register Vy in register Vx.
			c.V[x] = c.V[y]
			c.PC += 2
			break
		case 0x0001: // 8xy1 - OR Vx, Vy
			// Set Vx = Vx OR Vy.
			// Performs a bitwise OR on the values of Vx and Vy, then stores
			// the result in Vx. A bitwise OR compares the corrseponding bits
			// from two values, and if either bit is 1, then the same bit in
			// the result is also 1. Otherwise, it is 0.
			c.V[x] |= c.V[y]

			c.PC += 2
			break
		case 0x0002: // 8xy2 - AND Vx, Vy
			// Set Vx = Vx AND Vy.
			// Performs a bitwise AND on the values of Vx and Vy, then stores
			// the result in Vx. A bitwise AND compares the corrseponding bits
			// from two values, and if both bits are 1, then the same bit in
			// the result is also 1. Otherwise, it is 0.
			c.V[x] &= c.V[y]

			c.PC += 2
			break
		case 0x0003: // 8xy3 - XOR Vx, Vy
			// Set Vx = Vx XOR Vy.
			// Performs a bitwise exclusive OR on the values of Vx and Vy, then
			// stores the result in Vx. An exclusive OR compares the
			// corrseponding bits from two values, and if the bits are not both
			// the same, then the corresponding bit in the result is set to 1.
			// Otherwise, it is 0.
			c.V[x] ^= c.V[y]

			c.PC += 2
			break
		case 0x0004: // 8xy4 - ADD Vx, Vy
			// Set Vx = Vx + Vy, set VF = carry.
			// The values of Vx and Vy are added together. If the result is
			// greater than 8 bits (i.e., > 255,) VF is set to 1, otherwise 0.
			// Only the lowest 8 bits of the result are kept, and stored in Vx.
			sum := uint16(c.V[x]) + uint16(c.V[y])

			c.V[x] += c.V[y]

			if sum > 0xFF {
				c.V[0xF] = 0x01
			} else {
				c.V[0xF] = 0x00
			}
			c.PC += 2
			break
		case 0x0005: // 8xy5 - SUB Vx, Vy
			// Set Vx = Vx - Vy, set VF = NOT borrow.
			// If Vx > Vy, then VF is set to 1, otherwise 0. Then Vy is
			// subtracted from Vx, and the results stored in Vx.
			if c.V[x] > c.V[y] {
				c.V[0xF] = 0x01
			} else {
				c.V[0xF] = 0x00
			}
			c.V[x] -= c.V[y]

			c.PC += 2
			break
		case 0x0006: // 8xy6 - SHR Vx {, Vy}
			// Set Vx = Vx SHR 1.
			// If the least-significant bit of Vx is 1, then VF is set to 1,
			// otherwise 0. Then Vx is divided by 2.
			if (c.V[x] & 0x1) == 0x01 {
				c.V[0xF] = 0x01
			} else {
				c.V[0xF] = 0x00
			}
			c.V[x] /= 2
			// c.V[x] = c.V[x] >> 1

			c.PC += 2
			break
		case 0x0007: // 8xy7 - SUBN Vx, Vy
			// Set Vx = Vy - Vx, set VF = NOT borrow.
			// If Vy > Vx, then VF is set to 1, otherwise 0. Then Vx is
			// subtracted from Vy, and the results stored in Vx.
			if c.V[y] > c.V[x] {
				c.V[0xF] = 0x01
			} else {
				c.V[0xF] = 0x00
			}
			c.V[x] = c.V[y] - c.V[x]

			c.PC += 2
			break
		case 0x000E: // 8xyE - SHL Vx {, Vy}
			// Set Vx = Vx SHL 1.
			// If the most-significant bit of Vx is 1, then VF is set to 1,
			// otherwise to 0. Then Vx is multiplied by 2.
			if (c.V[x] & 0x80) == 0x80 {
				c.V[0xF] = 0x01
			} else {
				c.V[0xF] = 0x00
			}
			c.V[x] = c.V[x] << 1

			c.PC += 2
			break
		}
	case 0x9000: // 9xy0 - SNE Vx, Vy
		// Skip next instruction if Vx != Vy.
		// The values of Vx and Vy are compared, and if they are not equal, the
		// program counter is increased by 2.
		x := (op & 0x0F00) >> 8
		y := (op & 0x00F0) >> 4

		if c.V[x] != c.V[y] {
			c.PC += 2
		}

		c.PC += 2
		break
	case 0xA000: // Annn - LD I, addr
		// Set I = nnn.
		// The value of register I is set to nnn.
		c.I = op & 0x0FFF
		c.PC += 2
		break
	case 0xB000: // Bnnn - JP V0, addr
		// Jump to location nnn + V0.
		// The program counter is set to nnn plus the value of V0.
		c.PC = (op & 0x0FFF) + uint16(c.V[0])
		break
	case 0xC000: // Cxkk - RND Vx, byte
		// Set Vx = random byte AND kk.
		// The interpreter generates a random number from 0 to 255, which is
		// then ANDed with the value kk. The results are stored in Vx.
		x := (op & 0x0F00) >> 8
		kk := byte(op)
		rnd := byte(rand.Intn(256))

		c.V[x] = rnd & kk

		c.PC += 2
		break
	case 0xD000: // Dxyn - DRW Vx, Vy, nibble
		// Display n-byte sprite starting at memory location I at (Vx, Vy), set
		// VF = collision.
		// The interpreter reads n bytes from memory, starting at the address
		// stored in I. These bytes are then displayed as sprites on screen at
		// coordinates (Vx, Vy). Sprites are XORed onto the existing screen. If
		// this causes any pixels to be erased, VF is set to 1, otherwise it is
		// set to 0. If the sprite is positioned so part of it is outside the
		// coordinates of the display, it wraps around to the opposite side of
		// the screen.
		x := (op & 0x0F00) >> 8
		y := (op & 0x00F0) >> 4
		n := (op & 0x000F)
		c.V[0xF] = 0
		j := uint16(0)
		i := uint16(0)

		for j = 0; j < n; j++ {
			//TODO: remove log
			//log.Printf("Opcode: %04X loop: %d", op, j)
			pixel := c.memory[c.I+j]
			for i = 0; i < 8; i++ {
				//log.Printf("Opcode: %04X inner loop: %d", op, i)
				if (pixel & (0x80 >> i)) != 0 {
					if c.display[(c.V[y] + uint8(j))][c.V[x]+uint8(i)] == 1 {
						c.V[0xF] = 1
					}
					c.display[(c.V[y] + uint8(j))][c.V[x]+uint8(i)] ^= 1
				}
			}
		}
		c.PC += 2
		break
	case 0xE000:
		x := (op & 0x0F00) >> 8
		switch op & 0x00FF {
		case 0x9E: // Ex9E - SKP Vx
			// Skip next instruction if key with the value of Vx is pressed.
			// Checks the keyboard, and if the key corresponding to the value of Vx
			// is currently in the down position, PC is increased by 2.
			if c.keypad[c.V[x]] == 1 {
				c.PC += 2
			}
			c.PC += 2
			break
		case 0xA1: // ExA1 - SKNP Vx
			// Skip next instruction if key with the value of Vx is not pressed.
			// Checks the keyboard, and if the key corresponding to the value of Vx
			// is currently in the up position, PC is increased by 2.
			if c.keypad[c.V[x]] == 0 {
				c.PC += 2
			}
			c.PC += 2
			break
		default:
			return op, fmt.Errorf("Unknown opcode: 0x%04X", op)
		}
	case 0xF000:
		x := (op & 0x0F00) >> 8
		switch op & 0x00FF {
		case 0x07: // Fx07 - LD Vx, DT
			// Set Vx = delay timer value.
			// The value of DT is placed into Vx.
			c.V[x] = c.delayTimer

			c.PC += 2
			break
		case 0x0A: // Fx0A - LD Vx, K
			// Wait for a key press, store the value of the key in Vx.
			// All execution stops until a key is pressed, then the value
			// of that key is stored in Vx.
			pressed := false
			for !pressed {
				for i := 0; i < 16; i++ {
					if c.keypad[i] == 1 {
						c.V[x] = byte(i)
						pressed = true
					}
				}
			}
			c.PC += 2
			break
		case 0x15: // Fx15 - LD DT, Vx
			// Set delay timer = Vx.
			// DT is set equal to the value of Vx.
			c.delayTimer = c.V[x]
			c.PC += 2
			break
		case 0x18: // Fx18 - LD ST, Vx
			// Set sound timer = Vx.
			// ST is set equal to the value of Vx.
			c.soundTimer = c.V[x]
			c.PC += 2
			break
		case 0x1E: // Fx1E - ADD I, Vx
			// Set I = I + Vx.
			// The values of I and Vx are added, and the results are stored in I.
			c.I += uint16(c.V[x])
			c.PC += 2
			break
		case 0x29: // Fx29 - LD F, Vx
			// Set I = location of sprite for digit Vx.
			// The value of I is set to the location for the hexadecimal sprite
			// corresponding to the value of Vx.
			c.I += uint16(c.V[x]) * uint16(0x05)
			c.PC += 2
			break
		case 0x33: // Fx33 - LD B, Vx
			// Store BCD representation of Vx in memory locations I, I+1, and I+2.
			// The interpreter takes the decimal value of Vx, and places the
			// hundreds digit in memory at location in I, the tens digit at location
			// I+1, and the ones digit at location I+2.
			c.memory[c.I] = c.V[x] / 100
			c.memory[c.I+1] = (c.V[x] / 10) % 10
			c.memory[c.I+2] = (c.V[x] % 100) % 10

			c.PC += 2
			break
		case 0x55: // Fx55 - LD [I], Vx
			// Store registers V0 through Vx in memory starting at location I.
			// The interpreter copies the values of registers V0 through Vx into
			// memory, starting at the address in I.
			var i uint16
			for i = 0; i <= x; i++ {
				c.memory[c.I+i] = c.V[i]
			}
			c.PC += 2
			break
		case 0x65: // Fx65 - LD Vx, [I]
			// Read registers V0 through Vx from memory starting at location I.
			// The interpreter reads values from memory starting at location I into
			// registers V0 through Vx.
			var i uint16
			for i = 0; i <= x; i++ {
				c.V[i] = c.memory[c.I+i]
			}
			c.PC += 2
			break
		default:
			return op, fmt.Errorf("Unknown opcode: 0x%04X", op)
		}
	default:
		return op, fmt.Errorf("Unknown opcode: 0x%04X", op)
	}

	return op, nil
}
