package main

import (
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
	chip8.LoadBytes(0x50, interpreter.FontSet)
	chip8.LoadRom(game)
	err = chip8.Run()
	if err != nil {
		log.Fatal("|| Runtime error: %s", err)
	}
}
