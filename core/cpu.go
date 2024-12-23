package core

type CPU struct {
    // accumulator and flags
    AF uint16
    BC uint16
    DE uint16
    HL uint16
    // stack pointer
    SP uint16
    // program counter
    PC uint16
}

func DecodeInstruction(instruction byte) string {
    block := instruction >> 6
    switch block {
        case 0:
            switch instruction & 0b1111 {
                case 0b0000: return "NOP"
                case 0b0001: return "ld r16, imm16"
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
            }

            switch instruction & 0b111 {
                case 0b100: return "inc r8"
                case 0b101: return "dec r8"
                case 0b110: return "ld r8, imm8"
            }

        case 1:
        case 2:
        case 3:
    }

    return "unknown"
}
