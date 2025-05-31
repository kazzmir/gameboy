package core

import (
    // "log"
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
    envelopeSweepCounter uint8

    Period uint16

    PeriodHigh uint16
    PeriodLow uint16

    hasPeriodSweep bool
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

func (pulse *Pulse) SetVolume(volume uint8, envelopeDirection uint8, envelopeSweep uint8) {
    pulse.Volume = volume
    if envelopeDirection == 0 {
        pulse.EnvelopeDirection = -1
    } else {
        pulse.EnvelopeDirection = 1
    }

    pulse.EnvelopeSweep = envelopeSweep
    pulse.envelopeSweepCounter = 0
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

func (pulse *Pulse) doSweep(clock uint64) {
    if pulse.hasPeriodSweep && pulse.Pace > 0 {
        if clock % (CPUSpeed/128 * uint64(pulse.Pace)) == 0 {
            var oldPeriod int16 = int16((pulse.PeriodHigh << 8) | pulse.PeriodLow)
            switch pulse.Direction {
                case 0:
                    oldPeriod = oldPeriod + oldPeriod / (1 << pulse.Step)
                case 1:
                    oldPeriod = oldPeriod - oldPeriod / (1 << pulse.Step)
            }

            if oldPeriod > 0x7ff || oldPeriod < 0 {
                pulse.Enabled = false
            } else {
                pulse.PeriodHigh = uint16((oldPeriod >> 8) & 0x7)
                pulse.PeriodLow = uint16(oldPeriod & 0xff)
                pulse.Period = (pulse.PeriodHigh << 8) | pulse.PeriodLow
            }
        }
    }
}

func (pulse *Pulse) doVolume(clock uint64) {
    // tick at 64hz, which is every 65536 cycles
    if clock % (CPUSpeed/64) == 0 {
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
}

func (pulse *Pulse) doDutyCycle() {
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

// run 1 cycle
func (pulse *Pulse) Run(clock uint64) {
    pulse.doVolume(clock)
    pulse.doSweep(clock)
    pulse.doDutyCycle()

    if clock % (CPUSpeed/256) == 0 {
        pulse.DoLength()
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

func (wave *Wave) GenerateLeftSample() float32 {
    if wave.Enabled && wave.PanLeft {
    }

    return 0
}

func (wave *Wave) GenerateRightSample() float32 {
    if wave.Enabled && wave.PanLeft {
    }

    return 0
}

func (wave *Wave) Run(clock uint64) {
}

type Noise struct {
    Enabled bool
    PanLeft bool
    PanRight bool

    Volume uint8
    InitialVolume uint8
    envelopeSweepCounter uint8
    EnvelopeSweep uint8
    EnvelopeDirection int8

    LengthOriginal uint8
    Length uint8
    LengthEnable bool

    // 0 or 1, shifted out from lfsr
    LastBit uint8

    ClockShift uint8
    LFSR uint16 // 3-bit LFSR
    LFSRLength uint8 // 15 or 7, depending on LFSR type
    ClockDivider uint8 // 3-bit clock divider
}

func (noise *Noise) SetPanning(left bool, right bool) {
    noise.PanLeft = left
    noise.PanRight = right
}

func (noise *Noise) SetFrequency(clock_shift uint8, lfsr uint8, clock_divider uint8) {
    // log.Printf("noise: Set clock frequency=%v lfsr=%v divider=%v", clock_shift, lfsr, clock_divider)
    noise.ClockShift = clock_shift
    if lfsr == 0 {
        noise.LFSRLength = 15
    } else {
        noise.LFSRLength = 7
    }
    noise.ClockDivider = clock_divider
}

func (noise *Noise) SetVolume(volume uint8, envelopeDirection uint8, envelopeSweep uint8) {
    noise.Volume = volume
    noise.InitialVolume = volume
    if envelopeDirection == 0 {
        noise.EnvelopeDirection = -1
    } else {
        noise.EnvelopeDirection = 1
    }

    noise.EnvelopeSweep = envelopeSweep
    noise.envelopeSweepCounter = 0
}

func (noise *Noise) ResetLFSR() {
    noise.LFSR = 0xffff
}

func (noise *Noise) Trigger() {
    noise.Enabled = true
    /*
    if noise.Length >= 64 {
        noise.Length = noise.LengthOriginal
    }
    */

    noise.Length = noise.LengthOriginal

    // log.Printf("noise trigger, length=%v", noise.Length)

    noise.envelopeSweepCounter = 0
    noise.Volume = noise.InitialVolume
    noise.ResetLFSR()
}

func (noise *Noise) doVolume(clock uint64) {
    // tick at 64hz, which is every 65536 cycles
    if clock % (CPUSpeed/64) == 0 {
        noise.envelopeSweepCounter += 1
        if noise.envelopeSweepCounter >= noise.EnvelopeSweep {
            noise.envelopeSweepCounter = 0
            if noise.EnvelopeDirection == -1 {
                if noise.Volume > 0 {
                    noise.Volume -= 1
                }
            } else {
                if noise.Volume < 15 {
                    noise.Volume += 1
                }
            }
        }
    }
}

func (noise *Noise) Run(clock uint64) {
    if noise.Enabled {
        noise.doVolume(clock)

        noise.doLFSR(clock)
        // log.Printf("lfsr: 0x%x", noise.LFSR)

        if clock % (CPUSpeed/256) == 0 {
            noise.DoLength()
        }
    }
}

func (noise *Noise) doLFSR(clock uint64) {
    rate := uint64(262144 / (1 << noise.ClockShift))
    if noise.ClockDivider > 0 {
        rate /= uint64(noise.ClockDivider)
    } else {
        rate *= 2
    }

    if clock % (CPUSpeed/rate) == 0 {
        noise.LastBit = uint8(noise.LFSR & 1)
        noise.LFSR >>= 1
        newBit := (noise.LFSR & 1) ^ ((noise.LFSR & 0b10) >> 1)
        noise.LFSR |= newBit << 15
        if noise.LFSRLength == 7 {
            noise.LFSR |= newBit << 7
        }
    }
}

func (noise *Noise) DoLength() {
    if noise.LengthEnable {
        // log.Printf("noise length %v", noise.Length)
        if noise.Length < 64 {
            noise.Length += 1
        }

        if noise.Length >= 64 {
            // log.Printf("noise length disable")
            noise.Enabled = false
        }
    }
}

func (noise *Noise) GenerateLeftSample() float32 {
    if noise.Enabled && noise.PanLeft {
        var volume float32 = 1
        if noise.LastBit == 0 {
            volume = -1
        }

        scaled := float32(noise.Volume) / 15

        return volume * scaled
    } else {
        return 0
    }
}

func (noise *Noise) GenerateRightSample() float32 {
    if noise.Enabled && noise.PanRight {
        var volume float32 = 1
        if noise.LastBit == 0 {
            volume = -1
        }

        scaled := float32(noise.Volume) / 15

        return volume * scaled
    } else {
        return 0
    }
}

type APU struct {
    counter uint64
    Pulse1 Pulse
    Pulse2 Pulse
    Wave Wave
    Noise Noise
    MasterEnabled bool

    // 0-7, 0 is not entirely silent, just very quit. 7 is loudest
    LeftVolume uint8
    RightVolume uint8

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
        Pulse1: Pulse{
            hasPeriodSweep: true,
        },
        Pulse2: Pulse{
            hasPeriodSweep: false,
        },
        LeftVolume: 0x7,
        RightVolume: 0x7,
        Noise: Noise{
            LFSRLength: 15,
        },
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
    // log.Printf("audio buffer index %v: %v bytes", apu.audioBufferIndex, apu.audioBufferIndex * 4)
    // return apu.audioBuffer[:apu.audioBufferIndex]
    return apu.audioBuffer
}

func (apu *APU) ReleaseAudioBuffer(buffer []float32) {
    apu.audioBufferIndex = 0
    apu.audioLock.Unlock()
}

func (apu *APU) SetNoiseLength(value uint8) {
    length := value & 0b111_111
    apu.Noise.Length = length
    apu.Noise.LengthOriginal = length
}

func (apu *APU) SetNoiseVolume(value uint8) {
    volume := (value & 0b1111_0000) >> 4
    envelopeDirection := (value & 0b0000_1000) >> 3
    sweep := value & 0b111
    apu.Noise.SetVolume(volume, envelopeDirection, sweep)
}

func (apu *APU) SetNoiseControl(value uint8) {
    trigger := value & 0b1000_0000 != 0
    lengthEnable := value & 0b100_0000 != 0

    if trigger {
        apu.Noise.Trigger()
    }

    if lengthEnable {
        apu.Noise.LengthEnable = true
        apu.Noise.Length = 0
    } else {
        apu.Noise.LengthEnable = false
    }
}

func (apu *APU) SetNoiseFrequency(value uint8) {
    clock_shift := (value & 0b1111_0000) >> 4
    lfsr := (value & 0b1_000) >> 3
    clock_divider := value & 0b111
    apu.Noise.SetFrequency(clock_shift, lfsr, clock_divider)
}

func (apu *APU) SetMasterVolume(volume uint8) {
    // FIXME: use VIN left (bit 7) and VIN right (bit 3) to set volume

    left := (volume & 0b111_0000) >> 4
    right := (volume & 0b111)

    apu.LeftVolume = left
    apu.RightVolume = right
}

func (apu *APU) SetPulse1Volume(value uint8) {
    // volume top 4 bits, envelope direction bit 4, sweep page low 3 bits
    volume := (value & 0b1111_0000) >> 4
    envelopeDirection := (value & 0b0000_1000) >> 3
    sweep := value & 0b111
    apu.Pulse1.SetVolume(volume, envelopeDirection, sweep)
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

func (apu *APU) SetPulse2Volume(value uint8) {
    // volume top 4 bits, envelope direction bit 4, sweep page low 3 bits
    volume := (value & 0b1111_0000) >> 4
    envelopeDirection := (value & 0b0000_1000) >> 3
    sweep := value & 0b111
    apu.Pulse2.SetVolume(volume, envelopeDirection, sweep)
}

func (apu *APU) SetPulse2Duty(value uint8) {
    duty := (value & 0b11_000000) >> 6
    apu.Pulse2.SetDuty(duty)
    length := value & 0b111_111
    apu.Pulse2.SetLength(length)
}

func (apu *APU) SetPulse2PeriodHigh(value uint8) {
    period := value & 0b111
    apu.Pulse2.SetPeriodHigh(period)

    trigger := value & 0b1_0000000 != 0
    lengthEnable := value & 0b1_000000 != 0

    if trigger {
        apu.Pulse2.Trigger()
    }

    apu.Pulse2.LengthEnable = lengthEnable
    // log.Printf("length enable pulse 1: %v", apu.Pulse1.LengthEnable)
}

func (apu *APU) SetPulse2PeriodLow(value uint8) {
    apu.Pulse2.SetPeriodLow(value)
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

    sample := apu.Pulse1.GenerateLeftSample() + apu.Pulse2.GenerateLeftSample() + apu.Noise.GenerateLeftSample() + apu.Wave.GenerateLeftSample()
    // sample := apu.Noise.GenerateLeftSample()
    scaled := float32(apu.LeftVolume+1) / 8
    return sample * scaled
}

func (apu *APU) GenerateRightSample() float32 {
    sample := apu.Pulse1.GenerateRightSample() + apu.Pulse2.GenerateRightSample() + apu.Noise.GenerateRightSample() + apu.Wave.GenerateRightSample()
    // sample := apu.Noise.GenerateRightSample()
    scaled := float32(apu.RightVolume+1) / 8
    return sample * scaled
}

func (apu *APU) Run(cycles uint64) {
    if !apu.MasterEnabled {
        return
    }

    for cycles > 0 {
        cycles -= 1
        apu.counter += 1
        apu.Pulse1.Run(apu.counter)
        apu.Pulse2.Run(apu.counter)
        apu.Noise.Run(apu.counter)
        apu.Wave.Run(apu.counter)

        apu.DivCounter += 1
        if apu.DivCounter >= (CPUSpeed/512) {
            apu.DivCounter -= CPUSpeed/512
            apu.DivTicks += 1

            if apu.DivTicks % 2 == 0 {
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
