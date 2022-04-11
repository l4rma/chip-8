package main

import (
	"fmt"

	"github.com/l4rma/chip-8/interpreter"
)

func main() {
	chip8 := interpreter.NewChip8()
	fmt.Printf("Chip8: %v", chip8)
}
