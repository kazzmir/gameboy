package core

import (
    "log"
)

type CPU struct {
    // accumulator and flags
    A uint8
    F uint8
    BC uint16
    DE uint16
    HL uint16
    // stack pointer
    SP uint16
    // program counter
    PC uint16
    Cycles uint64

    Stopped bool
    Halted bool
}

type Opcode int
const (
    Nop Opcode = iota

    LoadBCImmediate
    LoadDEImmediate
    LoadHLImmediate
    LoadSPImmediate

    Load8Immediate
    StoreHLImmediate

    LoadR8R8

    StoreBCMemA
    StoreDEMemA
    StoreHLMemA
    StoreSPMemA

    LoadAMemBC
    LoadAMemDE
    LoadAMemHL
    LoadAMemSP

    StoreSPMem16

    IncBC
    IncDE
    IncHL
    IncSP

    Inc8B
    Inc8C
    Inc8D
    Inc8E
    Inc8H
    Inc8L
    Inc8HL
    Inc8A

    Dec8B
    Dec8C
    Dec8D
    Dec8E
    Dec8H
    Dec8L
    Dec8HL
    Dec8A

    DecBC
    DecDE
    DecHL
    DecSP

    AddHLBC
    AddHLDE
    AddHLHL
    AddHLSP

    AddAImmediate
    AdcAImmediate
    SubAImmediate

    AddAR8
    AddAHLMem

    SubAR8
    SubAHLMem

    AdcAR8
    AdcAHLMem

    AndAR8
    AndAHLMem

    XorAR8
    XorAHLMem

    OrAR8
    OrAHLMem

    SbcAR8
    SbcAHLMem

    CpAR8
    CpAHLMem

    RLCA
    RLA
    RRCA
    RRA

    DAA

    SCF

    CPL
    CCF

    JR

    JrNz
    JrZ
    JrNc
    JrC

    Stop
    Halt

    Unknown
)

type R16 uint8
const (
    R16BC R16 = 0
    R16DE R16 = 1
    R16HL R16 = 2
    R16SP R16 = 3
)

type R8 uint8
const (
    R8B R8 = 0
    R8C R8 = 1
    R8D R8 = 2
    R8E R8 = 3
    R8H R8 = 4
    R8L R8 = 5
    R8HL R8 = 6
    R8A R8 = 7
)

type Instruction struct {
    Opcode Opcode

    // if the instruction uses registers, these will be set to the register number
    R8_1 R8
    R8_2 R8
    R16_1 R16
    R16_2 R16
    Immediate8 uint8
    Immediate16 uint16
}

// pass in non-zero set value to set the bit to 1, 0 to set to 0
func setBit(value uint8, bit uint8, set uint8) uint8 {
    if set == 0 {
        return value & ^(1 << bit)
    }
    return value | (1 << bit)
}

func (cpu *CPU) SetFlagC(value uint8) {
    cpu.F = setBit(cpu.F, 4, value)
}

func (cpu *CPU) GetFlagC() uint8 {
    return (cpu.F >> 4) & 0b1
}

func (cpu *CPU) GetFlagZ() uint8 {
    return (cpu.F >> 7) & 0b1
}

func (cpu *CPU) SetFlagH(value uint8) {
    cpu.F = setBit(cpu.F, 5, value)
}

func (cpu *CPU) SetFlagN(value uint8) {
    cpu.F = setBit(cpu.F, 6, value)
}

func (cpu *CPU) SetFlagZ(value uint8) {
    cpu.F = setBit(cpu.F, 7, value)
}

func RotateRight(value uint8) (uint8, uint8) {
    carry := value & 0b1
    value = value >> 1
    value = value | (carry << 7)
    return value, carry
}

func RotateLeft(value uint8) (uint8, uint8) {
    carry := (value >> 7) & 0b1
    value = value << 1
    value = value | carry
    return value, carry
}

func (cpu *CPU) StoreMemory(address uint16, value uint8) {
    // TODO
}

func (cpu *CPU) LoadMemory8(address uint16) uint8 {
    // TODO
    return 0
}

func (cpu *CPU) AddHL(value uint16) {
    carry := uint32(cpu.HL) + uint32(value) > 0xffff

    halfCarry := ((cpu.HL & 0xfff) + (value & 0xfff)) & 0x1000 == 0x1000

    cpu.HL += value
    cpu.SetFlagN(0)

    if halfCarry {
        cpu.SetFlagH(1)
    } else {
        cpu.SetFlagH(0)
    }

    if carry {
        cpu.SetFlagC(1)
    } else {
        cpu.SetFlagC(0)
    }
}

func (cpu *CPU) GetRegister8(r8 R8) uint8 {
    switch r8 {
        case R8B: return uint8(cpu.BC >> 8)
        case R8C: return uint8(cpu.BC & 0xff)
        case R8D: return uint8(cpu.DE >> 8)
        case R8E: return uint8(cpu.DE & 0xff)
        case R8H: return uint8(cpu.HL >> 8)
        case R8L: return uint8(cpu.HL & 0xff)
        case R8A: return cpu.A
        // don't handle R8HL here
    }

    return 0
}

