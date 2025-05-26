package core

import (
    "io"
    "os"
)

type GameboyFile struct {
    Data []byte
}

func (gameboy *GameboyFile) GetRom() []uint8 {
    return gameboy.Data
}

func (gameboy *GameboyFile) GetTitle() string {
    start := 0x134
    end := 0x143 + 1

    if end > len(gameboy.Data) {
        return ""
    }

    last := start
    for last < end {
        if gameboy.Data[last] == 0 {
            break
        }
        last += 1
    }

    return string(gameboy.Data[start:last])
}

func (gameboy *GameboyFile) GetManufacturerCode() []byte {
    start := 0x13F
    end := 0x142 + 1

    if end > len(gameboy.Data) {
        return nil
    }

    return gameboy.Data[start:end]
}

// 0x80: Game supports CGB functions, but works on old gameboys also.
// 0xc0: Game works on CGB only (physically the same as 0x80).
func (gameboy *GameboyFile) GetCGBFlag() byte {
    offset := 0x143
    if offset >= len(gameboy.Data) {
        return 0
    }

    return gameboy.Data[offset]
}

func (gameboy *GameboyFile) GetNewLicenseeCode() []byte {
    start := 0x144
    end := 0x145 + 1

    if end > len(gameboy.Data) {
        return nil
    }

    return gameboy.Data[start:end]
}

// set to 0x3 to use SGB
func (gameboy *GameboyFile) GetSGBFlag() byte {
    offset := 0x146
    if offset >= len(gameboy.Data) {
        return 0
    }

    return gameboy.Data[offset]
}

// specifies the mapper
func (gameboy *GameboyFile) GetCartridgeType() byte {
    offset := 0x147
    if offset >= len(gameboy.Data) {
        return 0
    }

    return gameboy.Data[offset]
}

// returns rom size in bytes
func (gameboy *GameboyFile) GetRomSize() uint64 {
    offset := 0x148
    if offset >= len(gameboy.Data) {
        return 0
    }

    value := gameboy.Data[offset]

    return (32 * 1024) << value
}

// returns ram size in bytes
func (gameboy *GameboyFile) GetRAMSize() uint64 {
    offset := 0x149
    if offset >= len(gameboy.Data) {
        return 0
    }

    value := gameboy.Data[offset]
    switch value {
        case 0: return 0
        case 1: return 0
        case 2: return 8 * 1024
        case 3: return 32 * 1024
        case 4: return 128 * 1024
        case 5: return 64 * 1024
    }

    return 0
}

// 0x00: Japanese, 0x01: Non-Japanese
func (gameboy *GameboyFile) GetDestinationCode() byte {
    offset := 0x14a
    if offset >= len(gameboy.Data) {
        return 0
    }

    return gameboy.Data[offset]
}

func (gameboy *GameboyFile) GetOldLicenseeCode() byte {
    offset := 0x14b
    if offset >= len(gameboy.Data) {
        return 0
    }

    return gameboy.Data[offset]
}

func (gameboy *GameboyFile) GetMaskROMVersionNumber() byte {
    offset := 0x14c
    if offset >= len(gameboy.Data) {
        return 0
    }

    return gameboy.Data[offset]
}

func (gameboy *GameboyFile) GetHeaderChecksum() byte {
    offset := 0x14d
    if offset >= len(gameboy.Data) {
        return 0
    }

    return gameboy.Data[offset]
}

func (gameboy *GameboyFile) GetGlobalChecksum() uint16 {
    start := 0x14e
    end := 0x14f

    if end > len(gameboy.Data) {
        return 0
    }

    high := uint16(gameboy.Data[start])
    low := uint16(gameboy.Data[end])

    return (high << 8) | low
}

func LoadGameboy(reader io.Reader) (*GameboyFile, error) {
    data, err := io.ReadAll(reader)

    return &GameboyFile{Data: data}, err
}

func LoadGameboyFromFile(filename string) (*GameboyFile, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    return LoadGameboy(file)
}
