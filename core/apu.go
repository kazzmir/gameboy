package core

import (
    "log"
    "sync"
    "math"
    // "math/rand/v2"
)

type Pulse struct {
    Enabled bool
    PanLeft bool
    PanRight bool

    LengthEnable bool

    Duty uint8
    DutyIndex uint8
    Length uint8

    // 4-bit volume
    Volume uint8
    EnvelopeDirection int8 // -1 for decrease, 1 for increase
    EnvelopeSweep uint8 // 3-bit sweep
    envelopeCounter uint16
    envelopeSweepCounter uint8

    Period uint16

    PeriodHigh uint16
    PeriodLow uint16

    Pace uint8
    Direction uint8
    Step uint8

    cycles uint64
}

func (pulse *Pulse) Trigger() {
    pulse.Enabled = true
    pulse.Period = (pulse.PeriodHigh << 8) | pulse.PeriodLow
    // FIXME:
    //   expire length timer
    //   envelope timer is reset
    //   volume is reset to channel volume
    //   a bunch of sweep stuff happens
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

func (pulse *Pulse) SetDuty(duty uint8) {
    pulse.Duty = duty
}

func (pulse *Pulse) SetLength(length uint8) {
    pulse.Length = length & 0b111111
}

func (pulse *Pulse) SetPeriodHigh(value uint8) {
    pulse.PeriodHigh = uint16(value & 0b111)
}

func (pulse *Pulse) SetPeriodLow(value uint8) {
    pulse.PeriodLow = uint16(value)
}

func (pulse *Pulse) generateSample() float32 {
    var table []byte
    switch pulse.Duty {
        case 0:
            table = []byte{0, 0, 0, 0, 0, 0, 0, 1}
        case 1:
            table = []byte{0, 0, 0, 0, 0, 0, 1, 1}
        case 2:
            table = []byte{0, 0, 0, 0, 1, 1, 1, 1}
        case 3:
            table = []byte{0, 0, 1, 1, 1, 1, 1, 1}
    }

    volume := float32(pulse.Volume) / 15.0

    value := table[pulse.DutyIndex]
    if value == 0 {
        return -volume
    } else {
        return volume
    }
}

func (pulse *Pulse) GenerateLeftSample() float32 {
    if pulse.Enabled && pulse.PanLeft {
        return pulse.generateSample()
    }

    return 0
}

func (pulse *Pulse) GenerateRightSample() float32 {
    if pulse.Enabled && pulse.PanRight {
        return pulse.generateSample()
    }

    return 0
}

// run 1 cycle
func (pulse *Pulse) Run() {
    pulse.envelopeCounter += 1
    // tick at 64hz, which is every 65536 cycles
    if pulse.envelopeCounter == 0 {
        pulse.envelopeSweepCounter += 1
        if pulse.envelopeSweepCounter >= pulse.EnvelopeSweep {
            pulse.envelopeSweepCounter = 0
            if pulse.EnvelopeDirection == -1 {
                if pulse.Volume > 0 {
                    pulse.Volume -= 1
                }
            } else {
                if pulse.Volume < 15 {
                    pulse.Volume += 1
                }
            }
        }
    }

    pulse.cycles += 1
    for pulse.cycles >= 4 {
        pulse.cycles -= 4
        pulse.Period += 1
        if pulse.Period >= 2048 {
            pulse.Period = (pulse.PeriodHigh << 8) | pulse.PeriodLow
            pulse.DutyIndex += 1
            if pulse.DutyIndex >= 8 {
                pulse.DutyIndex = 0
            }
        }
    }
}

func (pulse *Pulse) DoLength() {
    if pulse.LengthEnable {
        if pulse.Length < 64 {
            pulse.Length += 1
            if pulse.Length == 64 {
                pulse.Enabled = false
            }
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

    SampleCounter float32
    SampleRate uint32

    DivCounter uint16
    DivTicks uint16

    audioLock sync.Mutex
    audioBuffer []float32
    audioBufferIndex int
}

func MakeAPU(sampleRate uint32) *APU {
    return &APU{
        SampleCounter: float32(CPUSpeed) / float32(sampleRate),
        SampleRate: sampleRate,
    }
}

type AudioStream struct {
    APU *APU
}

func (stream *AudioStream) Read(data []byte) (int, error) {
    floatData := stream.APU.GetAudioBuffer(len(data) / 4 / 2)

    for i := range len(floatData) {
        v := math.Float32bits(floatData[i])
        data[i*4+0] = byte(v)
        data[i*4+1] = byte(v >> 8)
        data[i*4+2] = byte(v >> 16)
        data[i*4+3] = byte(v >> 24)
    }

    stream.APU.ReleaseAudioBuffer(floatData)

    // log.Printf("Read %v audio bytes out of %v bytes", len(floatData) * 4, len(data))

    return len(floatData) * 4, nil
}

func (apu *APU) GetAudioStream() *AudioStream {
    return &AudioStream{
        APU: apu,
    }
}

// samples is the number of samples in one channel
func (apu *APU) GetAudioBuffer(samples int) []float32 {
    apu.audioLock.Lock()
    if len(apu.audioBuffer) != samples * 2 {
        apu.audioBuffer = make([]float32, samples * 2)
    }
    log.Printf("audio buffer index %v: %v bytes", apu.audioBufferIndex, apu.audioBufferIndex * 4)
    // return apu.audioBuffer[:apu.audioBufferIndex]
    return apu.audioBuffer
}

func (apu *APU) ReleaseAudioBuffer(buffer []float32) {
    apu.audioBufferIndex = 0
    apu.audioLock.Unlock()
}

func (apu *APU) SetPulse1Volume(value uint8) {
    // volume top 4 bits, envelope direction bit 4, sweep page low 3 bits
    volume := (value & 0b1111_0000) >> 4
    envelopeDirection := (value & 0b0000_1000) >> 3
    sweep := value & 0b111

    apu.Pulse1.Volume = volume
    if envelopeDirection == 0 {
        apu.Pulse1.EnvelopeDirection = -1
    } else {
        apu.Pulse1.EnvelopeDirection = 1
    }

    apu.Pulse1.EnvelopeSweep = sweep
    apu.Pulse1.envelopeSweepCounter = 0
}

func (apu *APU) SetPulse1Sweep(value uint8) {
    pace := (value & 0b111_0000) >> 4
    direction := (value & 0b1_000) >> 3
    step := value & 0b111
    apu.Pulse1.SetSweep(pace, direction, step)
}

func (apu *APU) SetPulse1Duty(value uint8) {
    duty := (value & 0b11_000000) >> 6
    apu.Pulse1.SetDuty(duty)
    length := value & 0b111_111
    apu.Pulse1.SetLength(length)
}

func (apu *APU) SetPulse1PeriodHigh(value uint8) {
    period := value & 0b111
    apu.Pulse1.SetPeriodHigh(period)

    trigger := value & 0b1_0000000 != 0
    lengthEnable := value & 0b1_000000 != 0

    if trigger {
        apu.Pulse1.Trigger()
    }

    apu.Pulse1.LengthEnable = lengthEnable
    // log.Printf("length enable pulse 1: %v", apu.Pulse1.LengthEnable)
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

func (apu *APU) GenerateLeftSample() float32 {
    // return rand.Float32() * 2 - 1 // Generate a random float between -1 and 1

    return apu.Pulse1.GenerateLeftSample()
}

func (apu *APU) GenerateRightSample() float32 {
    return apu.Pulse1.GenerateRightSample()
}

func (apu *APU) Run(cycles uint64) {
    for cycles > 0 {
        cycles -= 1
        apu.Pulse1.Run()

        apu.DivCounter += 1
        if apu.DivCounter >= 512 {
            apu.DivCounter -= 512
            apu.DivTicks += 1

            if apu.DivTicks % 8 == 0 {
                // FIXME: envelope sweep
            }

            if apu.DivTicks % 2 == 0 {
                apu.Pulse1.DoLength()
            }

            if apu.DivTicks % 4 == 0 {
                // FIXME: channel1 frequency sweep
            }
        }

        // generate 44.1khz samples, one sample every 'cpu speed'/'sample rate' cycles
        apu.SampleCounter -= 1
        if apu.SampleCounter <= 0 {
            // emit sample
            // log.Printf("Emitting sample at %d Hz", apu.SampleRate)

            apu.SampleCounter += float32(CPUSpeed) / float32(apu.SampleRate)

            apu.audioLock.Lock()
            if apu.audioBufferIndex < len(apu.audioBuffer) - 1 {
                apu.audioBuffer[apu.audioBufferIndex] = apu.GenerateLeftSample()
                apu.audioBufferIndex += 1
                apu.audioBuffer[apu.audioBufferIndex] = apu.GenerateRightSample()
                apu.audioBufferIndex += 1
            } else {
                // log.Printf("drop audio sample")
            }
            apu.audioLock.Unlock()
        }
    }
}
