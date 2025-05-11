package main

import (
    "os"
    "log"
    "time"

    "github.com/kazzmir/gameboy/core"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    Cpu *core.CPU
    cpuBudget int64
    ticker *time.Ticker
    rate int
    pixels []uint8
    needDraw bool
}

func MakeEngine(cpu *core.CPU) *Engine {
    rate := 60
    ticker := time.NewTicker(time.Second / time.Duration(rate))

    return &Engine{
        Cpu: cpu,
        cpuBudget: 0,
        ticker: ticker,
        rate: rate,
    }
}

func (engine *Engine) runEmulator() {

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

    // fpsTicker := time.NewTicker(time.Second)

    readBudget := true
    for readBudget {
        select {
            case <-engine.ticker.C:
                engine.cpuBudget += int64(core.CPUSpeed / engine.rate)
            default:
                readBudget = false
        }
    }

    for engine.cpuBudget > 0 {
        next, _ := engine.Cpu.DecodeInstruction()
        cpuCyclesTaken := engine.Cpu.Execute(next)
        engine.Cpu.PPU.Run(cpuCyclesTaken * 4)

        engine.cpuBudget -= int64(cpuCyclesTaken)

        select {
            case <-engine.Cpu.PPU.Draw:
                engine.needDraw = true
                // log.Printf("Draw screen")
                // frames += 1
            default:
        }

    }

    /*
    frames := 0
    for {
        if cpuBudget > 0 {
            next, _ := cpu.DecodeInstruction()
            cpuCyclesTaken := cpu.Execute(next)
            engine.Cpu.PPU.Run(cpuCyclesTaken * 4)

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
    */
}

func (engine *Engine) Update() error {
    keys := inpututil.AppendJustPressedKeys(nil)
    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    engine.runEmulator()

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    if !engine.needDraw {
        return
    }

    if len(engine.pixels) != screen.Bounds().Dx() * screen.Bounds().Dy() * 4 {
        engine.pixels = make([]uint8, screen.Bounds().Dx() * screen.Bounds().Dy() * 4)
    }

    virtualScreen := engine.Cpu.PPU.Screen

    for y := range len(virtualScreen) {
        for x := range len(virtualScreen[y]) {
            r := (virtualScreen[y][x] >> 24) & 0xff
            g := (virtualScreen[y][x] >> 16) & 0xff
            b := (virtualScreen[y][x] >> 8) & 0xff
            a := (virtualScreen[y][x] >> 0) & 0xff
            engine.pixels[y*160+x+0] = uint8(r)
            engine.pixels[y*160+x+1] = uint8(g)
            engine.pixels[y*160+x+2] = uint8(b)
            engine.pixels[y*160+x+3] = uint8(a)
        }
    }

    screen.WritePixels(engine.pixels)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (int, int) {
    return 160, 144
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

    ebiten.SetWindowSize(160*4, 144*4)
    ebiten.SetWindowTitle("Gameboy Emulator")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    engine := MakeEngine(cpu)

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
