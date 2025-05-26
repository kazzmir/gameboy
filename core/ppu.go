package core

import (
    "log"
    "image/color"
)

type PPU struct {
    ViewPortX uint8
    ViewPortY uint8
    WindowX uint8
    WindowY uint8
    Palette uint8
    ObjPalette0 uint8
    ObjPalette1 uint8
    LCDStatus uint8
    LCDControl uint8
    LCDY uint8
    LCDYCompare uint8

    VideoRam []uint8

    OAM []uint8
    Sprites []Sprite
    LineSprites []int

    Dot uint16
    Screen [][]color.RGBA
    // if the cpu should draw then this channel will have something in it
    Draw chan bool

    Debug bool
}

func MakePPU() *PPU {
    screen := make([][]color.RGBA, 144)
    for i := range screen {
        screen[i] = make([]color.RGBA, 160)
    }

    return &PPU{
        Screen: screen,
        VideoRam: make([]uint8, 8192),
        OAM: make([]uint8, 160),
        Sprites: make([]Sprite, 40),
        LineSprites: make([]int, 10),
        Draw: make(chan bool, 1),
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
        ppu.Sprites[index].Y = ppu.OAM[index*4]
        ppu.Sprites[index].X = ppu.OAM[index*4+1]
        ppu.Sprites[index].TileIndex = ppu.OAM[index*4+2]
        ppu.Sprites[index].Attributes = ppu.OAM[index*4+3]
    }
    // log.Printf("Sprites: %v", ppu.Sprites)

    return ppu.Sprites
}

func (ppu *PPU) WriteVRam(address uint16, value uint8) {
    if address < uint16(len(ppu.VideoRam)) {
        // if address >= 0x1800 && address <= 0x1fff {
        /*
        if address == 0x1800 {
            log.Printf("vram write 0x%x = 0x%x", address, value)
        }
        */
        ppu.VideoRam[address] = value
    } else {
        log.Printf("PPU: VRAM write out of bounds: %x", address)
    }
}

func (ppu *PPU) LoadVRam(address uint16) uint8 {
    if address < uint16(len(ppu.VideoRam)) {
        return ppu.VideoRam[address]
    }
    log.Printf("PPU: VRAM read out of bounds: %x", address)
    return 0
}

func (ppu *PPU) CopyOAM(data []uint8) {
    if len(data) > len(ppu.OAM) {
        log.Printf("PPU: OAM copy out of bounds: %x", len(data))
        return
    }

    copy(ppu.OAM, data)
}

// address is assumed to be in the range 0-160, not 0xfe00-0xfea0
func (ppu *PPU) WriteOAM(address uint16, value uint8) {
    if address < uint16(len(ppu.OAM)) {
        // log.Printf("Write oam 0x%x = 0x%x", address, value)
        ppu.OAM[address] = value
    } else {
        log.Printf("PPU: OAM write out of bounds: %x", address)
    }
}

func (ppu *PPU) LargeSpriteMode() bool {
    // bit 2 of LCDControl
    return (ppu.LCDControl & 0x4) != 0
}

func (ppu *PPU) BackgroundTileMapAddress() uint16 {
    // bit 3 of LCDControl
    if (ppu.LCDControl & 0x8) != 0 {
        return 0x9c00 - 0x8000
    }
    return 0x9800 - 0x8000
}

// returns the value of the bit at position bit
func bitN(value uint8, bit uint8) uint8 {
    return (value & (1<<bit)) >> bit

    /*
    if value&(1<<bit) != 0 {
        return 1
    }
    return 0
    */
}

type System interface {
    EnableStatInterrupt()
}

// set lower 2 bits of LCDStatus
func (ppu *PPU) SetLCDStatus(value uint8) {
    ppu.LCDStatus = (ppu.LCDStatus & 0b11111100) | (value & 0b11)
}

// lower 2 bits of LCDStatus
func (ppu *PPU) GetPPUMode() uint8 {
    return ppu.LCDStatus & 0b11
}

