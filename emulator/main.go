package main

import (
    "os"
    "fmt"
    "log"
    "time"
    "flag"
    "errors"
    "image/color"

    "github.com/kazzmir/gameboy/core"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "github.com/hajimehoshi/ebiten/v2/audio"
)

var RestartError = fmt.Errorf("Restart")

const SampleRate = 44100

type Engine struct {
    MakeCpu func () (*core.CPU, error)
    Cpu *core.CPU
    cpuBudget int64
    ticker *time.Ticker
    rate int64
    pixels []uint8
    needDraw bool
    maxCycle int64
    speed float64
    paused bool

    audioContext *audio.Context
    audioPlayer *audio.Player
}

func MakeEngine(makeCpu func () (*core.CPU, error), maxCycle int64, rate int64, speed float64, audioContext *audio.Context) (*Engine, error) {
    ticker := time.NewTicker(time.Second / time.Duration(rate))

    cpu, err := makeCpu()
    if err != nil {
        return nil, err
    }

    return &Engine{
        MakeCpu: makeCpu,
        Cpu: cpu,
        cpuBudget: 0,
        ticker: ticker,
        rate: rate,
        maxCycle: maxCycle,
        speed: speed,
        audioContext: audioContext,
    }, nil
}

// run the emulator for some number of cpu cycles
func (engine *Engine) runEmulator(cycles int64) error {
    // engine.cpuBudget += core.CPUSpeed / engine.rate

    // log.Printf("cpu budget: %v = %v/s. cpu speed = %v. diff = %v", engine.cpuBudget, engine.cpuBudget * engine.rate, core.CPUSpeed, engine.cpuBudget * engine.rate - core.CPUSpeed)

    engine.Cpu.Joypad.Reset()

    pressedKeys := inpututil.AppendJustPressedKeys(nil)
    for _, key := range pressedKeys {
        switch key {
            case ebiten.KeyA, ebiten.KeyS, ebiten.KeyEnter, ebiten.KeySpace:
                if engine.Cpu.Joypad.ReadButtons {
                    engine.Cpu.EnableJoypad()
                }
            case ebiten.KeyUp, ebiten.KeyDown, ebiten.KeyLeft, ebiten.KeyRight:
                if engine.Cpu.Joypad.ReadDpad {
                    engine.Cpu.EnableJoypad()
                }

            case ebiten.KeyR:
                return RestartError
            case ebiten.KeyP:
                engine.paused = !engine.paused
        }
    }

    var speedBoost float64 = 0

    pressedKeys = inpututil.AppendPressedKeys(nil)
    for _, key := range pressedKeys {
        switch key {
            case ebiten.KeyA:
                engine.Cpu.Joypad.A = true
            case ebiten.KeyS:
                engine.Cpu.Joypad.B = true
            case ebiten.KeyEnter:
                engine.Cpu.Joypad.Start = true
            case ebiten.KeySpace:
                engine.Cpu.Joypad.Select = true
            case ebiten.KeyUp:
                engine.Cpu.Joypad.Up = true
            case ebiten.KeyDown:
                engine.Cpu.Joypad.Down = true
            case ebiten.KeyLeft:
                engine.Cpu.Joypad.Left = true
            case ebiten.KeyRight:
                engine.Cpu.Joypad.Right = true
            case ebiten.KeyBackquote:
                speedBoost = 1.5
        }
    }

    if !engine.paused {
        // divide by 4 because the cpu clock is 1/4th of the master clock
        engine.cpuBudget += int64(float64(cycles) * (engine.speed + speedBoost)) / 4
    }

    for engine.cpuBudget > 0 {
        cpuCyclesTaken := engine.Cpu.HandleInterrupts()

        next, _ := engine.Cpu.DecodeInstruction()
        cpuCyclesTaken += engine.Cpu.Execute(next)
        engine.Cpu.PPU.Run(cpuCyclesTaken * 4, engine.Cpu)
        engine.Cpu.APU.Run(cpuCyclesTaken * 4)

        engine.cpuBudget -= int64(cpuCyclesTaken)

        select {
            case <-engine.Cpu.PPU.Draw:
                engine.needDraw = true
                engine.Cpu.EnableVBlank()
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

        engine.Cpu.RunTimer(cpuCyclesTaken)
    }

    return nil
}

func (engine *Engine) Update() error {
    keys := inpututil.AppendJustPressedKeys(nil)
    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    err := engine.runEmulator(core.CPUSpeed / engine.rate)

    if errors.Is(err, RestartError) {
        cpu, err := engine.MakeCpu()
        if err != nil {
            return err
        }
        if engine.audioPlayer != nil {
            engine.audioPlayer.Close()
        }
        engine.audioPlayer = nil
        engine.Cpu = cpu

        return nil
    }

    if engine.audioPlayer == nil {
        player, err := engine.audioContext.NewPlayerF32(engine.Cpu.APU.GetAudioStream())
        if err != nil {
            log.Printf("Error creating audio player: %v", err)
        } else {
            engine.audioPlayer = player
        }
    }

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
            index := (y*core.ScreenWidth + x) * 4
            engine.pixels[index+0] = uint8(r >> 8)
            engine.pixels[index+1] = uint8(g >> 8)
            engine.pixels[index+2] = uint8(b >> 8)
            engine.pixels[index+3] = uint8(a >> 8)
        }
    }

    screen.WritePixels(engine.pixels)

    if engine.paused {
        vector.DrawFilledRect(screen, 0, 0, float32(screen.Bounds().Dx()), float32(screen.Bounds().Dy()), color.RGBA{R: 0, G: 0, B: 0, A: 128}, true)
        ebitenutil.DebugPrintAt(screen, "Paused\nPress P to resume", screen.Bounds().Dx()/2-45, screen.Bounds().Dy()/2-20)
    }
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (int, int) {
    return core.ScreenWidth, core.ScreenHeight
}

