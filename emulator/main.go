package main

import (
    "os"
    "log"

    "github.com/kazzmir/gameboy/core"
)

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    if len(os.Args) != 2 {
        log.Printf("Usage: gameboy /path/to/rom")
        return
    }

    path := os.Args[1]

    gameboyFile, err := core.LoadGameboyFromFile(path)
    if err != nil {
        log.Printf("Error: %v", err)
    }

    log.Printf("Loaded %d bytes", len(gameboyFile.Data))

    log.Printf("Gameboy file '%v'", gameboyFile.GetTitle())
    log.Printf("Rom size: %v", gameboyFile.GetRomSize())

    cpu := core.MakeCPU(gameboyFile.GetRom())
    cpu.InitializeDMG()
    // cpu.PC = 0x100

    for range 55420 {
        log.Printf("PC: 0x%x", cpu.PC)
        next, _ := cpu.DecodeInstruction()
        log.Printf("Execute instruction: %+v", next)
        cpuCycles := cpu.Execute(next)
        cpu.PPU.Run(cpuCycles)
    }
}
