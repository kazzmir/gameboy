package main

import (
    "os"
    "fmt"
    "log"
    "time"
    "flag"

    "github.com/kazzmir/gameboy/core"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    Cpu *core.CPU
    cpuBudget int64
    ticker *time.Ticker
    rate int64
    pixels []uint8
    needDraw bool
    maxCycle int64
}

func MakeEngine(cpu *core.CPU, maxCycle int64) *Engine {
    rate := int64(60)
    ticker := time.NewTicker(time.Second / time.Duration(rate))

    return &Engine{
        Cpu: cpu,
        cpuBudget: 0,
        ticker: ticker,
        rate: rate,
        maxCycle: maxCycle,
    }
}

func (engine *Engine) runEmulator() error {

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

    /*
    readBudget := true
    for readBudget {
        select {
            case <-engine.ticker.C:
                engine.cpuBudget += int64(core.CPUSpeed / engine.rate)
            default:
                readBudget = false
        }
    }
    */

    engine.cpuBudget += core.CPUSpeed / engine.rate

    // log.Printf("cpu budget: %v = %v/s. cpu speed = %v. diff = %v", engine.cpuBudget, engine.cpuBudget * 60, core.CPUSpeed, engine.cpuBudget * 60 - core.CPUSpeed)

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

        if engine.maxCycle > 0 {
            engine.maxCycle -= int64(cpuCyclesTaken)
            if engine.maxCycle <= 0 {
                return fmt.Errorf("Max cycles reached")
            }
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

    return nil
}

func (engine *Engine) Update() error {
    keys := inpututil.AppendJustPressedKeys(nil)
    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    err := engine.runEmulator()

    return err
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
            r, g, b, a := virtualScreen[y][x].RGBA()
            index := (y*160 + x) * 4
            engine.pixels[index+0] = uint8(r >> 8)
            engine.pixels[index+1] = uint8(g >> 8)
            engine.pixels[index+2] = uint8(b >> 8)
            engine.pixels[index+3] = uint8(a >> 8)
        }
    }

    screen.WritePixels(engine.pixels)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (int, int) {
    return 160, 144
}

func main(){
    maxCycle := flag.Int64("max", 0, "Max cycles to run")
    flag.Parse()

    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    var path string
    for i := flag.NArg(); i < len(os.Args); i++ {
        path = os.Args[i]
    }

    if path == "" {
        log.Printf("Usage: gameboy /path/to/rom")
        return
    }

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
    cpu.Error = true
    cpu.PPU.Debug = false
    // cpu.PC = 0x100

    ebiten.SetWindowSize(160*4, 144*4)
    ebiten.SetWindowTitle("Gameboy Emulator")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    engine := MakeEngine(cpu, *maxCycle)

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
