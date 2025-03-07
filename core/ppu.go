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
}

func MakePPU() *PPU {
    return &PPU{}
}