func main(){
    maxCycle := flag.Int64("max", 0, "Max cycles to run")
    cpuDebug := flag.Bool("cpu-debug", false, "Enable CPU debug")
    ppuDebug := flag.Bool("ppu-debug", false, "Enable PPU debug")
    fps := flag.Int("fps", 60, "FPS")
    speed := flag.Float64("speed", 1.0, "Speed multiplier")
    flag.Parse()

    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    var path string
    for i := len(os.Args) - flag.NArg(); i < len(os.Args); i++ {
        path = os.Args[i]
    }

    if path == "" {
        log.Printf("Usage: gameboy /path/to/rom")
        return
    }

    gameboyFile, err := core.LoadGameboyFromFile(path)
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    log.Printf("Loaded %d bytes", len(gameboyFile.Data))

    log.Printf("Gameboy file '%v'", gameboyFile.GetTitle())
    log.Printf("Rom size: %v", gameboyFile.GetRomSize())
    log.Printf("CGB Flag: %v", gameboyFile.GetCGBFlag())
    log.Printf("SGB Flag: %v", gameboyFile.GetSGBFlag())
    log.Printf("Cartidge type: 0x%x", gameboyFile.GetCartridgeType())

    makeCpu := func() (*core.CPU, error) {
        mbc, err := core.MakeMBC(gameboyFile.GetCartridgeType(), gameboyFile.GetRom())
        if err != nil {
            return nil, fmt.Errorf("unhandled cartridge type 0x%x: %v", gameboyFile.GetCartridgeType(), err)
        }

        cpu := core.MakeCPU(mbc)
        cpu.InitializeDMG()
        cpu.Debug = *cpuDebug
        cpu.Error = true
        cpu.PPU.Debug = *ppuDebug
        return cpu, nil
    }
    // cpu.PC = 0x100

    if *maxCycle > 0 {
        log.Printf("Max cycles: %v, %0.3f seconds", *maxCycle, float64(*maxCycle) / (float64(core.CPUSpeed) * (*speed)))
    }

    audioContext := audio.NewContext(SampleRate)

    engine, err := MakeEngine(makeCpu, *maxCycle, int64(*fps), *speed, audioContext)
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    ebiten.SetTPS(*fps)
    ebiten.SetWindowSize(core.ScreenWidth*4, core.ScreenHeight*4)
    ebiten.SetWindowTitle("Gameboy Emulator")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
