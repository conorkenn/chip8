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
	keys    [16]bool
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
	case 0x1000: // 1NNN jump
		c.PC = opcode & 0x0FFF
	case 0x2000: // 2NNN call subroutine
		if int(c.SP)+1 >= len(c.stack) {
			panic("stack overflow")
		}
		c.stack[c.SP] = c.PC
		c.SP++
		c.PC = opcode & 0x0FFF
	case 0x3000: // 3XNN skip if vx == nn
		x := (opcode & 0x0F00) >> 8
		nn := byte(opcode & 0x00FF)
		if c.V[x] == nn {
			c.PC += 2
		}
	case 0x4000: // 4XNN skip if vx != nn
		x := (opcode & 0x0F00) >> 8
		nn := byte(opcode & 0x00FF)
		if c.V[x] != nn {
			c.PC += 2
		}
	case 0x5000: // 5XY0 skip if VX != VY
		x := (opcode & 0x0F00) >> 8
		y := (opcode & 0x00FF) >> 4
		if c.V[x] == c.V[y] {
			c.PC += 2
		}
	case 0x6000: // 6XNN set vx
		x := (opcode & 0x0F00) >> 8
		c.V[x] = byte(opcode & 0x00FF)
	case 0x7000: // 7XNN add to vx
		x := (opcode & 0x0F00) >> 8
		c.V[x] += byte(opcode & 0x00FF)
	case 0x8000:
		x := (opcode & 0x0F00) >> 8
		y := (opcode & 0x00F0) >> 4
		switch opcode & 0x000F {
		case 0x0: // 8XY0 vx = vy
			c.V[x] = c.V[y]
		case 0x1: // 8XY1 vx or vy
			c.V[x] |= c.V[y]
		case 0x2: // 8XY2 vx and vy
			c.V[x] &= c.V[y]
		case 0x3: // 8XY3 vx xor vy
			c.V[x] ^= c.V[y]
		case 0x4: // 8XY4 vx += vy
			sum := uint16(c.V[x]) + uint16(c.V[y])
			c.V[x] = byte(sum & 0xFF)
			c.V[0xF] = byte((sum >> 8) & 0x01)
		case 0x5: // 8XY5 vx -= vy
			diff := uint16(c.V[x]) - uint16(c.V[y])
			c.V[x] = byte(diff & 0xFF)
			c.V[0xF] = byte(0)
			if c.V[x] > c.V[y] {
				c.V[0xF] = 1
			}
		case 0x6: //8XY6 vx >>-1 vf lsb
			c.V[0xF] = c.V[x] & 0x01
			c.V[x] >>= 1
		case 0x7: // 8XY7 vx = vy - vx
			diff := uint16(c.V[y]) - uint16(c.V[x])
			c.V[x] = byte(diff & 0xFF)
			c.V[0xF] = byte(0) // No borrow if VY >= VX
			if c.V[y] > c.V[x] {
				c.V[0xF] = 1 // No borrow
			}
		case 0xE: // 8XYE vx <<=1 vf msb
			c.V[0xF] = (c.V[x] & 0x80) >> 7
			c.V[x] <<= 1
		}
	case 0x9000: // 9XY0 skip if VX != VY
		x := (opcode & 0x0F00) >> 8
		y := (opcode & 0x00F0) >> 4
		if c.V[x] != c.V[y] {
			c.PC += 2
		}
	case 0xA000: // ANNN: Set I = NNN
		c.I = opcode & 0x0FFF
	case 0xB000: // BNNN: Jump to NNN + V0
		c.PC = (opcode & 0x0FFF) + uint16(c.V[0])
	case 0xC000: // CXNN: VX = random & NN
		x := (opcode & 0x0F00) >> 8
		nn := byte(opcode & 0x00FF)
		randByte := byte(time.Now().Nanosecond() % 256)
		c.V[x] = randByte & nn
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
	case 0xF000:
		x := (opcode & 0x0F00) >> 8
		switch opcode & 0x00FF {
		case 0x07: // FX07: VX = DT
			c.V[x] = c.DT
		case 0x15: // FX15: DT = VX
			c.DT = c.V[x]
		case 0x18: // FX18: ST = VX
			c.ST = c.V[x]
		case 0x1E: // FX1E: I += VX
			i := uint16(c.V[x])
			if c.I+uint16(c.V[x]) > 0xFFF { // Optional: Set VF for overflow
				c.V[0xF] = 1
			} else {
				c.V[0xF] = 0
			}
			c.I += i
		case 0x0A: // FX0A wait for key press
			for i := 0; i < 16; i++ {
				if c.keys[i] {
					c.V[x] = byte(i)
					return
				}
			}
			c.PC -= 2
			return

		case 0x29: // FX29 set sprite address for digit
			c.I = uint16(c.V[x]&0x0F) * 5
		case 0x33: // FX33 store bcd of vx
			value := c.V[x]
			c.memory[c.I] = value / 100
			c.memory[c.I+1] = (value / 10) % 10
			c.memory[c.I+2] = value % 10
		case 0x55:
			for i := uint16(0); i <= x; i++ {
				c.memory[c.I+i] = c.V[i]
			}
		case 0x65:
			for i := uint16(0); i <= x; i++ {
				c.V[i] = c.memory[c.I+i]
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

func (c *Chip8) updateKeys() {
	var input byte
	fmt.Scanf("%c", &input)
	if input >= '0' && input <= '9' {
		c.keys[input-'0'] = true
	} else if input >= 'A' && input <= 'F' {
		c.keys[input-'A'+10] = true
	}
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

	for i := range 1000 {
		//emulator.updateKeys()
		emulator.Cycle()
		time.Sleep(2 * time.Millisecond) // ~500 Hz
		if i%100 == 0 {
			fmt.Printf("Cycle %d: PC=%04X, V0=%02X\n", i, emulator.PC, emulator.V[0])
			emulator.PrintDisplay()
		}
	}

}
