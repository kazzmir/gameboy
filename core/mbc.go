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
    romBank uint8 // 0x2000 register
    ramBank uint8 // 0x4000 register
    ramEnabled bool // 0x6000 register
    ram []uint8

    mode uint8
}

func (mbc1 *MBC1) Read(address uint16) uint8 {
    switch {
        case address < 0x4000:
            if mbc1.mode == 0 {
                return mbc1.rom[address]
            }

            address2 := (uint32(mbc1.ramBank) << 19) | uint32(address)
            if address2 >= uint32(len(mbc1.rom)) {
                log.Printf("Attempted to read from ROM at address 0x%x", address)
                return 0
            }
            return mbc1.rom[address2]
            
        case address >= 0x4000 && address < 0x8000:
            address2 := (uint32(mbc1.ramBank) << 19) | (uint32(mbc1.romBank) << 14) | uint32(address - 0x4000)
            if address2 >= uint32(len(mbc1.rom)) {
                log.Printf("Attempted to read from ROM at address 0x%x", address)
                return 0
            }
            // log.Printf("mbc1 read 0x4000 range address=0x%x, romBank=%d, ramBank=%d, address2=0x%x value=0x%x", address, mbc1.romBank, mbc1.ramBank, address2, mbc1.rom[address2])
            return mbc1.rom[address2]
        case address >= 0xA000 && address < 0xC000:
            if mbc1.ramEnabled {
                if mbc1.ramBank == 0 {
                    return mbc1.rom[address - 0xa000]
                }

                address2 := (uint32(mbc1.ramBank) << 13) | uint32(address - 0xa000)
                if address2 >= uint32(len(mbc1.ram)) {
                    log.Printf("Attempted to read from RAM at address 0x%x", address)
                    return 0
                }

                return mbc1.ram[address2]
            }
    }

    return 0
}

func (mbc1 *MBC1) Write(address uint16, value uint8) {
    // log.Printf("mbc1 write: 0x%x = 0x%x", address, value)
    switch {
        case address < 0x2000:
            if value & 0b1111 == 0xa {
                mbc1.ramEnabled = true
            } else {
                mbc1.ramEnabled = false
            }
        case address >= 0x2000 && address < 0x4000:
            mbc1.romBank = value & 0x1F
            if mbc1.romBank == 0 {
                mbc1.romBank = 1
            }
        case address >= 0x4000 && address < 0x6000:
            mbc1.ramBank = value & 0x03
        case address >= 0x6000 && address < 0x8000:
            // log.Printf("mbc1: set ram/rom mode to 0x%x", value)
            mbc1.mode = value & 0x01
        case address >= 0xA000 && address < 0xC000:
            if mbc1.ramEnabled {
                if mbc1.ramBank == 0 {
                    mbc1.rom[address - 0xa000] = value
                } else {
                    address2 := (uint32(mbc1.ramBank) << 13) | uint32(address - 0xa000)
                    if address2 >= uint32(len(mbc1.ram)) {
                        log.Printf("Attempted to write to RAM at address 0x%x", address)
                        return
                    }
                    mbc1.ram[address2] = value
                }
            } else {
                log.Printf("Warning: mbc1 write to RAM when disabled: 0x%x = 0x%x", address, value)
            }
        default:
            log.Printf("Warning: Attempted to write to ROM at address 0x%x: 0x%x", address, value)
    }
}

var _ MBC = &MBC0{}
var _ MBC = &MBC1{}

func MakeMBC(mbcType uint8, rom []uint8) (MBC, error) {
    switch mbcType {
        case 0:
            return &MBC0{rom: rom}, nil
        case 1:
            return &MBC1{
                rom: rom,
                romBank: 1,
                // FIXME: not all cartidges have all 32k
                ram: make([]uint8, 0x8000),
            }, nil
        default:
            return nil, fmt.Errorf("Unknown MBC type")
    }
}
