package interpreter

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadRom(t *testing.T) {
	chip8 := NewChip8()
	game, err := os.Open("../roms/space_invaders.ch8")
	if err != nil {
		log.Panicf("Error opening file: %s", err)
		t.Errorf("Error: %v", err)
	}
	got := chip8.memory[0x200]
	expected := 0
	if got != byte(expected) {
		t.Errorf("Expected: 0x%02X, Got: 0x%02X", expected, got)
	}
	chip8.LoadRom(game)
	got = chip8.memory[0x200]
	notExpected := 0
	if got == byte(notExpected) {
		t.Errorf("Expected NOT to be: 0x%02X, Got: 0x%02X", notExpected, got)
	}
}

func TestLoadBytes(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x42, 0x69}
	chip8.LoadBytes(0x3, testBytes)

	got := chip8.memory[0x3]
	expected := 0x42
	if got != byte(expected) {
		t.Errorf("Expected: 0x%02X, Got: 0x%02X", expected, got)
	}
	got = chip8.memory[0x4]
	expected = 0x69
	if got != byte(expected) {
		t.Errorf("Expected: 0x%02X, Got: 0x%02X", expected, got)
	}
}

/*** INSTRUCTION TESTS ***/
func TestFetchInstruction(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x42, 0x69, 0x68, 0x67}
	chip8.LoadBytes(0x200, testBytes)

	got := chip8.FetchInstruction()

	assert.Equal(t, uint16(0x4269), got)
}

func TestCLS(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x00, 0xE0, 0x00, 0x69}
	chip8.LoadBytes(0x200, testBytes)

	opcode := chip8.FetchInstruction()
	_, err := chip8.ExecuteOpcode(opcode)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	assert.Equal(t, uint16(0x202), chip8.PC)
}

func TestRET(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x00, 0xE0, 0x00, 0xEE, 0x00, 0x69}
	chip8.LoadBytes(0x200, testBytes)

	// Manually set stack
	chip8.stack[0] = uint16(0x666)
	chip8.stack[1] = uint16(0x222)
	chip8.SP = 1

	chip8.Run()

	assert.Equal(t, uint16(0x224), chip8.PC)
}

func TestJMP(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x12, 0x04, 0x00, 0x00, 0x00, 0x69}
	chip8.LoadBytes(0x200, testBytes)

	chip8.Run()

	currentOp := uint16(chip8.memory[chip8.PC])<<8 | uint16(chip8.memory[chip8.PC+1])

	assert.Equal(t, uint16(0x204), chip8.PC)
	assert.Equal(t, uint16(0x0069), currentOp)
}

func TestCAL(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x22, 0x04, 0x00, 0x00, 0x00, 0x69}
	chip8.LoadBytes(0x200, testBytes)

	assert.Equal(t, uint8(0x00), chip8.SP)
	chip8.Run()

	currentOp := uint16(chip8.memory[chip8.PC])<<8 | uint16(chip8.memory[chip8.PC+1])

	assert.Equal(t, uint8(0x01), chip8.SP)
	assert.Equal(t, uint16(0x200), chip8.stack[chip8.SP])
	assert.Equal(t, uint16(0x204), chip8.PC)
	assert.Equal(t, uint16(0x0069), currentOp)
}

func TestSkipInstruction3xkk(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x32, 0x69, 0x00, 0x00, 0x00, 0x69}
	chip8.LoadBytes(0x200, testBytes)
	chip8.V[2] = 0x69
	chip8.Run()

	currentOp := uint16(chip8.memory[chip8.PC])<<8 | uint16(chip8.memory[chip8.PC+1])

	assert.Equal(t, uint16(0x204), chip8.PC)
	assert.Equal(t, uint16(0x0069), currentOp)
}

func TestSkipInstruction4xkk(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x42, 0x69, 0x00, 0x00, 0x00, 0x69}
	chip8.LoadBytes(0x200, testBytes)
	chip8.V[2] = 0x42
	chip8.Run()

	currentOp := uint16(chip8.memory[chip8.PC])<<8 | uint16(chip8.memory[chip8.PC+1])

	assert.Equal(t, uint16(0x204), chip8.PC)
	assert.Equal(t, uint16(0x0069), currentOp)
}

func TestSkipInstruction5xy0(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x52, 0x40, 0x00, 0x00, 0x00, 0x69}
	chip8.LoadBytes(0x200, testBytes)
	chip8.V[2] = 0x42
	chip8.V[4] = 0x42

	chip8.Run()

	currentOp := uint16(chip8.memory[chip8.PC])<<8 | uint16(chip8.memory[chip8.PC+1])

	assert.Equal(t, uint16(0x204), chip8.PC)
	assert.Equal(t, uint16(0x0069), currentOp)
}

func TestLoadVx6xkk(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x62, 0x69}
	chip8.LoadBytes(0x200, testBytes)
	chip8.Run()

	assert.Equal(t, uint8(0x69), chip8.V[2])
}

func TestADD7xkk(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x72, 0x39}
	chip8.LoadBytes(0x200, testBytes)

	chip8.V[2] = 0x30
	assert.NotEqual(t, uint8(0x69), chip8.V[2])

	chip8.Run()

	assert.Equal(t, uint8(0x69), chip8.V[2])
}

func TestLoadVxVy8xy0(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x82, 0x30}
	chip8.LoadBytes(0x200, testBytes)

	chip8.V[2] = 0x30
	chip8.V[3] = 0x69
	assert.NotEqual(t, uint8(0x69), chip8.V[2])

	chip8.Run()

	assert.Equal(t, uint8(0x69), chip8.V[2])
}

func TestOrVxVy8xy1(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x82, 0x31}
	chip8.LoadBytes(0x200, testBytes)

	chip8.V[2] = 0x20
	chip8.V[3] = 0x69
	assert.NotEqual(t, uint8(0x69), chip8.V[2])

	chip8.Run()

	assert.Equal(t, uint8(0x69), chip8.V[2])
}

func TestANDVxVy8xy2(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x82, 0x32}
	chip8.LoadBytes(0x200, testBytes)

	chip8.V[2] = 0xFF
	chip8.V[3] = 0x69
	assert.NotEqual(t, uint8(0x69), chip8.V[2])

	chip8.Run()

	assert.Equal(t, uint8(0x69), chip8.V[2])
}

func TestXORVxVy8xy3(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x82, 0x33}
	chip8.LoadBytes(0x200, testBytes)
	chip8.V[2] = 0x41
	chip8.V[3] = 0x28
	assert.NotEqual(t, uint8(0x69), chip8.V[2])

	chip8.Run()

	assert.Equal(t, uint8(0x69), chip8.V[2])
}

func TestADDVxVy8xy4(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x82, 0x34}
	chip8.LoadBytes(0x200, testBytes)
	chip8.V[2] = 0x88
	chip8.V[3] = 0x99
	assert.NotEqual(t, uint8(0x01), chip8.V[0xF])

	chip8.Run()

	assert.Equal(t, uint8(0x01), chip8.V[0xF])
}
