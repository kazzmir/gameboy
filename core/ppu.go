package core

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

    Dot uint16
    Screen [][]uint8
}

func MakePPU() *PPU {
    screen := make([][]uint8, 144)
    for i := range screen {
        screen[i] = make([]uint8, 160)
    }

    return &PPU{
        Screen: screen,
    }
}

func (ppu *PPU) Run(cpuCycles uint64) {
    ppu.Dot += uint16(cpuCycles * 4)
    if ppu.Dot >= 456 {
        ppu.Dot = 0
        ppu.LCDY += 1

        if ppu.LCDY >= 154 {
            ppu.LCDY = 0
        }
    }
}
