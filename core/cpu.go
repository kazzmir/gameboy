package core

import (
    "log"
    "fmt"
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

    InterruptFlag bool // IME

    Stopped bool
    Halted bool

    Ram []uint8
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
    LdHlSpImmediate8
    LoadR8R8

    StoreBCMemA
    StoreDEMemA
    StoreHLIncMemA
    StoreHLDecMemA

    DisableInterrupts
    EnableInterrupts

    LoadAMemBC
    LoadAMemDE
    LoadAMemHLI
    LoadAMemHLD

    LdSpHl

    LdhCA
    LdhAC
    LdhImmediate8A
    LdhAImmediate8

    LdImmediate16A
    LdAImmediate16

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

    AddSpImmediate8

    AddAImmediate
    AdcAImmediate
    SubAImmediate
    SbcAImmediate
    AndAImmediate
    XorAImmediate
    OrAImmediate
    CpAImmediate

    AddAR8
    SubAR8
    AdcAR8
    AndAR8
    XorAR8
    OrAR8
    SbcAR8

    CpAR8

    PopAF
    PopR16
    PushR16
    PushAF

    RLCA
    RLC
    RLA
    RL
    RRCA
    RRC
    RRA
    RR

    SLA
    SRA
    SRL
    SWAP

    Bit
    Res
    Set

    DAA

    SCF

    CPL
    CCF

    JR

    JrNz
    JrZ
    JrNc
    JrC

    CallNzImmediate16
    CallZImmediate16
    CallNcImmediate16
    CallCImmediate16

    CallImmediate16
    CallResetVector

    Return
    ReturnFromInterrupt

    JpImmediate16
    JpHL

    JpNzImmediate16
    JpZImmediate16
    JpNcImmediate16
    JpCImmediate16

    RetNz
    RetZ
    RetNc
    RetC

    Stop
    Halt

    Unknown
)

func (opcode Opcode) String() string {
    switch opcode {
        case Nop: return "nop"
        case LoadBCImmediate: return "ld bc, nn"
        case LoadDEImmediate: return "ld de, nn"
        case LoadHLImmediate: return "ld hl, nn"
        case LoadSPImmediate: return "ld sp, nn"

        case Load8Immediate: return "ld r8, n"
        case StoreHLImmediate: return "ld (hl), n"

        case LdHlSpImmediate8: return "ld hl, sp+n"

        case LoadR8R8: return "ld r8, r8"

        case StoreBCMemA: return "ld (bc), a"
        case StoreDEMemA: return "ld (de), a"
        case StoreHLIncMemA: return "ld (hli), a"
        case StoreHLDecMemA: return "ld (hld), a"

        case IncBC: return "inc bc"
        case IncDE: return "inc de"
        case IncHL: return "inc hl"
        case IncSP: return "inc sp"

        case Inc8B: return "inc b"
        case Inc8C: return "inc c"
        case Inc8D: return "inc d"
        case Inc8E: return "inc e"
        case Inc8H: return "inc h"
        case Inc8L: return "inc l"
        case Inc8HL: return "inc (hl)"
        case Inc8A: return "inc a"

        case Dec8B: return "dec b"
        case Dec8C: return "dec c"
        case Dec8D: return "dec d"
        case Dec8E: return "dec e"
        case Dec8H: return "dec h"
        case Dec8L: return "dec l"
        case Dec8HL: return "dec (hl)"
        case Dec8A: return "dec a"

        case AddAR8: return "add a, r8"
        case SubAR8: return "sub a, r8"
        case SbcAR8: return "sbc a, r8"
        case AdcAR8: return "adc a, r8"
        case AndAR8: return "and a, r8"
        case XorAR8: return "xor a, r8"
        case OrAR8: return "or a, r8"
        case CpAR8: return "cp a, r8"

        case AddAImmediate: return "add a, n"
        case AdcAImmediate: return "adc a, n"
        case SubAImmediate: return "sub a, n"
        case SbcAImmediate: return "sbc a, n"
        case AndAImmediate: return "and a, n"
        case XorAImmediate: return "xor a, n"
        case OrAImmediate: return "or a, n"
        case CpAImmediate: return "cp a, n"

        case PopAF: return "pop af"
        case PopR16: return "pop r16"
        case PushR16: return "push r16"
        case PushAF: return "push af"

        case DecBC: return "dec bc"
        case DecDE: return "dec de"
        case DecHL: return "dec hl"
        case DecSP: return "dec sp"

        case CPL: return "cpl"
        case CCF: return "ccf"

        case RLCA: return "rlca"
        case RLC: return "rlc"
        case RLA: return "rla"
        case RL: return "rl"
        case RRCA: return "rrca"
        case RRC: return "rrc"
        case RRA: return "rra"
        case RR: return "rr"

        case SLA: return "sla"
        case SRA: return "sra"
        case SRL: return "srl"
        case SWAP: return "swap"

        case Bit: return "bit"
        case Res: return "res"
        case Set: return "set"

        case SCF: return "scf"

        case Stop: return "stop"
        case Halt: return "halt"

        case RetNz: return "ret nz"
        case RetZ: return "ret z"
        case RetNc: return "ret nc"
        case RetC: return "ret c"

        case JpImmediate16: return "jp nn"
        case JpHL: return "jp hl"

        case JpNzImmediate16: return "jp nz, nn"
        case JpZImmediate16: return "jp z, nn"
        case JpNcImmediate16: return "jp nc, nn"
        case JpCImmediate16: return "jp c, nn"

        case JR: return "jr n"
        case JrNz: return "jr nz, n"
        case JrZ: return "jr z, n"
        case JrNc: return "jr nc, n"
        case JrC: return "jr c, n"

        case CallNzImmediate16: return "call nz, nn"
        case CallZImmediate16: return "call z, nn"
        case CallNcImmediate16: return "call nc, nn"
        case CallCImmediate16: return "call c, nn"

        case CallImmediate16: return "call nn"
        case CallResetVector: return "rst n"

        case Return: return "ret"
        case ReturnFromInterrupt: return "reti"

        case DAA: return "daa"

        case StoreSPMem16: return "ld (nn), sp"

        case AddHLBC: return "add hl, bc"
        case AddHLDE: return "add hl, de"
        case AddHLHL: return "add hl, hl"
        case AddHLSP: return "add hl, sp"

        case AddSpImmediate8: return "add sp, n"

        case LdImmediate16A: return "ld (nn), a"
        case LdAImmediate16: return "ld a, (nn)"

        case LoadAMemBC: return "ld a, (bc)"
        case LoadAMemDE: return "ld a, (de)"
        case LoadAMemHLI: return "ld a, (hl+)"
        case LoadAMemHLD: return "ld a, (hl-)"

        case LdSpHl: return "ld sp, hl"
        case LdhCA: return "ldh (c), a"
        case LdhAC: return "ldh a, (c)"
        case LdhImmediate8A: return "ldh (n), a"
        case LdhAImmediate8: return "ldh a, (n)"

        // todo rest

        case Unknown: return "unknown"
    }

    return fmt.Sprintf("? %v", int(opcode))
}

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
func setBit(value uint8, bit uint8, set bool) uint8 {
    if !set {
        return value & ^(1 << bit)
    }
    return value | (1 << bit)
}

