package main

import (
	"fmt"
	"log"
	"os"

	"github.com/l4rma/chip-8/interpreter"
)

func main() {
	chip8 := interpreter.NewChip8()
	game, err := os.Open("./roms/space_invaders.ch8")
	if err != nil {
		log.Panicf("Error opening file: %s", err)
	}
	chip8.LoadRom(game)
	fmt.Printf("Chip8: %v", chip8)
}
