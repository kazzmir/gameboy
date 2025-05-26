package core

type Pulse struct {
    Enabled bool
    PanLeft bool
    PanRight bool

    Period uint16

    PeriodHigh uint16
    PeriodLow uint16

    Pace uint8
    Direction uint8
    Step uint8

    cycles uint64
}

func (pulse *Pulse) SetSweep(pace uint8, direction uint8, step uint8) {
    pulse.Pace = pace
    pulse.Direction = direction
    pulse.Step = step
}

func (pulse *Pulse) SetPanning(left bool, right bool) {
    pulse.PanLeft = left
    pulse.PanRight = right
}

func (pulse *Pulse) SetPeriodHigh(value uint8) {
    pulse.PeriodHigh = uint16(value & 0b111)
}

func (pulse *Pulse) SetPeriodLow(value uint8) {
    pulse.PeriodLow = uint16(value)
}

// run 1 cycle
func (pulse *Pulse) Run() {
    pulse.cycles += 1
    for pulse.cycles >= 4 {
        pulse.cycles -= 4
        pulse.Period += 1
        if pulse.Period >= 2048 {
            pulse.Period = (pulse.PeriodHigh << 8) | pulse.PeriodLow
        }
    }
}

type Wave struct {
    Enabled bool
    PanLeft bool
    PanRight bool
}

func (wave *Wave) SetPanning(left bool, right bool) {
    wave.PanLeft = left
    wave.PanRight = right
}

type Noise struct {
    Enabled bool
    PanLeft bool
    PanRight bool
}

func (noise *Noise) SetPanning(left bool, right bool) {
    noise.PanLeft = left
    noise.PanRight = right
}

type APU struct {
    Pulse1 Pulse
    Pulse2 Pulse
    Wave Wave
    Noise Noise
    MasterEnabled bool
}

func MakeAPU() *APU {
    return &APU{
    }
}

type AudioStream struct {
}

func (stream *AudioStream) Read(data []byte) (int, error) {
    return 0, nil
}

func (apu *APU) GetAudioStream() *AudioStream {
    return &AudioStream{}
}

func (apu *APU) SetPulse1Sweep(value uint8) {
    pace := (value & 0b111_0000) >> 4
    direction := (value & 0b1_000) >> 3
    step := value & 0b111
    apu.Pulse1.SetSweep(pace, direction, step)
}

func (apu *APU) SetPulse1PeriodHigh(value uint8) {
    apu.Pulse1.SetPeriodHigh(value)
}

func (apu *APU) SetPulse1PeriodLow(value uint8) {
    apu.Pulse1.SetPeriodLow(value)
}

func (apu *APU) SetMasterEnabled(enabled bool) {
    apu.MasterEnabled = enabled
}

func (apu *APU) SetPanning(value uint8) {
    ch4_left  := value & 0b1000_0000 != 0
    ch3_left  := value & 0b0100_0000 != 0
    ch2_left  := value & 0b0010_0000 != 0
    ch1_left  := value & 0b0001_0000 != 0
    ch4_right := value & 0b0000_1000 != 0
    ch3_right := value & 0b0000_0100 != 0
    ch2_right := value & 0b0000_0010 != 0
    ch1_right := value & 0b0000_0001 != 0

    apu.Pulse1.SetPanning(ch1_left, ch1_right)
    apu.Pulse2.SetPanning(ch2_left, ch2_right)
    apu.Wave.SetPanning(ch3_left, ch3_right)
    apu.Noise.SetPanning(ch4_left, ch4_right)
}

func (apu *APU) ReadSoundPanning() uint8 {
    var out uint8

    if apu.Pulse1.PanLeft {
        out |= 0b0001_0000
    }
    if apu.Pulse1.PanRight {
        out |= 0b0000_0001
    }

    if apu.Pulse2.PanLeft {
        out |= 0b0010_0000
    }
    if apu.Pulse2.PanRight {
        out |= 0b0000_0010
    }

    if apu.Wave.PanLeft {
        out |= 0b0100_0000
    }
    if apu.Wave.PanRight {
        out |= 0b0000_0100
    }

    if apu.Noise.PanLeft {
        out |= 0b1000_0000
    }
    if apu.Noise.PanRight {
        out |= 0b0000_1000
    }

    return out
}

func (apu *APU) ReadMasterControl() uint8 {
    var out uint8
    if apu.MasterEnabled {
        out |= 0x80
    }

    if apu.Pulse1.Enabled {
        out |= 0x01
    }

    if apu.Pulse2.Enabled {
        out |= 0x02
    }

    if apu.Wave.Enabled {
        out |= 0x04
    }

    if apu.Noise.Enabled {
        out |= 0x08
    }

    return out
}

func (apu *APU) Run(cycles uint64) {
    for cycles > 0 {
        cycles -= 1
        apu.Pulse1.Run()
    }
}
