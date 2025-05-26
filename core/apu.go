package core

type Pulse struct {
    Enabled bool
    PanLeft bool
    PanRight bool
}

func (pulse *Pulse) SetPanning(left bool, right bool) {
    pulse.PanLeft = left
    pulse.PanRight = right
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
}