func (cpu *CPU) SetFlagC(on bool) {
    cpu.F = setBit(cpu.F, 4, on)
}

func (cpu *CPU) GetFlagC() uint8 {
    return (cpu.F >> 4) & 0b1
}

func (cpu *CPU) GetFlagN() uint8 {
    return (cpu.F >> 6) & 0b1
}

func (cpu *CPU) GetFlagZ() uint8 {
    return (cpu.F >> 7) & 0b1
}

func (cpu *CPU) GetFlagH() uint8 {
    return (cpu.F >> 5) & 0b1
}

func (cpu *CPU) SetFlagH(on bool) {
    cpu.F = setBit(cpu.F, 5, on)
}

func (cpu *CPU) SetFlagN(on bool) {
    cpu.F = setBit(cpu.F, 6, on)
}

func (cpu *CPU) SetFlagZ(on bool) {
    cpu.F = setBit(cpu.F, 7, on)
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
    cpu.Ram[address] = value
}

func (cpu *CPU) LoadMemory8(address uint16) uint8 {
    return cpu.Ram[address]
}

func (cpu *CPU) LoadMemory16(address uint16) uint16 {
    low := cpu.LoadMemory8(address)
    high := cpu.LoadMemory8(address+1)
    return (uint16(high) << 8) | uint16(low)
}

