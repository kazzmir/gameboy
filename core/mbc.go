package core

import (
    "fmt"
    "log"
)

type MBC interface {
    Read(address uint16) uint8
    Write(address uint16, value uint8)
}

func (mbc0 *MBC0) Read(address uint16) uint8 {
    if address < 0x4000 {
        return mbc0.rom[address]
    } else if address < 0x8000 {
        return mbc0.rom[address]
    }
    return 0
}

func (mbc0 *MBC0) Write(address uint16, value uint8) {
    if address == 0x2000 {
        // ignore
    } else if mbc0.showError {
        log.Printf("Attempted to write to ROM at address 0x%x", address)
    }
}

type MBC0 struct {
    rom []uint8
    showError bool
}

type MBC1 struct {
    rom []uint8
}

func (mbc1 *MBC1) Read(address uint16) uint8 {
    if address < 0x4000 {
        return mbc1.rom[address]
    } else if address < 0x8000 {
        return mbc1.rom[address]
    }
    return 0
}

func (mbc1 *MBC1) Write(address uint16, value uint8) {
}

var _ MBC = &MBC0{}
var _ MBC = &MBC1{}

func MakeMBC(mbcType uint8, rom []uint8) (MBC, error) {
    switch mbcType {
        case 0:
            return &MBC0{rom: rom}, nil
        case 1:
            return &MBC1{}, nil
        default:
            return nil, fmt.Errorf("Unknown MBC type")
    }
}
