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
        ppu.Sprites[index].X = ppu.OAM[index*4]
        ppu.Sprites[index].Y = ppu.OAM[index*4+1]
        ppu.Sprites[index].TileIndex = ppu.OAM[index*4+2]
        ppu.Sprites[index].Attributes = ppu.OAM[index*4+3]
    }

    return ppu.Sprites
}

func (ppu *PPU) WriteVRam(address uint16, value uint8) {
    if address < uint16(len(ppu.VideoRam)) {
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
                        tileIndex := ppu.VideoRam[tileMap1Address + uint16(ppu.LCDY/8) * 32 + uint16(x/8)]

                        vramIndex := uint16(tileIndex)*16
                        lowByte := ppu.VideoRam[vramIndex]
                        highByte := ppu.VideoRam[vramIndex+1]
                        bit := uint8(x - (x/8)*8)
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
                    }

                    for _, spriteIndex := range ppu.LineSprites {
                        if x >= uint16(ppu.Sprites[spriteIndex].X) && x < uint16(ppu.Sprites[spriteIndex].X+size) {
                            vramIndex := uint16(ppu.Sprites[spriteIndex].TileIndex)*16+uint16(ppu.LCDY-ppu.Sprites[spriteIndex].Y)
                            lowByte := ppu.VideoRam[vramIndex]
                            highByte := ppu.VideoRam[vramIndex+1]

                            // FIXME: what about size 16?
                            bit := uint8(x - uint16(ppu.Sprites[spriteIndex].X))
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

                // clear screen, not needed later once every pixel is drawn
                /*
                for y := range len(ppu.Screen) {
                    for x := range len(ppu.Screen[y]) {
                        ppu.Screen[y][x] = color.RGBA{A: 0xff}
                    }
                }
                */
            }
        }
    }
}
