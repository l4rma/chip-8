package interpreter

type chip8 struct {
	memory [4096]byte
}

func NewChip8() chip8 {
	return chip8{}
}