func (cpu *CPU) AddHL(value uint16) {
    carry := uint32(cpu.HL) + uint32(value) > 0xffff

    halfCarry := ((cpu.HL & 0xfff) + (value & 0xfff)) & 0x1000 == 0x1000

    cpu.HL += value
    cpu.SetFlagN(false)

    if halfCarry {
        cpu.SetFlagH(true)
    } else {
        cpu.SetFlagH(false)
    }

    if carry {
        cpu.SetFlagC(true)
    } else {
        cpu.SetFlagC(false)
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

func (cpu *CPU) Pop16() uint16 {
    low := cpu.LoadMemory8(cpu.SP)
    cpu.SP += 1
    high := cpu.LoadMemory8(cpu.SP)
    cpu.SP += 1
    return (uint16(high) << 8) | uint16(low)
}

func (cpu *CPU) Push16(value uint16) {
    low := uint8(value & 0xff)
    high := uint8((value >> 8) & 0xff)

    cpu.SP -= 1
    cpu.StoreMemory(cpu.SP, high)
    cpu.SP -= 1
    cpu.StoreMemory(cpu.SP, low)
}

func (cpu *CPU) doRetCond(cond bool) {
    if cond {
        cpu.Cycles += 5
        cpu.PC = cpu.Pop16()
    } else {
        cpu.Cycles += 2
        cpu.PC += 1
    }
}

func (cpu *CPU) doCallCond(address uint16, cond bool) {
    if cond {
        cpu.Cycles += 6

        // push address of instruction after call onto stack
        returnAddress := cpu.PC + 3
        cpu.Push16(returnAddress)
        cpu.PC = address

    } else {
        cpu.Cycles += 3
        cpu.PC += 3
    }
}

func (cpu *CPU) doJrCond(offset int8, cond bool) {
    if cond {
        cpu.Cycles += 3
        cpu.PC = uint16(int32(cpu.PC) + int32(offset) + 2)
    } else {
        cpu.Cycles += 2
        cpu.PC += 2
    }
}

func (cpu *CPU) doJpCond(address uint16, cond bool) {
    if cond {
        cpu.Cycles += 4
        cpu.PC = address
    } else {
        cpu.Cycles += 3
        cpu.PC += 3
    }
}

func (cpu *CPU) doAddA(value uint8) {
    carry := uint8(0)
    if uint32(cpu.A) + uint32(value) > 0xff {
        carry = 1
    }
    halfCarry := uint8(0)
    if ((cpu.A & 0xf) + (value & 0xf)) & 0x10 == 0x10 {
        halfCarry = 1
    }
    cpu.A += value
    cpu.SetFlagC(carry == 1)
    cpu.SetFlagH(halfCarry == 1)
    cpu.SetFlagZ(cpu.A == 0)
    cpu.SetFlagN(false)
}

func (cpu *CPU) doAdcA(value uint8) {
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
    cpu.SetFlagC(carry == 1)
    cpu.SetFlagH(halfCarry == 1)
    cpu.SetFlagZ(cpu.A == 0)
    cpu.SetFlagN(false)
}

func (cpu *CPU) doSbcA(value uint8) {
    oldCarry := cpu.GetFlagC()

    carry := uint8(0)
    if uint16(value) + uint16(oldCarry) > uint16(cpu.A) {
        carry = 1
    }

    halfCarry := uint8(0)
    if ((cpu.A & 0xf) - (value & 0xf) - oldCarry) & 0x10 == 0x10 {
        halfCarry = 1
    }

    cpu.A -= value
    cpu.A -= oldCarry
    cpu.SetFlagN(true)
    cpu.SetFlagC(carry == 1)
    cpu.SetFlagH(halfCarry == 1)
    cpu.SetFlagZ(cpu.A == 0)
}

func (cpu *CPU) doCpA(value uint8) {
    carry := uint8(0)
    if value > cpu.A {
        carry = 1
    }
    halfCarry := uint8(0)
    if ((cpu.A & 0xf) - (value & 0xf)) & 0x10 == 0x10 {
        halfCarry = 1
    }

    cpu.SetFlagN(true)
    cpu.SetFlagC(carry == 1)
    cpu.SetFlagH(halfCarry == 1)
    cpu.SetFlagZ(cpu.A - value == 0)
}

func (cpu *CPU) doAndA(value uint8) {
    cpu.A &= value
    cpu.SetFlagC(false)
    cpu.SetFlagH(true)
    cpu.SetFlagZ(cpu.A == 0)
    cpu.SetFlagN(false)
}

func (cpu *CPU) doSubA(value uint8) {
    carry := uint8(0)
    if value > cpu.A {
        carry = 1
    }
    halfCarry := uint8(0)
    if ((cpu.A & 0xf) - (value & 0xf)) & 0x10 == 0x10 {
        halfCarry = 1
    }

    cpu.A -= value
    cpu.SetFlagN(true)
    cpu.SetFlagC(carry == 1)
    cpu.SetFlagH(halfCarry == 1)
    cpu.SetFlagZ(cpu.A == 0)
}

func (cpu *CPU) doOrA(value uint8) {
    cpu.A |= value
    cpu.SetFlagC(false)
    cpu.SetFlagH(false)
    cpu.SetFlagZ(cpu.A == 0)
    cpu.SetFlagN(false)
}

func (cpu *CPU) doXorA(value uint8) {
    cpu.A ^= value
    cpu.SetFlagC(false)
    cpu.SetFlagH(false)
    cpu.SetFlagZ(cpu.A == 0)
    cpu.SetFlagN(false)
}

func (cpu *CPU) Execute(instruction Instruction) {
    // log.Printf("Executing instruction: %+v", instruction)
    switch instruction.Opcode {
        case Nop:
            cpu.Cycles += 1
            cpu.PC += 1
        case LoadBCImmediate:
            cpu.Cycles += 3
            cpu.BC = instruction.Immediate16
            cpu.PC += 3
        case LoadDEImmediate:
            cpu.Cycles += 3
            cpu.DE = instruction.Immediate16
            cpu.PC += 3
        case LoadHLImmediate:
            cpu.Cycles += 3
            cpu.HL = instruction.Immediate16
            cpu.PC += 3
        case LoadSPImmediate:
            cpu.Cycles += 3
            cpu.SP = instruction.Immediate16
            cpu.PC += 3
        case StoreBCMemA:
            cpu.Cycles += 2
            cpu.StoreMemory(cpu.BC, cpu.A)
            cpu.PC += 1
        case StoreDEMemA:
            cpu.Cycles += 2
            cpu.StoreMemory(cpu.DE, cpu.A)
            cpu.PC += 1
        case StoreHLIncMemA:
            cpu.Cycles += 2
            cpu.StoreMemory(cpu.HL, cpu.A)
            cpu.HL += 1
            cpu.PC += 1
        case StoreHLDecMemA:
            cpu.Cycles += 2
            cpu.StoreMemory(cpu.HL, cpu.A)
            cpu.HL -= 1
            cpu.PC += 1
        case LoadAMemBC:
            cpu.Cycles += 2
            cpu.A = cpu.LoadMemory8(cpu.BC)
            cpu.PC += 1
        case LoadAMemDE:
            cpu.Cycles += 2
            cpu.A = cpu.LoadMemory8(cpu.DE)
            cpu.PC += 1
        case LoadAMemHLI:
            cpu.Cycles += 2
            cpu.A = cpu.LoadMemory8(cpu.HL)
            cpu.HL += 1
            cpu.PC += 1
        case LoadAMemHLD:
            cpu.Cycles += 2
            cpu.A = cpu.LoadMemory8(cpu.HL)
            cpu.HL -= 1
            cpu.PC += 1
        case StoreSPMem16:
            cpu.Cycles += 5

            value1 := uint8(cpu.SP & 0xff)
            value2 := uint8((cpu.SP >> 8) & 0xff)

            cpu.StoreMemory(instruction.Immediate16, value1)
            cpu.StoreMemory(instruction.Immediate16+1, value2)

            cpu.PC += 3

        case LdSpHl:
            cpu.Cycles += 2
            cpu.HL = cpu.SP

        case LdhCA:
            cpu.Cycles += 2
            cpu.PC += 1
            address := 0xff00 + uint16(cpu.GetRegister8(R8C))
            cpu.StoreMemory(address, cpu.A)

        case LdhAC:
            cpu.Cycles += 2
            address := 0xff00 + uint16(cpu.GetRegister8(R8C))
            cpu.A = cpu.LoadMemory8(address)
            cpu.PC += 1

        case LdhImmediate8A:
            cpu.Cycles += 2
            cpu.PC += 2
            address := 0xff00 + uint16(instruction.Immediate8)
            cpu.StoreMemory(address, cpu.A)

        case LdhAImmediate8:
            cpu.Cycles += 3
            address := 0xff00 + uint16(instruction.Immediate8)
            cpu.A = cpu.LoadMemory8(address)
            cpu.PC += 2

        case LdImmediate16A:
            cpu.Cycles += 4
            cpu.StoreMemory(instruction.Immediate16, cpu.A)
            cpu.PC += 3

        case LdAImmediate16:
            cpu.Cycles += 4
            cpu.A = cpu.LoadMemory8(instruction.Immediate16)
            cpu.PC += 3

        case JR:
            cpu.Cycles += 3
            offset := int8(instruction.Immediate8)
            cpu.PC = uint16(int32(cpu.PC) + int32(offset) + 2)

        case CallNzImmediate16:
            cpu.doCallCond(instruction.Immediate16, cpu.GetFlagZ() == 0)
        case CallZImmediate16:
            cpu.doCallCond(instruction.Immediate16, cpu.GetFlagZ() != 0)
        case CallNcImmediate16:
            cpu.doCallCond(instruction.Immediate16, cpu.GetFlagC() == 0)
        case CallCImmediate16:
            cpu.doCallCond(instruction.Immediate16, cpu.GetFlagC() != 0)

        case CallImmediate16:
            cpu.Cycles += 6
            address := instruction.Immediate16
            returnAddress := cpu.PC + 3
            cpu.Push16(returnAddress)
            cpu.PC = address

        case CallResetVector:
            cpu.Cycles += 4

            address := instruction.Immediate8

            // push address of instruction after call onto stack
            returnAddress := cpu.PC + 1
            cpu.Push16(returnAddress)
            cpu.PC = uint16(address)

        // return from a call
        case Return:
            cpu.Cycles += 4
            cpu.PC = cpu.Pop16()

        case ReturnFromInterrupt:
            cpu.Cycles += 4
            cpu.PC = cpu.Pop16()
            cpu.InterruptFlag = true

        case JrNz:
            cpu.doJrCond(int8(instruction.Immediate8), cpu.GetFlagZ() == 0)
        case JrZ:
            cpu.doJrCond(int8(instruction.Immediate8), cpu.GetFlagZ() == 1)
        case JrNc:
            cpu.doJrCond(int8(instruction.Immediate8), cpu.GetFlagC() == 0)
        case JrC:
            cpu.doJrCond(int8(instruction.Immediate8), cpu.GetFlagC() == 1)

        case JpHL:
            cpu.Cycles += 1
            cpu.PC = cpu.HL

        case JpImmediate16:
            cpu.Cycles += 4
            cpu.PC = instruction.Immediate16

        case JpNzImmediate16:
            cpu.doJpCond(instruction.Immediate16, cpu.GetFlagZ() == 0)
        case JpZImmediate16:
            cpu.doJpCond(instruction.Immediate16, cpu.GetFlagZ() == 1)
        case JpNcImmediate16:
            cpu.doJpCond(instruction.Immediate16, cpu.GetFlagC() == 0)
        case JpCImmediate16:
            cpu.doJpCond(instruction.Immediate16, cpu.GetFlagC() == 1)

        case RetNz:
            cpu.doRetCond(cpu.GetFlagZ() == 0)
        case RetZ:
            cpu.doRetCond(cpu.GetFlagZ() == 1)
        case RetNc:
            cpu.doRetCond(cpu.GetFlagC() == 0)
        case RetC:
            cpu.doRetCond(cpu.GetFlagC() == 1)

        case DisableInterrupts:
            cpu.Cycles += 1
            cpu.InterruptFlag = false

        case EnableInterrupts:
            cpu.Cycles += 1
            cpu.InterruptFlag = true

        case IncBC:
            cpu.Cycles += 2
            cpu.BC += 1
            cpu.PC += 1
        case IncDE:
            cpu.Cycles += 2
            cpu.DE += 1
            cpu.PC += 1
        case IncHL:
            cpu.Cycles += 2
            cpu.HL += 1
            cpu.PC += 1
        case IncSP:
            cpu.Cycles += 2
            cpu.SP += 1
            cpu.PC += 1

        case Inc8B:
            cpu.Cycles += 1

            b := uint8(cpu.BC >> 8)
            lower := b & 0b1111

            cpu.BC += uint16(1) << 8
            h := uint8(0)
            if lower == 0b1111 {
                h = 1
            }
            cpu.SetFlagH(h == 1)
            cpu.SetFlagN(false)
            cpu.SetFlagZ(cpu.BC >> 8 == 0)
            cpu.PC += 1

        case Inc8C:
            cpu.Cycles += 1
            b := cpu.BC >> 8
            c := uint8(cpu.BC & 0xff)

            lower := c & 0b1111

            h := uint8(0)
            if lower == 0b1111 {
                h = 1
            }
            cpu.SetFlagH(h == 1)

            c += 1
            cpu.SetFlagN(false)
            cpu.SetFlagZ(c == 0)

            cpu.BC = (uint16(b) << 8) | uint16(c)
            cpu.PC += 1

        case Inc8D:
            cpu.Cycles += 1
            d := uint8(cpu.DE >> 8)
            e := uint8(cpu.DE & 0xff)

            lower := d & 0b1111

            h := uint8(0)
            if lower == 0b1111 {
                h = 1
            }
            cpu.SetFlagH(h == 1)

            d += 1
            cpu.SetFlagN(false)
            cpu.SetFlagZ(d == 0)
            cpu.DE = (uint16(d) << 8) | uint16(e)
            cpu.PC += 1

        case Inc8E:
            cpu.Cycles += 1
            d := uint8(cpu.DE >> 8)
            e := uint8(cpu.DE & 0xff)

            h := uint8(0)
            if e & 0b1111 == 0b1111 {
                h = 1
            }
            cpu.SetFlagH(h == 1)

            e += 1
            cpu.SetFlagN(false)
            cpu.SetFlagZ(e == 0)
            cpu.DE = (uint16(d) << 8) | uint16(e)
            cpu.PC += 1

        case Inc8H:
            cpu.Cycles += 1
            h := uint8(cpu.HL >> 8)
            l := uint8(cpu.HL & 0xff)

            carry := uint8(0)
            if h & 0b1111 == 0b1111 {
                carry = 1
            }
            cpu.SetFlagH(carry == 1)

            h += 1

            cpu.SetFlagN(false)
            cpu.SetFlagZ(h == 0)
            cpu.HL = (uint16(h) << 8) | uint16(l)
            cpu.PC += 1

        case Inc8L:
            cpu.Cycles += 1
            h := uint8(cpu.HL >> 8)
            l := uint8(cpu.HL & 0xff)

            carry := uint8(0)
            if l & 0b1111 == 0b1111 {
                carry = 1
            }
            cpu.SetFlagH(carry == 1)

            l += 1

            cpu.SetFlagN(false)
            cpu.SetFlagZ(l == 0)
            cpu.HL = (uint16(h) << 8) | uint16(l)
            cpu.PC += 1

        case Inc8HL:
            cpu.Cycles += 3
            value := cpu.LoadMemory8(cpu.HL)
            cpu.SetFlagN(false)

            carry := uint8(0)
            if value & 0b1111 == 0b1111 {
                carry = 1
            }
            cpu.SetFlagH(carry == 1)

            value += 1
            cpu.SetFlagZ(value == 0)

            cpu.StoreMemory(cpu.HL, value)
            cpu.PC += 1

        case Inc8A:
            cpu.Cycles += 1
            a := cpu.A

            carry := uint8(0)
            if a & 0b1111 == 0b1111 {
                carry = 1
            }
            cpu.SetFlagH(carry == 1)

            a += 1
            cpu.SetFlagN(false)
            cpu.A = a
            cpu.SetFlagZ(a == 0)

            cpu.PC += 1

        case Dec8B:
            cpu.Cycles += 1
            b := uint8(cpu.BC >> 8)
            c := uint8(cpu.BC & 0xff)

            h := uint8(0)
            if b & 0b1111 == 0 {
                h = 1
            }

            cpu.SetFlagH(h == 1)
            b -= 1
            cpu.SetFlagN(true)
            cpu.SetFlagZ(b == 0)
            cpu.BC = (uint16(b) << 8) | uint16(c)
            cpu.PC += 1

        case Dec8C:
            cpu.Cycles += 1
            b := uint8(cpu.BC >> 8)
            c := uint8(cpu.BC & 0xff)

            h := uint8(0)
            if c & 0b1111 == 0 {
                h = 1
            }

            cpu.SetFlagH(h == 1)

            c -= 1
            cpu.SetFlagN(true)
            cpu.SetFlagZ(c == 0)
            cpu.BC = (uint16(b) << 8) | uint16(c)
            cpu.PC += 1

        case Dec8D:
            cpu.Cycles += 1
            d := uint8(cpu.DE >> 8)
            e := uint8(cpu.DE & 0xff)

            h := uint8(0)
            if d & 0b1111 == 0 {
                h = 1
            }

            cpu.SetFlagH(h == 1)
            d -= 1
            cpu.SetFlagN(true)
            cpu.SetFlagZ(d == 0)
            cpu.DE = (uint16(d) << 8) | uint16(e)
            cpu.PC += 1

        case Dec8E:
            cpu.Cycles += 1
            d := uint8(cpu.DE >> 8)
            e := uint8(cpu.DE & 0xff)

            h := uint8(0)
            if e & 0b1111 == 0 {
                h = 1
            }

            cpu.SetFlagH(h == 1)
            e -= 1
            cpu.SetFlagN(true)
            cpu.SetFlagZ(e == 0)
            cpu.DE = (uint16(d) << 8) | uint16(e)
            cpu.PC += 1

        case Dec8H:
            cpu.Cycles += 1
            h := uint8(cpu.HL >> 8)
            l := uint8(cpu.HL & 0xff)

            carry := uint8(0)
            if h & 0b1111 == 0 {
                carry = 1
            }

            cpu.SetFlagH(carry == 1)
            h -= 1
            cpu.SetFlagN(true)
            cpu.SetFlagZ(h == 0)
            cpu.HL = (uint16(h) << 8) | uint16(l)
            cpu.PC += 1

        case Dec8L:
            cpu.Cycles += 1
            h := uint8(cpu.HL >> 8)
            l := uint8(cpu.HL & 0xff)

            carry := uint8(0)
            if l & 0b1111 == 0 {
                carry = 1
            }

            cpu.SetFlagH(carry == 1)

            l -= 1
            cpu.SetFlagN(true)

            cpu.SetFlagZ(l == 0)
            cpu.HL = (uint16(h) << 8) | uint16(l)
            cpu.PC += 1

        case Dec8HL:
            cpu.Cycles += 3
            value := cpu.LoadMemory8(cpu.HL)

            carry := uint8(0)
            if value & 0b1111 == 0 {
                carry = 1
            }
            cpu.SetFlagH(carry == 1)

            value -= 1
            cpu.SetFlagN(true)
            cpu.SetFlagZ(value == 0)
            cpu.StoreMemory(cpu.HL, value)

            cpu.PC += 1

        case Dec8A:
            cpu.Cycles += 1
            a := cpu.A

            carry := uint8(0)
            if a & 0b1111 == 0 {
                carry = 1
            }
            cpu.SetFlagH(carry == 1)

            a -= 1
            cpu.SetFlagN(true)
            cpu.SetFlagZ(a == 0)
            cpu.A = a

            cpu.PC += 1

        case DecBC:
            cpu.Cycles += 2
            cpu.BC -= 1
            cpu.PC += 1
        case DecDE:
            cpu.Cycles += 2
            cpu.DE -= 1
            cpu.PC += 1
        case DecHL:
            cpu.Cycles += 2
            cpu.HL -= 1
            cpu.PC += 1
        case DecSP:
            cpu.Cycles += 2
            cpu.SP -= 1
            cpu.PC += 1

        case Load8Immediate:
            cpu.Cycles += 2
            cpu.SetRegister8(instruction.R8_1, instruction.Immediate8)
            cpu.PC += 2

        case StoreHLImmediate:
            cpu.Cycles += 3
            cpu.StoreMemory(cpu.HL, instruction.Immediate8)
            cpu.PC += 2

        case LoadR8R8:
            cpu.Cycles += 1

            var value uint8

            if instruction.R8_2 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_2)
            }

            if instruction.R8_1 == R8HL {
                cpu.StoreMemory(cpu.HL, value)
            } else {
                cpu.SetRegister8(instruction.R8_1, value)
            }

            cpu.PC += 1

        case PopAF:
            cpu.Cycles += 3
            low := cpu.LoadMemory8(cpu.SP)
            cpu.SP += 1
            high := cpu.LoadMemory8(cpu.SP)
            cpu.SP += 1
            cpu.F = low & 0b11110000
            cpu.A = high
            cpu.PC += 1

        case PopR16:
            cpu.Cycles += 3
            low := cpu.LoadMemory8(cpu.SP)
            cpu.SP += 1
            high := cpu.LoadMemory8(cpu.SP)
            cpu.SP += 1

            full := (uint16(high) << 8) | uint16(low)

            switch instruction.R16_1 {
                case 0: cpu.BC = full
                case 1: cpu.DE = full
                case 2: cpu.HL = full
            }

            cpu.PC += 1

        case PushR16:
            cpu.Cycles += 4

            var value uint16

            switch instruction.R16_1 {
                case 0: value = cpu.BC
                case 1: value = cpu.DE
                case 2: value = cpu.HL
            }

            cpu.Push16(value)
            cpu.PC += 1

        case PushAF:
            cpu.Cycles += 4
            cpu.SP -= 1
            cpu.StoreMemory(cpu.SP, cpu.A)
            cpu.SP -= 1
            cpu.StoreMemory(cpu.SP, cpu.F)

        case AddAR8:
            cpu.Cycles += 1
            cpu.PC += 1
            var value uint8

            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            cpu.doAddA(value)

        case AddAImmediate:
            cpu.Cycles += 2
            cpu.PC += 2
            cpu.doAddA(instruction.Immediate8)

        case LdHlSpImmediate8:
            cpu.Cycles += 3
            value := int8(instruction.Immediate8)

            carry := uint8(0)
            if int32(cpu.SP) + int32(value) > 0xffff {
                carry = 1
            }

            halfCarry := uint8(0)
            if ((cpu.SP & 0xf) + (uint16(value) & 0xf)) & 0x10 == 0x10 {
                halfCarry = 1
            }

            cpu.HL = uint16(int32(cpu.SP) + int32(value))
            cpu.SetFlagC(carry == 1)
            cpu.SetFlagH(halfCarry == 1)
            cpu.SetFlagZ(false)
            cpu.SetFlagN(false)

        case AddSpImmediate8:
            cpu.Cycles += 4
            cpu.PC += 2
            value := int8(instruction.Immediate8)

            low := int8(cpu.SP & 0xff)

            carry := uint8(0)
            if (cpu.SP & 0xff) + (uint16(value) & 0xff) > 0xff {
                carry = 1
            }
            halfCarry := uint8(0)
            if ((uint8(low) & 0xf) + (uint8(value) & 0xf)) > 0xf {
                halfCarry = 1
            }

            low += value
            cpu.SP = (uint16(cpu.SP) & 0xff00) | uint16(low)
            cpu.SetFlagC(carry == 1)
            cpu.SetFlagH(halfCarry == 1)
            cpu.SetFlagZ(false)
            cpu.SetFlagN(false)

        case AdcAR8:
            cpu.Cycles += 1
            cpu.PC += 1
            var value uint8

            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            cpu.doAdcA(value)

        case AdcAImmediate:
            cpu.Cycles += 2
            cpu.PC += 2
            cpu.doAdcA(instruction.Immediate8)

        case AndAR8:
            cpu.Cycles += 1
            cpu.PC += 1

            var value uint8
            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            cpu.doAndA(value)

        case AndAImmediate:
            cpu.Cycles += 2
            cpu.PC += 2

            cpu.doAndA(instruction.Immediate8)

        case XorAR8:
            cpu.Cycles += 1
            cpu.PC += 1

            var value uint8
            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            cpu.doXorA(value)

        case XorAImmediate:
            cpu.Cycles += 2
            cpu.PC += 2
            cpu.doXorA(instruction.Immediate8)

        case OrAR8:
            cpu.Cycles += 1
            cpu.PC += 1

            var value uint8
            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            cpu.doOrA(value)

        case OrAImmediate:
            cpu.Cycles += 2
            cpu.PC += 2
            cpu.doOrA(instruction.Immediate8)

        case SubAR8:
            cpu.Cycles += 1
            cpu.PC += 1

            var value uint8
            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            cpu.doSubA(value)

        case SubAImmediate:
            cpu.Cycles += 2
            cpu.PC += 2
            cpu.doSubA(instruction.Immediate8)

        case SbcAR8:
            cpu.Cycles += 1
            cpu.PC += 1

            var value uint8
            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            cpu.doSbcA(value)

        case SbcAImmediate:
            cpu.Cycles += 2
            cpu.PC += 2
            cpu.doSbcA(instruction.Immediate8)

        case CpAR8:
            // same as sub, but don't store result
            cpu.Cycles += 1
            cpu.PC += 1

            var value uint8
            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            cpu.doCpA(value)

        case CpAImmediate:
            // same as sub, but don't store result
            cpu.Cycles += 2
            cpu.PC += 2
            cpu.doCpA(instruction.Immediate8)

        case AddHLBC:
            cpu.Cycles += 2
            cpu.AddHL(cpu.BC)
            cpu.PC += 1
        case AddHLDE:
            cpu.Cycles += 2
            cpu.AddHL(cpu.DE)
            cpu.PC += 1
        case AddHLHL:
            cpu.Cycles += 2
            cpu.AddHL(cpu.HL)
            cpu.PC += 1
        case AddHLSP:
            cpu.Cycles += 2
            cpu.AddHL(cpu.SP)
            cpu.PC += 1

        case CPL:
            cpu.Cycles += 1
            cpu.A = ^cpu.A
            cpu.SetFlagN(true)
            cpu.SetFlagH(true)
            cpu.PC += 1

        case CCF:
            cpu.Cycles += 1
            carry := cpu.GetFlagC()
            cpu.SetFlagC((1 - carry) == 1)
            cpu.SetFlagH(false)
            cpu.SetFlagN(false)

            cpu.PC += 1

        case RLCA:
            cpu.Cycles += 1

            newA, carry := RotateLeft(cpu.A)
            cpu.A = newA
            cpu.SetFlagZ(false)
            cpu.SetFlagH(false)
            cpu.SetFlagN(false)
            cpu.SetFlagC(carry == 1)

            cpu.PC += 1

        case RLC:
            cpu.Cycles += 2
            cpu.PC += 2

            var value uint8
            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            newValue, carry := RotateLeft(value)
            if instruction.R8_1 == R8HL {
                cpu.StoreMemory(cpu.HL, newValue)
            } else {
                cpu.SetRegister8(instruction.R8_1, newValue)
            }
            cpu.SetFlagZ(newValue == 0)
            cpu.SetFlagH(false)
            cpu.SetFlagN(false)
            cpu.SetFlagC(carry == 1)

        case RRCA:
            cpu.Cycles += 1

            newA, carry := RotateRight(cpu.A)
            cpu.A = newA
            cpu.SetFlagZ(false)
            cpu.SetFlagH(false)
            cpu.SetFlagN(false)
            cpu.SetFlagC(carry == 1)
            cpu.PC += 1

        case RRC:
            cpu.Cycles += 2
            cpu.PC += 2

            var value uint8
            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            newValue, carry := RotateRight(value)
            if instruction.R8_1 == R8HL {
                cpu.StoreMemory(cpu.HL, newValue)
            } else {
                cpu.SetRegister8(instruction.R8_1, newValue)
            }
            cpu.SetFlagZ(newValue == 0)
            cpu.SetFlagH(false)
            cpu.SetFlagN(false)
            cpu.SetFlagC(carry == 1)

        case RLA:
            cpu.Cycles += 1

            oldCarry := cpu.GetFlagC()
            newCarry := (cpu.A >> 7) & 0b1
            cpu.A = (cpu.A << 1) | oldCarry

            cpu.SetFlagZ(false)
            cpu.SetFlagH(false)
            cpu.SetFlagN(false)
            cpu.SetFlagC(newCarry == 1)

            cpu.PC += 1

        case RL:
            cpu.Cycles += 2
            cpu.PC += 2

            var value uint8
            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            oldCarry := cpu.GetFlagC()
            newCarry := (value >> 7) & 0b1
            newValue := (value << 1) | oldCarry

            if instruction.R8_1 == R8HL {
                cpu.StoreMemory(cpu.HL, newValue)
            } else {
                cpu.SetRegister8(instruction.R8_1, newValue)
            }
            cpu.SetFlagZ(newValue == 0)
            cpu.SetFlagH(false)
            cpu.SetFlagN(false)
            cpu.SetFlagC(newCarry == 1)

        case RRA:
            cpu.Cycles += 1

            oldCarry := cpu.GetFlagC()
            newCarry := cpu.A & 0b1
            cpu.A = (cpu.A >> 1) | (oldCarry << 7)

            cpu.SetFlagZ(false)
            cpu.SetFlagH(false)
            cpu.SetFlagN(false)
            cpu.SetFlagC(newCarry == 1)
            cpu.PC += 1

        case RR:
            cpu.Cycles += 2
            cpu.PC += 2

            var value uint8
            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            oldCarry := cpu.GetFlagC()
            newCarry := value & 0b1
            newValue := (value >> 1) | (oldCarry << 7)

            if instruction.R8_1 == R8HL {
                cpu.StoreMemory(cpu.HL, newValue)
            } else {
                cpu.SetRegister8(instruction.R8_1, newValue)
            }
            cpu.SetFlagZ(newValue == 0)
            cpu.SetFlagH(false)
            cpu.SetFlagN(false)
            cpu.SetFlagC(newCarry == 1)

        case SLA:
            cpu.Cycles += 2
            cpu.PC += 2

            var value uint8
            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            carry := value >> 7
            newValue := value << 1
            if instruction.R8_1 == R8HL {
                cpu.StoreMemory(cpu.HL, newValue)
            } else {
                cpu.SetRegister8(instruction.R8_1, newValue)
            }
            cpu.SetFlagZ(newValue == 0)
            cpu.SetFlagH(false)
            cpu.SetFlagN(false)
            cpu.SetFlagC(carry == 1)

        case SRA:
            cpu.Cycles += 2
            cpu.PC += 2

            var value uint8
            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            carry := value & 0b1
            newValue := (value >> 1) | (value & 0x80)
            if instruction.R8_1 == R8HL {
                cpu.StoreMemory(cpu.HL, newValue)
            } else {
                cpu.SetRegister8(instruction.R8_1, newValue)
            }
            cpu.SetFlagZ(newValue == 0)
            cpu.SetFlagH(false)
            cpu.SetFlagN(false)
            cpu.SetFlagC(carry == 1)

        case SRL:
            cpu.Cycles += 2
            cpu.PC += 2

            var value uint8
            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            carry := value & 0b1
            newValue := value >> 1
            if instruction.R8_1 == R8HL {
                cpu.StoreMemory(cpu.HL, newValue)
            } else {
                cpu.SetRegister8(instruction.R8_1, newValue)
            }
            cpu.SetFlagZ(newValue == 0)
            cpu.SetFlagH(false)
            cpu.SetFlagN(false)
            cpu.SetFlagC(carry == 1)

        case SWAP:
            cpu.Cycles += 2
            cpu.PC += 2

            var value uint8
            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            upper := value >> 4
            lower := value & 0b1111

            newValue := (lower << 4) | upper

            if instruction.R8_1 == R8HL {
                cpu.StoreMemory(cpu.HL, newValue)
            } else {
                cpu.SetRegister8(instruction.R8_1, newValue)
            }

            cpu.SetFlagN(false)
            cpu.SetFlagH(false)
            cpu.SetFlagC(false)
            cpu.SetFlagZ(newValue == 0)

        case Bit:
            cpu.Cycles += 2
            cpu.PC += 2

            var value uint8
            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            cpu.SetFlagZ(value & (1 << instruction.Immediate8) == 0)
            cpu.SetFlagN(false)
            cpu.SetFlagH(true)

        case Res:
            cpu.Cycles += 2
            cpu.PC += 2

            var value uint8
            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            newValue := value & ^(1 << instruction.Immediate8)

            if instruction.R8_1 == R8HL {
                cpu.StoreMemory(cpu.HL, newValue)
            } else {
                cpu.SetRegister8(instruction.R8_1, newValue)
            }

        case Set:
            cpu.Cycles += 2
            cpu.PC += 2

            var value uint8
            if instruction.R8_1 == R8HL {
                value = cpu.LoadMemory8(cpu.HL)
            } else {
                value = cpu.GetRegister8(instruction.R8_1)
            }

            newValue := value | (1 << instruction.Immediate8)

            if instruction.R8_1 == R8HL {
                cpu.StoreMemory(cpu.HL, newValue)
            } else {
                cpu.SetRegister8(instruction.R8_1, newValue)
            }

        case Stop:
            cpu.Stopped = true
            cpu.PC += 1

        case Halt:
            cpu.Halted = true

        case DAA:
            // BCD fixup after add/subtract

            // https://blog.ollien.com/posts/gb-daa/
            // https://ehaskins.com/2018-01-30%20Z80%20DAA/

            var offset uint8 = 0
            c := uint8(0)

            is_subtract := cpu.GetFlagN() == 1

            if (!is_subtract && cpu.A & 0xf > 9) || cpu.GetFlagH() == 1 {
                offset += 6
            }

            if (!is_subtract && cpu.A > 0x99) || cpu.GetFlagC() == 1 {
                offset += 0x60
                c = 1
            }

            if !is_subtract {
                cpu.A += offset
            } else {
                cpu.A -= offset
            }

            cpu.SetFlagZ(cpu.A == 0)
            cpu.SetFlagH(false)
            cpu.SetFlagC(c == 1)

            cpu.Cycles += 1
            cpu.PC += 1

        case SCF:
            cpu.Cycles += 1
            cpu.SetFlagN(false)
            cpu.SetFlagH(false)
            cpu.SetFlagC(true)
            cpu.PC += 1

        default:
            log.Printf("Execute error: unknown opcode %v", instruction.Opcode)
    }
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
        case R16HL: return Instruction{Opcode: StoreHLIncMemA}
        case R16SP: return Instruction{Opcode: StoreHLDecMemA}
    }

    return Instruction{Opcode: Unknown}
}