func (cpu *CPU) SetRegister8(r8 R8, value uint8) {
    switch r8 {
        case R8B:
            c := uint8(cpu.BC & 0xff)
            b := value
            cpu.BC = (uint16(b) << 8) | uint16(c)
        case R8C:
            b := uint8(cpu.BC >> 8)
            c := value
            cpu.BC = (uint16(b) << 8) | uint16(c)
        case R8D:
            e := uint8(cpu.DE & 0xff)
            d := value
            cpu.DE = (uint16(d) << 8) | uint16(e)
        case R8E:
            d := uint8(cpu.DE >> 8)
            e := value
            cpu.DE = (uint16(d) << 8) | uint16(e)
        case R8H:
            l := uint8(cpu.HL & 0xff)
            h := value
            cpu.HL = (uint16(h) << 8) | uint16(l)
        case R8L:
            h := uint8(cpu.HL >> 8)
            l := value
            cpu.HL = (uint16(h) << 8) | uint16(l)
        case R8A:
            cpu.A = value
    }
}

func (cpu *CPU) Execute(instruction Instruction) {
    switch instruction.Opcode {
        case Nop:
            cpu.Cycles += 1
        case LoadBCImmediate:
            cpu.Cycles += 3
            cpu.BC = instruction.Immediate16
        case LoadDEImmediate:
            cpu.Cycles += 3
            cpu.DE = instruction.Immediate16
        case LoadHLImmediate:
            cpu.Cycles += 3
            cpu.HL = instruction.Immediate16
        case LoadSPImmediate:
            cpu.Cycles += 3
            cpu.SP = instruction.Immediate16
        case StoreBCMemA:
            cpu.Cycles += 2
            cpu.StoreMemory(cpu.BC, cpu.A)
        case StoreDEMemA:
            cpu.Cycles += 2
            cpu.StoreMemory(cpu.DE, cpu.A)
        case StoreHLMemA:
            cpu.Cycles += 2
            cpu.StoreMemory(cpu.HL, cpu.A)
        case StoreSPMemA:
            cpu.Cycles += 2
            cpu.StoreMemory(cpu.SP, cpu.A)
        case LoadAMemBC:
            cpu.Cycles += 2
            cpu.A = cpu.LoadMemory8(cpu.BC)
        case LoadAMemDE:
            cpu.Cycles += 2
            cpu.A = cpu.LoadMemory8(cpu.DE)
        case LoadAMemHL:
            cpu.Cycles += 2
            cpu.A = cpu.LoadMemory8(cpu.HL)
        case LoadAMemSP:
            cpu.Cycles += 2
            cpu.A = cpu.LoadMemory8(cpu.SP)
        case StoreSPMem16:
            cpu.Cycles += 5

            value1 := uint8(cpu.SP & 0xff)
            value2 := uint8((cpu.SP >> 8) & 0xff)

            cpu.StoreMemory(instruction.Immediate16, value1)
            cpu.StoreMemory(instruction.Immediate16+1, value2)

        case JR:
            cpu.Cycles += 3
            offset := int8(instruction.Immediate8)
            cpu.PC = uint16(int32(cpu.PC) + int32(offset) + 2)

        case JrNz:
            if cpu.GetFlagZ() != 0 {
                cpu.Cycles += 3
                offset := int8(instruction.Immediate8)
                cpu.PC = uint16(int32(cpu.PC) + int32(offset) + 2)
            } else {
                cpu.Cycles += 2
            }
        case JrZ:
            if cpu.GetFlagZ() == 0 {
                cpu.Cycles += 3
                offset := int8(instruction.Immediate8)
                cpu.PC = uint16(int32(cpu.PC) + int32(offset) + 2)
            } else {
                cpu.Cycles += 2
            }
        case JrNc:
            if cpu.GetFlagC() != 0 {
                cpu.Cycles += 3
                offset := int8(instruction.Immediate8)
                cpu.PC = uint16(int32(cpu.PC) + int32(offset) + 2)
            } else {
                cpu.Cycles += 2
            }
        case JrC:
            if cpu.GetFlagC() == 0 {
                cpu.Cycles += 3
                offset := int8(instruction.Immediate8)
                cpu.PC = uint16(int32(cpu.PC) + int32(offset) + 2)
            } else {
                cpu.Cycles += 2
            }

        case IncBC:
            cpu.Cycles += 2
            cpu.BC += 1
        case IncDE:
            cpu.Cycles += 2
            cpu.DE += 1
        case IncHL:
            cpu.Cycles += 2
            cpu.HL += 1
        case IncSP:
            cpu.Cycles += 2
            cpu.SP += 1

        case Inc8B:
            cpu.Cycles += 1
            cpu.BC += uint16(1) << 8
            cpu.SetFlagH((uint8(cpu.BC >> 8) >> 4) & 0b1)
            cpu.SetFlagN(0)
            cpu.SetFlagZ(uint8(cpu.BC >> 8))

        case Inc8C:
            cpu.Cycles += 1
            b := cpu.BC >> 8
            c := uint8(cpu.BC & 0xff)
            c += 1
            cpu.SetFlagH((c & 0b10000))
            cpu.SetFlagN(0)
            cpu.SetFlagZ(c)
            cpu.BC = (uint16(b) << 8) | uint16(c)

        case Inc8D:
            cpu.Cycles += 1
            d := uint8(cpu.DE >> 8)
            e := uint8(cpu.DE & 0xff)
            d += 1
            cpu.SetFlagH(d & 0b10000)
            cpu.SetFlagN(0)
            cpu.SetFlagZ(d)
            cpu.DE = (uint16(d) << 8) | uint16(e)

        case Inc8E:
            cpu.Cycles += 1
            d := uint8(cpu.DE >> 8)
            e := uint8(cpu.DE & 0xff)
            e += 1
            cpu.SetFlagH(e & 0b10000)
            cpu.SetFlagN(0)
            cpu.SetFlagZ(e)
            cpu.DE = (uint16(d) << 8) | uint16(e)


        case Inc8H:
            cpu.Cycles += 1
            h := uint8(cpu.HL >> 8)
            l := uint8(cpu.HL & 0xff)
            h += 1
            cpu.SetFlagH(h & 0b10000)
            cpu.SetFlagN(0)
            cpu.SetFlagZ(h)
            cpu.HL = (uint16(h) << 8) | uint16(l)

        case Inc8L:
            cpu.Cycles += 1
            h := uint8(cpu.HL >> 8)
            l := uint8(cpu.HL & 0xff)
            l += 1
            cpu.SetFlagH(l & 0b10000)
            cpu.SetFlagN(0)
            cpu.SetFlagZ(l)
            cpu.HL = (uint16(h) << 8) | uint16(l)

        case Inc8HL:
            cpu.Cycles += 3
            value := cpu.LoadMemory8(cpu.HL)
            value += 1
            cpu.SetFlagH(value & 0b10000)
            cpu.SetFlagN(0)
            cpu.SetFlagZ(value)
            cpu.StoreMemory(cpu.HL, value)

        case Inc8A:
            cpu.Cycles += 1
            a := cpu.A
            a += 1
            cpu.SetFlagH(a & 0b10000)
            cpu.SetFlagN(0)
            cpu.SetFlagZ(a)
            cpu.A = a

        case Dec8B:
            cpu.Cycles += 1
            b := uint8(cpu.BC >> 8)
            c := uint8(cpu.BC & 0xff)
            cpu.SetFlagH(^(b & 0b1111))
            b -= 1
            cpu.SetFlagN(1)
            cpu.SetFlagZ(b)
            cpu.BC = (uint16(b) << 8) | uint16(c)

        case Dec8C:
            cpu.Cycles += 1
            b := uint8(cpu.BC >> 8)
            c := uint8(cpu.BC & 0xff)
            cpu.SetFlagH(^(c & 0b1111))
            c -= 1
            cpu.SetFlagN(1)
            cpu.SetFlagZ(c)
            cpu.BC = (uint16(b) << 8) | uint16(c)

        case Dec8D:
            cpu.Cycles += 1
            d := uint8(cpu.DE >> 8)
            e := uint8(cpu.DE & 0xff)
            cpu.SetFlagH(^(d & 0b1111))
            d -= 1
            cpu.SetFlagN(1)
            cpu.SetFlagZ(d)
            cpu.DE = (uint16(d) << 8) | uint16(e)

        case Dec8E:
            cpu.Cycles += 1
            d := uint8(cpu.DE >> 8)
            e := uint8(cpu.DE & 0xff)
            cpu.SetFlagH(^(e & 0b1111))
            e -= 1
            cpu.SetFlagN(1)
            cpu.SetFlagZ(e)
            cpu.DE = (uint16(d) << 8) | uint16(e)

        case Dec8H:
            cpu.Cycles += 1
            h := uint8(cpu.HL >> 8)
            l := uint8(cpu.HL & 0xff)
            cpu.SetFlagH(^(h & 0b1111))
            h -= 1
            cpu.SetFlagN(1)
            cpu.SetFlagZ(h)
            cpu.HL = (uint16(h) << 8) | uint16(l)

        case Dec8L:
            cpu.Cycles += 1
            h := uint8(cpu.HL >> 8)
            l := uint8(cpu.HL & 0xff)
            cpu.SetFlagH(^(l & 0b1111))
            l -= 1
            cpu.SetFlagN(1)
            cpu.SetFlagZ(l)
            cpu.HL = (uint16(h) << 8) | uint16(l)

        case Dec8HL:
            cpu.Cycles += 3
            value := cpu.LoadMemory8(cpu.HL)
            cpu.SetFlagH(^(value & 0b1111))
            value -= 1
            cpu.SetFlagN(1)
            cpu.SetFlagZ(value)
            cpu.StoreMemory(cpu.HL, value)

        case Dec8A:
            cpu.Cycles += 1
            a := cpu.A
            cpu.SetFlagH(^(a & 0b1111))
            a -= 1
            cpu.SetFlagN(1)
            cpu.SetFlagZ(a)
            cpu.A = a

        case DecBC:
            cpu.Cycles += 2
            cpu.BC -= 1
        case DecDE:
            cpu.Cycles += 2
            cpu.DE -= 1
        case DecHL:
            cpu.Cycles += 2
            cpu.HL -= 1
        case DecSP:
            cpu.Cycles += 2
            cpu.SP -= 1

        case Load8Immediate:
            cpu.Cycles += 2
            cpu.SetRegister8(instruction.R8_1, instruction.Immediate8)

        case StoreHLImmediate:
            cpu.Cycles += 3
            cpu.StoreMemory(cpu.HL, instruction.Immediate8)

        case LoadR8R8:
            cpu.Cycles += 1
            value := cpu.GetRegister8(instruction.R8_2)
            cpu.SetRegister8(instruction.R8_1, value)

        case AddAR8:
            cpu.Cycles += 1
            value := cpu.GetRegister8(instruction.R8_1)
            carry := uint8(0)
            if uint32(cpu.A) + uint32(value) > 0xff {
                carry = 1
            }
            halfCarry := uint8(0)
            if ((cpu.A & 0xf) + (value & 0xf)) & 0x10 == 0x10 {
                halfCarry = 1
            }
            cpu.A += value
            cpu.SetFlagC(carry)
            cpu.SetFlagH(halfCarry)
            cpu.SetFlagZ(cpu.A)
            cpu.SetFlagN(0)

        case AddAImmediate:
            cpu.Cycles += 2
            value := instruction.Immediate8
            carry := uint8(0)
            if uint32(cpu.A) + uint32(value) > 0xff {
                carry = 1
            }
            halfCarry := uint8(0)
            if ((cpu.A & 0xf) + (value & 0xf)) & 0x10 == 0x10 {
                halfCarry = 1
            }
            cpu.A += value
            cpu.SetFlagC(carry)
            cpu.SetFlagH(halfCarry)
            cpu.SetFlagZ(cpu.A)
            cpu.SetFlagN(0)

        case AdcAR8:
            cpu.Cycles += 1
            value := cpu.GetRegister8(instruction.R8_1)
            oldCarry := cpu.GetFlagC()
            carry := uint8(0)
            if uint32(cpu.A) + uint32(value) + uint32(oldCarry) > 0xff {
                carry = 1
            }
            halfCarry := uint8(0)
            if ((cpu.A & 0xf) + (value & 0xf) + oldCarry) & 0x10 == 0x10 {
                halfCarry = 1
            }
            cpu.A += value + oldCarry
            cpu.SetFlagC(carry)
            cpu.SetFlagH(halfCarry)
            cpu.SetFlagZ(cpu.A)
            cpu.SetFlagN(0)

        case AdcAImmediate:
            cpu.Cycles += 2
            value := instruction.Immediate8
            oldCarry := cpu.GetFlagC()
            carry := uint8(0)
            if uint32(cpu.A) + uint32(value) + uint32(oldCarry) > 0xff {
                carry = 1
            }
            halfCarry := uint8(0)
            if ((cpu.A & 0xf) + (value & 0xf) + oldCarry) & 0x10 == 0x10 {
                halfCarry = 1
            }
            cpu.A += value + oldCarry
            cpu.SetFlagC(carry)
            cpu.SetFlagH(halfCarry)
            cpu.SetFlagZ(cpu.A)
            cpu.SetFlagN(0)

        case AddAHLMem:
            cpu.Cycles += 2
            value := cpu.LoadMemory8(cpu.HL)
            carry := uint8(0)
            if uint32(cpu.A) + uint32(value) > 0xff {
                carry = 1
            }
            halfCarry := uint8(0)
            if ((cpu.A & 0xf) + (value & 0xf)) & 0x10 == 0x10 {
                halfCarry = 1
            }
            cpu.A += value
            cpu.SetFlagC(carry)
            cpu.SetFlagH(halfCarry)
            cpu.SetFlagZ(cpu.A)
            cpu.SetFlagN(0)

        case AdcAHLMem:
            cpu.Cycles += 2
            oldCarry := cpu.GetFlagC()
            value := cpu.LoadMemory8(cpu.HL)
            carry := uint8(0)
            if uint32(cpu.A) + uint32(value) + uint32(oldCarry) > 0xff {
                carry = 1
            }
            halfCarry := uint8(0)
            if ((cpu.A & 0xf) + (value & 0xf) + oldCarry) & 0x10 == 0x10 {
                halfCarry = 1
            }
            cpu.A += value + oldCarry
            cpu.SetFlagC(carry)
            cpu.SetFlagH(halfCarry)
            cpu.SetFlagZ(cpu.A)
            cpu.SetFlagN(0)

        case AndAR8:
            cpu.Cycles += 1
            value := cpu.GetRegister8(instruction.R8_1)
            cpu.A &= value
            cpu.SetFlagC(0)
            cpu.SetFlagH(1)
            cpu.SetFlagZ(cpu.A)
            cpu.SetFlagN(0)

        case AndAHLMem:
            cpu.Cycles += 2
            value := cpu.LoadMemory8(cpu.HL)
            cpu.A &= value
            cpu.SetFlagC(0)
            cpu.SetFlagH(1)
            cpu.SetFlagZ(cpu.A)
            cpu.SetFlagN(0)

        case XorAR8:
            cpu.Cycles += 1
            value := cpu.GetRegister8(instruction.R8_1)
            cpu.A ^= value
            cpu.SetFlagC(0)
            cpu.SetFlagH(0)
            cpu.SetFlagZ(cpu.A)
            cpu.SetFlagN(0)

        case XorAHLMem:
            cpu.Cycles += 2
            value := cpu.LoadMemory8(cpu.HL)
            cpu.A ^= value
            cpu.SetFlagC(0)
            cpu.SetFlagH(0)
            cpu.SetFlagZ(cpu.A)
            cpu.SetFlagN(0)

        case OrAR8:
            cpu.Cycles += 1
            value := cpu.GetRegister8(instruction.R8_1)
            cpu.A |= value
            cpu.SetFlagC(0)
            cpu.SetFlagH(0)
            cpu.SetFlagZ(cpu.A)
            cpu.SetFlagN(0)

        case OrAHLMem:
            cpu.Cycles += 2
            value := cpu.LoadMemory8(cpu.HL)
            cpu.A |= value
            cpu.SetFlagC(0)
            cpu.SetFlagH(0)
            cpu.SetFlagZ(cpu.A)
            cpu.SetFlagN(0)

        case SubAR8:
            cpu.Cycles += 1
            value := cpu.GetRegister8(instruction.R8_1)
            carry := uint8(0)
            if value > cpu.A {
                carry = 1
            }
            halfCarry := uint8(0)
            if ((cpu.A & 0xf) - (value & 0xf)) & 0x10 == 0x10 {
                halfCarry = 1
            }

            cpu.A -= value
            cpu.SetFlagN(1)
            cpu.SetFlagC(carry)
            cpu.SetFlagH(halfCarry)
            cpu.SetFlagZ(cpu.A)

        case SubAImmediate:
            cpu.Cycles += 2
            value := instruction.Immediate8
            carry := uint8(0)
            if value > cpu.A {
                carry = 1
            }
            halfCarry := uint8(0)
            if ((cpu.A & 0xf) - (value & 0xf)) & 0x10 == 0x10 {
                halfCarry = 1
            }

            cpu.A -= value
            cpu.SetFlagN(1)
            cpu.SetFlagC(carry)
            cpu.SetFlagH(halfCarry)
            cpu.SetFlagZ(cpu.A)

        case SubAHLMem:
            cpu.Cycles += 2
            value := cpu.LoadMemory8(cpu.HL)
            carry := uint8(0)
            if value > cpu.A {
                carry = 1
            }
            halfCarry := uint8(0)
            if ((cpu.A & 0xf) - (value & 0xf)) & 0x10 == 0x10 {
                halfCarry = 1
            }
            cpu.A -= value
            cpu.SetFlagN(1)
            cpu.SetFlagC(carry)
            cpu.SetFlagH(halfCarry)
            cpu.SetFlagZ(cpu.A)

        case SbcAR8:
            cpu.Cycles += 1
            oldCarry := cpu.GetFlagC()
            value := cpu.GetRegister8(instruction.R8_1)
            carry := uint8(0)
            if value + oldCarry > cpu.A {
                carry = 1
            }
            halfCarry := uint8(0)
            if ((cpu.A & 0xf) - (value & 0xf) - oldCarry) & 0x10 == 0x10 {
                halfCarry = 1
            }

            cpu.A -= value
            cpu.A -= oldCarry
            cpu.SetFlagN(1)
            cpu.SetFlagC(carry)
            cpu.SetFlagH(halfCarry)
            cpu.SetFlagZ(cpu.A)

        case SbcAHLMem:
            cpu.Cycles += 2
            oldCarry := cpu.GetFlagC()
            value := cpu.LoadMemory8(cpu.HL)
            carry := uint8(0)
            if value + oldCarry > cpu.A {
                carry = 1
            }
            halfCarry := uint8(0)
            if ((cpu.A & 0xf) - (value & 0xf) - oldCarry) & 0x10 == 0x10 {
                halfCarry = 1
            }
            cpu.A -= value
            cpu.A -= oldCarry
            cpu.SetFlagN(1)
            cpu.SetFlagC(carry)
            cpu.SetFlagH(halfCarry)
            cpu.SetFlagZ(cpu.A)

        case CpAR8:
            // same as sub, but don't store result
            cpu.Cycles += 1
            value := cpu.GetRegister8(instruction.R8_1)

            carry := uint8(0)
            if value > cpu.A {
                carry = 1
            }
            halfCarry := uint8(0)
            if ((cpu.A & 0xf) - (value & 0xf)) & 0x10 == 0x10 {
                halfCarry = 1
            }

            cpu.SetFlagN(1)
            cpu.SetFlagC(carry)
            cpu.SetFlagH(halfCarry)
            cpu.SetFlagZ(cpu.A)

        case CpAHLMem:
            cpu.Cycles += 2
            value := cpu.LoadMemory8(cpu.HL)
            carry := uint8(0)
            if value > cpu.A {
                carry = 1
            }
            halfCarry := uint8(0)
            if ((cpu.A & 0xf) - (value & 0xf)) & 0x10 == 0x10 {
                halfCarry = 1
            }
            cpu.SetFlagN(1)
            cpu.SetFlagC(carry)
            cpu.SetFlagH(halfCarry)
            cpu.SetFlagZ(cpu.A)

        case AddHLBC:
            cpu.Cycles += 2
            cpu.AddHL(cpu.BC)
        case AddHLDE:
            cpu.Cycles += 2
            cpu.AddHL(cpu.DE)
        case AddHLHL:
            cpu.Cycles += 2
            cpu.AddHL(cpu.HL)
        case AddHLSP:
            cpu.Cycles += 2
            cpu.AddHL(cpu.SP)

        case CPL:
            cpu.Cycles += 1
            cpu.A = ^cpu.A

        case CCF:
            cpu.Cycles += 1
            carry := cpu.GetFlagC()
            cpu.SetFlagC(1 - carry)

        case RLCA:
            cpu.Cycles += 1

            newA, carry := RotateLeft(cpu.A)
            cpu.A = newA
            cpu.SetFlagZ(0)
            cpu.SetFlagH(0)
            cpu.SetFlagN(0)
            cpu.SetFlagC(carry)

        case RRCA:
            cpu.Cycles += 1

            newA, carry := RotateRight(cpu.A)
            cpu.A = newA
            cpu.SetFlagZ(0)
            cpu.SetFlagH(0)
            cpu.SetFlagN(0)
            cpu.SetFlagC(carry)

        case RLA:
            cpu.Cycles += 1

            oldCarry := cpu.GetFlagC()
            newCarry := (cpu.A >> 7) & 0b1
            cpu.A = (cpu.A << 1) | oldCarry

            cpu.SetFlagZ(0)
            cpu.SetFlagH(0)
            cpu.SetFlagN(0)
            cpu.SetFlagC(newCarry)

        case RRA:
            cpu.Cycles += 1

            oldCarry := cpu.GetFlagC()
            newCarry := cpu.A & 0b1
            cpu.A = (cpu.A >> 1) | (oldCarry << 7)

            cpu.SetFlagZ(0)
            cpu.SetFlagH(0)
            cpu.SetFlagN(0)
            cpu.SetFlagC(newCarry)

        case Stop:
            cpu.Stopped = true

        case Halt:
            cpu.Halted = true

        case DAA:
            // BCD fixup after add/subtract

            // https://blog.ollien.com/posts/gb-daa/
            // https://ehaskins.com/2018-01-30%20Z80%20DAA/
            cpu.Cycles += 1
            log.Printf("DAA not implemented")

        case SCF:
            cpu.Cycles += 1
            cpu.SetFlagC(1)

        default:
            log.Printf("Execute error: unknown opcode %v", instruction.Opcode)
    }
}

