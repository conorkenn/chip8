package main

import (
	"fmt"
	"os"
	"time"
)

type Chip8 struct {
	memory  [4096]byte
	display [64][32]bool // 64 x 32 display
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

func (c *Chip8) Fetch() uint16 {
	if int(c.PC)+1 >= len(c.memory) {
		panic(fmt.Sprintf("PC out of bounds: %04X", c.PC))
	}
	opcode := uint16(c.memory[c.PC])<<8 | uint16(c.memory[c.PC+1])
	c.PC += 2
	return opcode
}

func (c *Chip8) Execute(opcode uint16) {
	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode {
		case 0x00E0: // clear
			for x := range c.display {
				for y := range c.display[x] {
					c.display[x][y] = false
				}
			}
		case 0x00EE: // return from subroutine
			if c.SP == 0 {
				panic("stack underflow")
			}
			c.SP--
			c.PC = c.stack[c.SP]
		default:
			fmt.Printf("Unknown 0x0 opcode: %04X\n", opcode)

		}
	case 0x1000: // jump
		c.PC = opcode & 0x0FFF
	case 0x6000: // set vx
		x := (opcode & 0x0F00) >> 8
		c.V[x] = byte(opcode & 0x00FF)
	case 0x7000: // add to vx
		x := (opcode & 0x0F00) >> 8
		c.V[x] += byte(opcode & 0x00FF)
	case 0xD000: //dxyn
		x := c.V[(opcode&0x0F00)>>8]
		y := c.V[(opcode&0x00F0)>>4]
		height := opcode & 0x000F
		c.V[0xF] = 0

		for row := uint16(0); row < height; row++ {
			spriteByte := c.memory[c.I+row]
			for col := uint8(0); col < 8; col++ {
				if (spriteByte & (0x80 >> col)) != 0 {
					xPos := (x + col) % 64
					yPos := (y + uint8(row)) % 32
					current := c.display[xPos][yPos]
					c.display[xPos][yPos] = current != true
					if current && !c.display[xPos][yPos] {
						c.V[0xF] = 1
					}
				}
			}
		}
	default:
		fmt.Printf("Unkown opcode: %04X\n", opcode)
	}
}

func (c *Chip8) Cycle() {
	opcode := c.Fetch()
	c.Execute(opcode)
}

func (c *Chip8) PrintDisplay() {
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			if c.display[x][y] {
				fmt.Print("█")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
	fmt.Println("---")
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

	for i := 0; i < 1000; i++ {
		emulator.Cycle()
		time.Sleep(2 * time.Millisecond) // ~500 Hz
		if i%100 == 0 {
			fmt.Printf("Cycle %d: PC=%04X, V0=%02X\n", i, emulator.PC, emulator.V[0])
			emulator.PrintDisplay()
		}
	}

}
