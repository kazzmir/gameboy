package core

type PPU struct {
    ViewPortX uint8
    ViewPortY uint8
    WindowX uint8
    WindowY uint8
}

func MakePPU() *PPU {
    return &PPU{}
}
