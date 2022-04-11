package interpreter

import (
	"log"
	"os"
	"testing"
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
		t.Errorf("Expected: 0x%x, Got: 0x%x", expected, got)
	}
	chip8.LoadRom(game)
	got = chip8.memory[0x200]
	notExpected := 0
	if got == byte(notExpected) {
		t.Errorf("Expected NOT to be: 0x%x, Got: 0x%x", notExpected, got)
	}
}

func TestLoadBytes(t *testing.T) {
	chip8 := NewChip8()
	testBytes := []byte{0x42, 0x69}
	chip8.LoadBytes(0x3, testBytes)

	got := chip8.memory[0x3]
	expected := 0x42
	if got != byte(expected) {
		t.Errorf("Expected: 0x%x, Got: 0x%x", expected, got)
	}
	got = chip8.memory[0x4]
	expected = 0x69
	if got != byte(expected) {
		t.Errorf("Expected: 0x%x, Got: 0x%x", expected, got)
	}
}
