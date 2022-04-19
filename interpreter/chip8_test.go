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

func TestFetchInstruction(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x42, 0x69, 0x68, 0x67}
	chip8.LoadBytes(0x200, testBytes)

	got := chip8.FetchInstruction()

	assert.Equal(t, uint16(0x4269), got)
}

func TestExecuteInstructionCLS(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x00, 0xE0, 0x42, 0x69}
	chip8.LoadBytes(0x200, testBytes)

	opcode := chip8.FetchInstruction()
	_, err := chip8.ExecuteOpcode(opcode)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	assert.Equal(t, uint16(0x202), chip8.PC)
}

func TestExecuteInstructionRET(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x00, 0xE0, 0x00, 0xEE, 0x42, 0x69}
	chip8.LoadBytes(0x200, testBytes)

	// Manually set stack
	chip8.stack[0] = uint16(0x666)
	chip8.stack[1] = uint16(0x222)
	chip8.SP = 1

	chip8.Run()

	assert.Equal(t, uint16(0x222), chip8.PC)
}
