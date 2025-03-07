package core

type PPU struct {
    ViewPortY uint8
    ViewPortX uint8
}

func MakePPU() *PPU {
    return &PPU{}
}
