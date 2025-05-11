package main

import (
    "os"
    "log"
    "time"

    "github.com/kazzmir/gameboy/core"
)

func runEmulator(cpu *core.CPU) {

    /*
    for range 55420 {
        if cpu.Debug {
            log.Printf("PC: 0x%x", cpu.PC)
        }
        next, _ := cpu.DecodeInstruction()
        if cpu.Debug {
            log.Printf("Execute instruction: %+v", next)
        }
        cpuCyclesTaken := cpu.Execute(next)
        cpu.PPU.Run(cpuCyclesTaken * 4)
    }
    */

    rate := 256

    ticker := time.NewTicker(time.Second / time.Duration(rate))

    var cpuBudget int64

    fpsTicker := time.NewTicker(time.Second)

    frames := 0
    for {
        if cpuBudget > 0 {
            next, _ := cpu.DecodeInstruction()
            cpuCyclesTaken := cpu.Execute(next)
            cpu.PPU.Run(cpuCyclesTaken * 4)

            cpuBudget -= int64(cpuCyclesTaken)

            select {
                case <-cpu.PPU.Draw:
                    // log.Printf("Draw screen")
                    frames += 1
                default:
            }

        } else {
            // log.Printf("Done with CPU cycles, waiting for next tick")
            select {
                case <-ticker.C:
                    cpuBudget += int64(core.CPUSpeed / rate)
            }
            // log.Printf("Execute %v cycles", cpuBudget)
        }

        select {
            case <-fpsTicker.C:
                log.Printf("FPS: %v", frames)
                frames = 0
            default:
        }

    }
}

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
    cpu.Debug = false
    cpu.Error = false
    cpu.PPU.Debug = false
    // cpu.PC = 0x100

    runEmulator(cpu)

}
