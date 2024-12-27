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
}

type Opcode int
const (
    Nop Opcode = iota

    LoadBCImmediate
    LoadDEImmediate
    LoadHLImmediate
    LoadSPImmediate

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

    DecBC
    DecDE
    DecHL
    DecSP

    AddHLBC
    AddHLDE
    AddHLHL
    AddHLSP

    RLCA
    RLA
    RRCA
    RRA

    DAA

    SCF

    Unknown
)

type R16 uint8
const (
    R16BC R16 = 0
    R16DE R16 = 1
    R16HL R16 = 2
    R16SP R16 = 3
)

type Instruction struct {
    Opcode Opcode
    Immediate8 uint8
    Immediate16 uint16
}

func setBit(value uint8, bit uint8, set uint8) uint8 {
    set &= 0b1
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
                        /*
                        case 0b0010: return "cpl"
                        case 0b0011: return "ccf"
                        */
                    }
            }

            /*
            switch instruction & 0b111 {
                case 0b000:
                    if instruction == 0b00011000 {
                        return "jr imm8"
                    }

                    if instruction == 0b00010000 {
                        return "stop"
                    }

                    if instruction >> 5 == 0b001 {
                        return "jr cond, imm8"
                    }

                case 0b100: return "inc r8"
                case 0b101: return "dec r8"
                case 0b110: return "ld r8, imm8"
            }
            */

            /*
        case 0b01:
            if instruction & 0b11111 == 0b110110 {
                return "halt"
            }

            // source := instruction & 0b111
            // dest := (instruction >> 3) & 0b111

            return "ld r8, r8"

        case 0b10:
            switch (instruction >> 3) & 0b111 {
                case 0b000: return "add a, r8"
                case 0b001: return "adc a, r8"
                case 0b010: return "sub a, r8"
                case 0b011: return "sbc a, r8"
                case 0b100: return "and a, r8"
                case 0b101: return "xor a, r8"
                case 0b110: return "or a, r8"
                case 0b111: return "cp a, r8"
            }
        case 0b11:
            switch instruction & 0b111111 {
                case 0b000110: return "add a, imm8"
                case 0b001110: return "adc a, imm8"
                case 0b010110: return "sub a, imm8"
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
            }

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
