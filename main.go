package main

import (
	"fmt"
	"os"
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

var fontset = [80]byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

func (c *Chip8) Init() {
	c.PC = 0x200
	copy(c.memory[0:80], fontset[:])
}

func (c *Chip8) LoadROM(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	if len(data) > 4096-0x200 {
		return fmt.Errorf("ROM too large: %d bytes", len(data))
	}

	copy(c.memory[0x200:0x200+len(data)], data)
	return nil
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
	emulator.StartTimers()

	emulator.DT = 60
	emulator.ST = 30

	if err := emulator.LoadROM("assets/roms/ibm.ch8"); err != nil {
		fmt.Println("Error loading ROM: ", err)
		return
	}

	for i := 0; i < 5; i++ {
		time.Sleep(500 * time.Millisecond)
		fmt.Printf("DT: %d, ST: %d\n", emulator.DT, emulator.ST)
	}

}