func (ppu *PPU) Run(ppuCycles uint64, system System) {
    for range ppuCycles {
        if ppu.LCDStatus & 0b100000 != 0 {
            if ppu.LCDYCompare == ppu.LCDY {
                system.EnableStatInterrupt()
            }
        }

        if ppu.LCDYCompare == ppu.LCDY {
            ppu.LCDStatus |= 0b100
        } else {
            ppu.LCDStatus &= ^uint8(0b100)
        }

        ppu.Dot += 1
        if ppu.Dot < 80 {
            ppu.SetLCDStatus(2)

            // find sprites that hit this scanline
            // scan on the last possible dot
            if ppu.Dot == 79 {
                var size uint8 = 8
                if ppu.LargeSpriteMode() {
                    size = 16
                }

                ppu.LineSprites = ppu.LineSprites[:0]

                sprites := ppu.ReadSprites()

                for index := range len(sprites) {
                    spriteY := int(sprites[index].Y) - 16

                    if int(ppu.LCDY) >= spriteY && int(ppu.LCDY) < spriteY+int(size) {
                        ppu.LineSprites = append(ppu.LineSprites, index)
                    }
                }

                if ppu.Debug && len(ppu.LineSprites) > 0 {
                    log.Printf("PPU: Found %d sprites on line %d", len(ppu.LineSprites), ppu.LCDY)
                }
            }

            // OAM search, mode 2
        } else if ppu.Dot >= 80 {
            ppu.SetLCDStatus(2)
            // draw pixels, mode 3. usually 172 dots, but could be less
            // after 172 dots, enter mode 0 horizontal blank

            if ppu.Dot < 252 {
                ppu.SetLCDStatus(3)
                x := ppu.Dot - 80

                if x < 160 && ppu.LCDY < 144 {

                    var size uint8 = 8
                    if ppu.LargeSpriteMode() {
                        size = 16
                    }

                    {
                        // get background tile index
                        tileMap1Address := ppu.BackgroundTileMapAddress()

                        // each tile map is 32x32, where each tile is 8x8 so a total of 256x256 pixels
                        // to find the pixel value at position x,y we compute the tile index as y/8*32+x/8

                        var backgroundX uint16 = (uint16(ppu.ViewPortX) + x/8) % 256
                        var backgroundY uint16 = (uint16(ppu.ViewPortY) + uint16(ppu.LCDY)/8) % 256

                        // tileIndex := ppu.VideoRam[tileMap1Address + uint16(ppu.LCDY/8) * 32 + uint16(x/8)]
                        tileIndex := ppu.VideoRam[tileMap1Address + backgroundY * 32 + backgroundX]

                        vramIndex := uint16(tileIndex)*16

                        /*
                        if ppu.LCDY == 8 || ppu.LCDY == 9 {
                            log.Printf("dot x=%v y=%v, bg x=%v y=%v tile=0x%x vram=0x%x", x, ppu.LCDY, backgroundX, backgroundY, tileIndex, vramIndex)
                        }
                        */

                        yValue := uint16(ppu.LCDY) % 8

                        lowByte := ppu.VideoRam[vramIndex + yValue * 2]
                        highByte := ppu.VideoRam[vramIndex + yValue * 2 + 1]
                        bit := uint8(7 - (x & 7))
                        paletteColor := bitN(lowByte, bit) | (bitN(highByte, bit) << 1)

                        /*
                        if tileIndex == 0x1d {
                            log.Printf("x=%v y=%v low=%v high=%v bit=%v palette=%v", x, ppu.LCDY, lowByte, highByte, bit, paletteColor)
                        }
                        */

                        var pixelColor color.RGBA

                        // FIXME: use real palette
                        switch paletteColor {
                            case 0: pixelColor = color.RGBA{255, 255, 255, 255} // white
                            case 1: pixelColor = color.RGBA{192, 192, 192, 255} // light gray
                            case 2: pixelColor = color.RGBA{96, 96, 96, 255} // dark gray
                            case 3: pixelColor = color.RGBA{0, 0, 0, 255} // black
                        }

                        // r, g, b, a := pixelColor.RGBA()
                        // convert to RGBA8888
                        // ppu.Screen[ppu.LCDY][x] = (r << 24) | (g << 16) | (b << 8) | (a << 0)
                        ppu.Screen[ppu.LCDY][x] = pixelColor
                    }

                    for _, spriteIndex := range ppu.LineSprites {
                        sprite := &ppu.Sprites[spriteIndex]
                        spriteX := int(sprite.X) - 8
                        // spriteY := int(sprite.Y) - 16

                        if int(x) >= spriteX && int(x) < spriteX+int(size) {
                            // FIXME: handle size
                            vramIndex := uint16(sprite.TileIndex)*16

                            yValue := uint16(ppu.LCDY) % 8

                            lowByte := ppu.VideoRam[vramIndex + yValue * 2]
                            highByte := ppu.VideoRam[vramIndex + yValue * 2 + 1]
                            // bit := uint8(7 - (x & 7))

                            // FIXME: what about size 16?
                            bit := uint8(7 - (int(x) - spriteX))
                            paletteColor := bitN(lowByte, bit) | (bitN(highByte, bit) << 1)

                            var pixelColor color.RGBA

                            // FIXME: use real palette
                            switch paletteColor {
                                case 0: pixelColor = color.RGBA{255, 255, 255, 255} // white
                                case 1: pixelColor = color.RGBA{192, 192, 192, 255} // light gray
                                case 2: pixelColor = color.RGBA{96, 96, 96, 255} // dark gray
                                case 3: pixelColor = color.RGBA{0, 0, 0, 255} // black
                            }

                            // r, g, b, a := pixelColor.RGBA()
                            // convert to RGBA8888
                            // ppu.Screen[ppu.LCDY][x] = (r << 24) | (g << 16) | (b << 8) | (a << 0)
                            ppu.Screen[ppu.LCDY][x] = pixelColor

                            // find pixel and write it into the screen
                        }
                    }
                }
            } else {
                ppu.SetLCDStatus(0)
            }

        }

        switch ppu.GetPPUMode() {
            case 0:
                if ppu.LCDStatus & 0b100000 != 0 {
                    system.EnableStatInterrupt()
                }
            case 1:
                if ppu.LCDStatus & 0b10000 != 0 {
                    system.EnableStatInterrupt()
                }
            case 2:
                if ppu.LCDStatus & 0b1000 != 0 {
                    system.EnableStatInterrupt()
                }
        }

        if ppu.Dot >= 456 {
            ppu.Dot = 0
            ppu.LCDY += 1

            if ppu.LCDY == 144 {
                select {
                    case ppu.Draw <- true:
                    default:
                }
            }

            if ppu.LCDY >= 154 {
                ppu.LCDY = 0
            }
        }
    }
}