func makeLoadAFromR16MemInstruction(r16 R16) Instruction {
    switch r16 {
        case R16BC: return Instruction{Opcode: LoadAMemBC}
        case R16DE: return Instruction{Opcode: LoadAMemDE}
        case R16HL: return Instruction{Opcode: LoadAMemHLI}
        case R16SP: return Instruction{Opcode: LoadAMemHLD}
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
func (cpu *CPU) DecodeInstruction() (Instruction, uint8) {
    instruction := cpu.LoadMemory8(cpu.PC)

    // special case for CB prefix
    if instruction == 0xcb {
        instruction = cpu.LoadMemory8(cpu.PC + 1)

        switch instruction >> 6 {
            case 0b00:

                r8 := R8(instruction & 0b111)

                switch instruction >> 3 {
                    // rlc
                    case 0b00000: return Instruction{Opcode: RLC, R8_1: r8}, 2
                    // rrc
                    case 0b00001: return Instruction{Opcode: RRC, R8_1: r8}, 2
                    // rl
                    case 0b00010: return Instruction{Opcode: RL, R8_1: r8}, 2
                    // rr
                    case 0b00011: return Instruction{Opcode: RR, R8_1: r8}, 2
                    // sla
                    case 0b00100: return Instruction{Opcode: SLA, R8_1: r8}, 2
                    // sra
                    case 0b00101: return Instruction{Opcode: SRA, R8_1: r8}, 2
                    // swap
                    case 0b00110: return Instruction{Opcode: SWAP, R8_1: r8}, 2
                    // srl
                    case 0b00111: return Instruction{Opcode: SRL, R8_1: r8}, 2
                }

            case 0b01:
                // bit
                r8 := R8(instruction & 0b111)
                bit := (instruction >> 3) & 0b111
                return Instruction{Opcode: Bit, R8_1: r8, Immediate8: bit}, 2
            case 0b10:
                // res
                r8 := R8(instruction & 0b111)
                bit := (instruction >> 3) & 0b111
                return Instruction{Opcode: Res, R8_1: r8, Immediate8: bit}, 2
            case 0b11:
                // set
                r8 := R8(instruction & 0b111)
                bit := (instruction >> 3) & 0b111
                return Instruction{Opcode: Set, R8_1: r8, Immediate8: bit}, 2
        }

        return Instruction{Opcode: Unknown}, 2
    }

    block := instruction >> 6
    // check top 2 bits first
    switch block {
        case 0b00:
            switch instruction & 0b1111 {
                case 0b0000:
                    if instruction == 0b00000000 {
                        return Instruction{Opcode: Nop}, 1
                    }
                case 0b0001:
                    r16 := (instruction >> 4) & 0b11

                    /*
                    if len(instructions[1:]) < 2 {
                        return Instruction{Opcode: Unknown}, 1
                    }
                    */

                    return makeLoadR16Imm16Instruction(R16(r16), cpu.LoadMemory16(cpu.PC+1)), 3
                case 0b0010:
                    //return "ld [r16mem], a"
                    r16 := R16((instruction >> 4) & 0b11)
                    return makeStoreR16MemAInstruction(r16), 1
                case 0b1010:
                    r16 := R16((instruction >> 4) & 0b11)
                    return makeLoadAFromR16MemInstruction(r16), 1
                    //return "ld a, [r16mem]"

                case 0b1000:
                    // ambiguous with jr n
                    if instruction == 0b00001000 {
                        // immediate := makeImm16(instructions[1:])
                        immediate := cpu.LoadMemory16(cpu.PC+1)
                        return Instruction{Opcode: StoreSPMem16, Immediate16: immediate}, 3
                    }

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
                        return Instruction{Opcode: JR, Immediate8: cpu.LoadMemory8(cpu.PC+1)}, 2
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

                        return Instruction{Opcode: opcode, Immediate8: cpu.LoadMemory8(cpu.PC+1)}, 2
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
                        return Instruction{Opcode: StoreHLImmediate, Immediate8: cpu.LoadMemory8(cpu.PC+1)}, 2
                    }

                    return Instruction{Opcode: Load8Immediate, R8_1: r8, Immediate8: cpu.LoadMemory8(cpu.PC+1)}, 2
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
                    return Instruction{Opcode: AddAR8, R8_1: r8}, 1
                case 0b001:
                    r8 := R8(instruction & 0b111)
                    return Instruction{Opcode: AdcAR8, R8_1: r8}, 1

                case 0b010:
                    r8 := R8(instruction & 0b111)
                    return Instruction{Opcode: SubAR8, R8_1: r8}, 1

                case 0b011:
                    r8 := R8(instruction & 0b111)
                    return Instruction{Opcode: SbcAR8, R8_1: r8}, 1

                case 0b100:
                    r8 := R8(instruction & 0b111)
                    return Instruction{Opcode: AndAR8, R8_1: r8}, 1

                case 0b101:
                    r8 := R8(instruction & 0b111)
                    return Instruction{Opcode: XorAR8, R8_1: r8}, 1

                case 0b110:
                    r8 := R8(instruction & 0b111)
                    return Instruction{Opcode: OrAR8, R8_1: r8}, 1

                case 0b111:
                    r8 := R8(instruction & 0b111)
                    return Instruction{Opcode: CpAR8, R8_1: r8}, 1
            }
        case 0b11:
            switch instruction & 0b111111 {
                case 0b000110:
                    return Instruction{Opcode: AddAImmediate, Immediate8: cpu.LoadMemory8(cpu.PC+1)}, 2
                case 0b001110:
                    return Instruction{Opcode: AdcAImmediate, Immediate8: cpu.LoadMemory8(cpu.PC+1)}, 2
                case 0b010110:
                    return Instruction{Opcode: SubAImmediate, Immediate8: cpu.LoadMemory8(cpu.PC+1)}, 2
                case 0b011110:
                    return Instruction{Opcode: SbcAImmediate, Immediate8: cpu.LoadMemory8(cpu.PC+1)}, 2
                case 0b100110:
                    return Instruction{Opcode: AndAImmediate, Immediate8: cpu.LoadMemory8(cpu.PC+1)}, 2
                case 0b101110:
                    return Instruction{Opcode: XorAImmediate, Immediate8: cpu.LoadMemory8(cpu.PC+1)}, 2
                case 0b110110:
                    return Instruction{Opcode: OrAImmediate, Immediate8: cpu.LoadMemory8(cpu.PC+1)}, 2
                case 0b111110:
                    return Instruction{Opcode: CpAImmediate, Immediate8: cpu.LoadMemory8(cpu.PC+1)}, 2
                case 0b100010:
                    return Instruction{Opcode: LdhCA}, 1
                    // return "ldh [c], a"

                case 0b100000:
                    return Instruction{Opcode: LdhImmediate8A, Immediate8: cpu.LoadMemory8(cpu.PC+1)}, 2
                    // return "ldh [imm8], a"

                case 0b101010:
                    return Instruction{Opcode: LdImmediate16A, Immediate16: cpu.LoadMemory16(cpu.PC+1)}, 3
                    // return "ld [imm16], a"

                case 0b110010:
                    return Instruction{Opcode: LdhAC}, 1
                    // return "ldh a, [c]"

                case 0b110000:
                    return Instruction{Opcode: LdhAImmediate8, Immediate8: cpu.LoadMemory8(cpu.PC+1)}, 2
                    // return "ldh a, [imm8]"

                case 0b111010:
                    return Instruction{Opcode: LdAImmediate16, Immediate16: cpu.LoadMemory16(cpu.PC+1)}, 3
                    // return "ld a, [imm16]"

                case 0b101000:
                    return Instruction{Opcode: AddSpImmediate8, Immediate8: cpu.LoadMemory8(cpu.PC+1)}, 2
                    // return "add sp, imm8"

                case 0b111000:
                    return Instruction{Opcode: LdHlSpImmediate8, Immediate8: cpu.LoadMemory8(cpu.PC+1)}, 2
                    // return "ld hl, sp + imm8"

                case 0b111001:
                    return Instruction{Opcode: LdSpHl}, 1
                    // return "ld sp, hl"

                case 0b110011:
                    return Instruction{Opcode: DisableInterrupts}, 1
                    // return "di"

                case 0b111011:
                    return Instruction{Opcode: EnableInterrupts}, 1
                    // return "ei"
            }

            switch instruction & 0b1111 {
                case 0b0001:
                    r16 := R16((instruction >> 4) & 0b11)
                    if r16 == 3 {
                        return Instruction{Opcode: PopAF}, 1
                    }

                    return Instruction{Opcode: PopR16, R16_1: r16}, 1
                    // return "pop r16stk"

                case 0b0101:
                    r16 := R16((instruction >> 4) & 0b11)
                    if r16 == 3 {
                        return Instruction{Opcode: PushAF}, 1
                    }

                    return Instruction{Opcode: PushR16, R16_1: r16}, 1
                    // return "push r16stk"
            }

            if instruction & 0b111 == 0b111 {
                // return "rst tgt3"
                address := uint8((instruction >> 3) & 0b111)
                return Instruction{Opcode: CallResetVector, Immediate8: address*8}, 1
            }

            if instruction == 0b11101001 {
                return Instruction{Opcode: JpHL}, 1
                // return "jp hl"
            }

            if instruction >> 5 == 0b110 {
                switch instruction & 0b111 {
                    case 0b000:
                        cond := (instruction >> 3) & 0b11
                        opcode := RetNz

                        switch cond {
                            case 0b00: opcode = RetNz
                            case 0b01: opcode = RetZ
                            case 0b10: opcode = RetNc
                            case 0b11: opcode = RetC
                        }

                        return Instruction{Opcode: opcode}, 1
                        // return "ret cond"

                    case 0b010:
                        cond := (instruction >> 3) & 0b11
                        imm16 := cpu.LoadMemory16(cpu.PC+1)
                        opcode := JpNzImmediate16

                        switch cond {
                            case 0b00: opcode = JpNzImmediate16
                            case 0b01: opcode = JpZImmediate16
                            case 0b10: opcode = JpNcImmediate16
                            case 0b11: opcode = JpCImmediate16
                        }

                        return Instruction{Opcode: opcode, Immediate16: imm16}, 3
                        // return "jp cond, imm16"

                    case 0b011:
                        imm16 := cpu.LoadMemory16(cpu.PC+1)
                        return Instruction{Opcode: JpImmediate16, Immediate16: imm16}, 3
                        // return "jp imm16"

                    case 0b100:
                        cond := (instruction >> 3) & 0b11
                        imm16 := cpu.LoadMemory16(cpu.PC+1)
                        opcode := CallNzImmediate16

                        switch cond {
                            case 0b00: opcode = CallNzImmediate16
                            case 0b01: opcode = CallZImmediate16
                            case 0b10: opcode = CallNcImmediate16
                            case 0b11: opcode = CallCImmediate16
                        }

                        return Instruction{Opcode: opcode, Immediate16: imm16}, 3

                        // return "call cond, imm16"

                }

                switch instruction {
                    case 0b11001001:
                        return Instruction{Opcode: Return}, 1
                        // return "ret"
                    case 0b11011001:
                        return Instruction{Opcode: ReturnFromInterrupt}, 1
                        // return "reti"

                    case 0b11001101:
                        imm16 := cpu.LoadMemory16(cpu.PC+1)
                        return Instruction{Opcode: CallImmediate16, Immediate16: imm16}, 3
                        // return "call imm16"
                }
            }
    }

    return Instruction{Opcode: Unknown}, 1
}