// little-endian 16-bit immediate
func makeImm16(data []byte) uint16 {
    if len(data) < 2 {
        panic("makeImm16: data too short")
    }
    return uint16(data[0]) | (uint16(data[1]) << 8)
}

func makeLoadR16Imm16Instruction(r16 R16, immediate uint16) Instruction {
    switch r16 {
        case R16BC: return Instruction{Opcode: LoadBCImmediate, Immediate16: immediate}
        case R16DE: return Instruction{Opcode: LoadDEImmediate, Immediate16: immediate}
        case R16HL: return Instruction{Opcode: LoadHLImmediate, Immediate16: immediate}
        case R16SP: return Instruction{Opcode: LoadSPImmediate, Immediate16: immediate}
    }

    return Instruction{Opcode: Unknown}
}

func makeStoreR16MemAInstruction(r16 R16) Instruction {
    switch r16 {
        case R16BC: return Instruction{Opcode: StoreBCMemA}
        case R16DE: return Instruction{Opcode: StoreDEMemA}
        case R16HL: return Instruction{Opcode: StoreHLMemA}
        case R16SP: return Instruction{Opcode: StoreSPMemA}
    }

    return Instruction{Opcode: Unknown}
}

func makeLoadAFromR16MemInstruction(r16 R16) Instruction {
    switch r16 {
        case R16BC: return Instruction{Opcode: LoadAMemBC}
        case R16DE: return Instruction{Opcode: LoadAMemDE}
        case R16HL: return Instruction{Opcode: LoadAMemHL}
        case R16SP: return Instruction{Opcode: LoadAMemSP}
    }

    return Instruction{Opcode: Unknown}
}

