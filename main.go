package main

import (
	"time"
)

type Chip8 struct {
	memory  [4096]byte
	display [64][32]bool // 64 x 32 display
	VF      [16]byte     // 16 8-bit registers
	PC      uint16       // program counter
	I       uint16       // index register
	stack   [16]uint16   // stack for subroutines
	SP      byte         // stack pointer
	V       [16]byte     // 8 bit general registers
	DT      byte         // delay timer
	ST      byte         // sound timer
}

func (c *Chip8) Init() {
	c.PC = 0x200
}

func (c *Chip8) StartTimers() {
	go func() {
		ticker := time.NewTicker(time.Second / 60)
		defer ticker.Stop()
		for range ticker.C {
			if c.DT > 0 {
				c.DT--
			}
			if c.ST > 0 {
				c.ST--
				// beep
			}
		}
	}()
}

func main() {
	emulator := Chip8{}
	emulator.Init()

	emulator.DT = 60
	emulator.ST = 30

}
