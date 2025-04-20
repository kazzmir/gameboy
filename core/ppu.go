package core

import (
    "log"
)

type PPU struct {
    ViewPortX uint8
    ViewPortY uint8
    WindowX uint8
    WindowY uint8
    Palette uint8
    ObjPalette0 uint8
    ObjPalette1 uint8
    LCDControl uint8
    LCDY uint8

    OAM []uint8
    Sprites []Sprite
    LineSprites []int

    Dot uint16
    Screen [][]uint8

    Debug bool
}

func MakePPU() *PPU {
    screen := make([][]uint8, 144)
    for i := range screen {
        screen[i] = make([]uint8, 160)
    }

    return &PPU{
        Screen: screen,
        OAM: make([]uint8, 160),
        Sprites: make([]Sprite, 40),
        LineSprites: make([]int, 10),
    }
}

type Sprite struct {
    X uint8
    Y uint8
    TileIndex uint8
    Attributes uint8
}

func (ppu *PPU) ReadSprites() []Sprite {
    for index := range len(ppu.Sprites) {
        ppu.Sprites[index].X = ppu.OAM[index*4]
        ppu.Sprites[index].Y = ppu.OAM[index*4+1]
        ppu.Sprites[index].TileIndex = ppu.OAM[index*4+2]
        ppu.Sprites[index].Attributes = ppu.OAM[index*4+3]
    }

    return ppu.Sprites
}

// address is assumed to be in the range 0-160, not 0xfe00-0xfea0
func (ppu *PPU) WriteOAM(address uint16, value uint8) {
    if address < uint16(len(ppu.OAM)) {
        ppu.OAM[address] = value
    } else {
        log.Printf("PPU: OAM write out of bounds: %x", address)
    }
}

func (ppu *PPU) LargeSpriteMode() bool {
    // bit 2 of LCDControl
    return (ppu.LCDControl & 0x4) != 0
}

func (ppu *PPU) Run(ppuCycles uint64) {
    for range ppuCycles {
        ppu.Dot += 1
        if ppu.Dot < 80 {
            // find sprites that hit this scanline
            // scan on the last possible dot
            if ppu.Dot == 79 {
                var size uint8 = 8
                if ppu.LargeSpriteMode() {
                    size = 16
                }

                ppu.LineSprites = ppu.LineSprites[:0]
                for index := range ppu.Sprites {
                    if ppu.LCDY >= ppu.Sprites[index].Y && ppu.LCDY < ppu.Sprites[index].Y+size {
                        ppu.LineSprites = append(ppu.LineSprites, index)
                    }
                }

                if ppu.Debug {
                    log.Printf("PPU: Found %d sprites on line %d", len(ppu.LineSprites), ppu.LCDY)
                }
            }

            // OAM search, mode 2
        } else if ppu.Dot >= 80 {
            // draw pixels, mode 3. usually 172 dots, but could be less
            // after 172 dots, enter mode 0 horizontal blank

            if ppu.Dot < 252 {
                x := ppu.Dot - 80

                var size uint8 = 8
                if ppu.LargeSpriteMode() {
                    size = 16
                }

                for _, spriteIndex := range ppu.LineSprites {
                    if x >= uint16(ppu.Sprites[spriteIndex].X) && x < uint16(ppu.Sprites[spriteIndex].X+size) {
                        // find pixel and write it into the screen
                    }
                }
            }

        }

        if ppu.Dot >= 456 {
            ppu.Dot = 0
            ppu.LCDY += 1

            if ppu.LCDY >= 154 {
                ppu.LCDY = 0
            }
        }
    }
}