func makeIncInstruction(r16 R16) Instruction {
    switch r16 {
        case R16BC: return Instruction{Opcode: IncBC}
        case R16DE: return Instruction{Opcode: IncDE}
        case R16HL: return Instruction{Opcode: IncHL}
        case R16SP: return Instruction{Opcode: IncSP}
    }

    return Instruction{Opcode: Unknown}
}

func makeDecInstruction(r16 R16) Instruction {
    switch r16 {
        case R16BC: return Instruction{Opcode: DecBC}
        case R16DE: return Instruction{Opcode: DecDE}
        case R16HL: return Instruction{Opcode: DecHL}
        case R16SP: return Instruction{Opcode: DecSP}
    }

    return Instruction{Opcode: Unknown}
}

func makeAddHLR16Instruction(r16 R16) Instruction {
    switch r16 {
        case R16BC: return Instruction{Opcode: AddHLBC}
        case R16DE: return Instruction{Opcode: AddHLDE}
        case R16HL: return Instruction{Opcode: AddHLHL}
        case R16SP: return Instruction{Opcode: AddHLSP}
    }

    return Instruction{Opcode: Unknown}
}

func makeIncR8Instruction(r8 R8) Instruction {
    switch r8 {
        case R8B: return Instruction{Opcode: Inc8B}
        case R8C: return Instruction{Opcode: Inc8C}
        case R8D: return Instruction{Opcode: Inc8D}
        case R8E: return Instruction{Opcode: Inc8E}
        case R8H: return Instruction{Opcode: Inc8H}
        case R8L: return Instruction{Opcode: Inc8L}
        case R8HL: return Instruction{Opcode: Inc8HL}
        case R8A: return Instruction{Opcode: Inc8A}
    }

    return Instruction{Opcode: Unknown}
}

