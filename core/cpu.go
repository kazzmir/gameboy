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
    RRCA

    LoadBCImmediate
    LoadDEImmediate
    LoadHLImmediate
    LoadSPImmediate

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

func carryFlag(value uint8) uint8 {
    return value << 4
}

func RotateRight(value uint8) (uint8, uint8) {
    carry := value & 0b1
    value = value >> 1
    value = value | (carry << 7)
    return value, carry
}

func (cpu *CPU) Execute(instruction Instruction) {
    switch instruction.Opcode {
        case Nop:
            cpu.Cycles += 1
        case RRCA:
            cpu.Cycles += 1
            newA, carry := RotateRight(cpu.A)
            cpu.A = newA
            cpu.F = carryFlag(carry)
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
        default:
            log.Printf("Execute error: unknown opcode %v", instruction.Opcode)
    }
}

func makeImm16(data []byte) uint16 {
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
                /*
                case 0b0010: return "ld [r16mem], a"
                case 0b1010: return "ld a, [r16mem]"
                case 0b1000: return "ld [imm16], sp"
                case 0b0011: return "inc r16"
                case 0b1011: return "dec r16"
                case 0b1001: return "add hl, r16"

                case 0b0111:
                    switch instruction >> 4 {
                        case 0b0000: return "rlca"
                        case 0b0001: return "rla"
                        case 0b0010: return "daa"
                        case 0b0011: return "scf"
                    }
                case 0b1111: 
                    switch instruction >> 4 {
                        case 0b0000: return "rrca"
                        case 0b0001: return "rra"
                        case 0b0010: return "cpl"
                        case 0b0011: return "ccf"
                    }
                    */
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