func makeDecR8Instruction(r8 R8) Instruction {
    switch r8 {
        case R8B: return Instruction{Opcode: Dec8B}
        case R8C: return Instruction{Opcode: Dec8C}
        case R8D: return Instruction{Opcode: Dec8D}
        case R8E: return Instruction{Opcode: Dec8E}
        case R8H: return Instruction{Opcode: Dec8H}
        case R8L: return Instruction{Opcode: Dec8L}
        case R8HL: return Instruction{Opcode: Dec8HL}
        case R8A: return Instruction{Opcode: Dec8A}
    }

    return Instruction{Opcode: Unknown}
}

// instructions should be at least 3 bytes long for 'opcode immediate immediate'
func DecodeInstruction(instructions []byte) (Instruction, uint8) {
    instruction := instructions[0]
    block := instruction >> 6
    // check top 2 bits first
    switch block {
        case 0b00:
            switch instruction & 0b1111 {
                case 0b0000: return Instruction{Opcode: Nop}, 1
                case 0b0001:
                    r16 := (instruction >> 4) & 0b11
                    return makeLoadR16Imm16Instruction(R16(r16), makeImm16(instructions[1:])), 3
                case 0b0010:
                    //return "ld [r16mem], a"
                    r16 := R16((instruction >> 4) & 0b11)
                    return makeStoreR16MemAInstruction(r16), 1
                case 0b1010:
                    r16 := R16((instruction >> 4) & 0b11)
                    return makeLoadAFromR16MemInstruction(r16), 1
                    //return "ld a, [r16mem]"

                case 0b1000:
                    immediate := makeImm16(instructions[1:])
                    return Instruction{Opcode: StoreSPMem16, Immediate16: immediate}, 3

                    //return "ld [imm16], sp"

                case 0b0011:
                    r16 := R16((instruction >> 4) & 0b11)
                    return makeIncInstruction(r16), 1

                    // return "inc r16"

                case 0b1011:
                    r16 := R16((instruction >> 4) & 0b11)
                    return makeDecInstruction(r16), 1

                case 0b1001:
                    r16 := R16((instruction >> 4) & 0b11)
                    return makeAddHLR16Instruction(r16), 1

                    // return "add hl, r16"

                case 0b0111:
                    switch instruction >> 4 {
                        case 0b0000: return Instruction{Opcode: RLCA}, 1
                        case 0b0001: return Instruction{Opcode: RLA}, 1
                        case 0b0010: return Instruction{Opcode: DAA}, 1
                        case 0b0011: return Instruction{Opcode: SCF}, 1
                    }

                case 0b1111: 
                    switch instruction >> 4 {
                        case 0b0000: return Instruction{Opcode: RRCA}, 1
                        case 0b0001: return Instruction{Opcode: RRA}, 1
                        case 0b0010: return Instruction{Opcode: CPL}, 1
                        case 0b0011: return Instruction{Opcode: CCF}, 1
                    }
            }

            switch instruction & 0b111 {
                case 0b000:
                    if instruction == 0b00011000 {
                        return Instruction{Opcode: JR, Immediate8: instructions[1]}, 2
                    }

                    if instruction == 0b00010000 {
                        return Instruction{Opcode: Stop}, 2
                    }

                    if instruction >> 5 == 0b001 {
                        cond := (instruction >> 3) & 0b11
                        opcode := JrNz

                        switch cond {
                            case 0b00: opcode = JrNz
                            case 0b01: opcode = JrZ
                            case 0b10: opcode = JrNc
                            case 0b11: opcode = JrC
                        }

                        return Instruction{Opcode: opcode, Immediate8: instructions[1]}, 2
                    }

                case 0b100:
                    r8 := R8((instruction >> 3) & 0b111)
                    return makeIncR8Instruction(r8), 1

                case 0b101:
                    r8 := R8((instruction >> 3) & 0b111)
                    return makeDecR8Instruction(r8), 1

                case 0b110:
                    r8 := R8((instruction >> 3) & 0b111)
                    if r8 == R8HL {
                        return Instruction{Opcode: StoreHLImmediate, Immediate8: instructions[1]}, 2
                    }

                    return Instruction{Opcode: Load8Immediate, R8_1: r8, Immediate8: instructions[1]}, 2
            }

        case 0b01:
            if instruction & 0b11111 == 0b110110 {
                return Instruction{Opcode: Halt}, 1
            }

            source := R8(instruction & 0b111)
            dest := R8((instruction >> 3) & 0b111)

            return Instruction{Opcode: LoadR8R8, R8_1: dest, R8_2: source}, 1

        case 0b10:
            switch (instruction >> 3) & 0b111 {
                case 0b000:
                    r8 := R8(instruction & 0b111)
                    if r8 == R8HL {
                        return Instruction{Opcode: AddAHLMem}, 1
                    }
                    return Instruction{Opcode: AddAR8, R8_1: r8}, 1
                case 0b001:
                    r8 := R8(instruction & 0b111)
                    if r8 == R8HL {
                        return Instruction{Opcode: AdcAHLMem}, 1
                    }
                    return Instruction{Opcode: AdcAR8, R8_1: r8}, 1

                case 0b010:
                    r8 := R8(instruction & 0b111)
                    if r8 == R8HL {
                        return Instruction{Opcode: SubAHLMem}, 1
                    }
                    return Instruction{Opcode: SubAR8, R8_1: r8}, 1

                case 0b011:
                    r8 := R8(instruction & 0b111)
                    if r8 == R8HL {
                        return Instruction{Opcode: SbcAHLMem}, 1
                    }

                    return Instruction{Opcode: SbcAR8, R8_1: r8}, 1

                case 0b100:
                    r8 := R8(instruction & 0b111)
                    if r8 == R8HL {
                        return Instruction{Opcode: AndAHLMem}, 1
                    }

                    return Instruction{Opcode: AndAR8, R8_1: r8}, 1

                case 0b101:
                    r8 := R8(instruction & 0b111)
                    if r8 == R8HL {
                        return Instruction{Opcode: XorAHLMem}, 1
                    }

                    return Instruction{Opcode: XorAR8, R8_1: r8}, 1

                case 0b110:
                    r8 := R8(instruction & 0b111)
                    if r8 == R8HL {
                        return Instruction{Opcode: OrAHLMem}, 1
                    }

                    return Instruction{Opcode: OrAR8, R8_1: r8}, 1

                case 0b111:
                    r8 := R8(instruction & 0b111)
                    if r8 == R8HL {
                        return Instruction{Opcode: CpAHLMem}, 1
                    }
                    return Instruction{Opcode: CpAR8, R8_1: r8}, 1
            }
        case 0b11:
            switch instruction & 0b111111 {
                case 0b000110:
                    return Instruction{Opcode: AddAImmediate, Immediate8: instructions[1]}, 2
                case 0b001110:
                    return Instruction{Opcode: AdcAImmediate, Immediate8: instructions[1]}, 2
                case 0b010110:
                    return Instruction{Opcode: SubAImmediate, Immediate8: instructions[1]}, 2

                /*
                case 0b011110: return "sbc a, imm8"
                case 0b100110: return "and a, imm8"
                case 0b101110: return "xor a, imm8"
                case 0b110110: return "or a, imm8"
                case 0b111110: return "cp a, imm8"

                case 0b100010: return "ldh [c], a"
                case 0b100000: return "ldh [imm8], a"
                case 0b101010: return "ld [imm16], a"
                case 0b110010: return "ldh a, [c]"
                case 0b110000: return "ldh a, [imm8]"
                case 0b111010: return "ld a, [imm16]"

                case 0b101000: return "add sp, imm8"
                case 0b111000: return "ld hl, sp + imm8"
                case 0b111001: return "ld sp, hl"

                case 0b110011: return "di"
                case 0b111011: return "ei"
                */
            }

            /*
            switch instruction & 0b1111 {
                case 0b0001: return "pop r16stk"
                case 0b0101: return "push r16stk"
            }

            if instruction >> 5 == 0b110 {
                switch instruction & 0b111 {
                    case 0b000: return "ret cond"
                    case 0b010: return "jp cond, imm16"
                    case 0b011: return "jp imm16"
                    case 0b100: return "call cond, imm16"
                    case 0b111: return "rst tgt3"
                }

                switch instruction {
                    case 0b11001001: return "ret"
                    case 0b11011001: return "reti"
                    case 0b11101001: return "jp hl"
                    case 0b11001101: return "call imm16"
                }
            }

            if instruction == 0xcb {
                // special prefix instruction
            }
            */
    }

    return Instruction{Opcode: Unknown}, 1
}
