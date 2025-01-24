package main

import (
    "os"
    "log"
    "fmt"
    "reflect"
    "compress/gzip"
    "encoding/json"

    "github.com/kazzmir/gameboy/core"
)

func loadData(data map[string]interface{}) core.CPU {
    cpu := core.CPU{
    }

    cpu.SetRegister8(core.R8A, uint8(data["a"].(float64)))
    cpu.SetRegister8(core.R8B, uint8(data["b"].(float64)))
    cpu.SetRegister8(core.R8C, uint8(data["c"].(float64)))
    cpu.SetRegister8(core.R8D, uint8(data["d"].(float64)))
    cpu.SetRegister8(core.R8E, uint8(data["e"].(float64)))
    cpu.SetRegister8(core.R8H, uint8(data["h"].(float64)))
    cpu.SetRegister8(core.R8L, uint8(data["l"].(float64)))
    cpu.F = uint8(data["f"].(float64))
    cpu.PC = uint16(data["pc"].(float64))
    cpu.SP = uint16(data["sp"].(float64))

    ramContents := data["ram"].([]interface{})

    cpu.Ram = make([]uint8, 0x10000)

    for _, ram := range ramContents {
        values := ram.([]interface{})
        address := uint16(values[0].(float64))
        value := uint8(values[1].(float64))
        cpu.Ram[address] = value
    }

    return cpu
}

func runTest(test map[string]interface{}) {
    name := test["name"]
    initial := test["initial"].(map[string]interface{})
    final := test["final"].(map[string]interface{})
    cycles := test["cycles"]
    _ = cycles

    log.Printf("Running test: %v", name)

    cpu := loadData(initial)

    for range 10000 {
        instruction, amount := core.DecodeInstruction(cpu.Ram[cpu.PC:])
        if instruction.Opcode == core.Nop {
            break
        }
        log.Printf("Instruction: %+v amount: %v", instruction, amount)
        cpu.Execute(instruction)
    }

    // log.Printf("%+v", cpu)

    expected := loadData(final)

    if cpu.A != expected.A {
        log.Printf("A register mismatch: %v != %v", cpu.A, expected.A)
    }

    if cpu.BC != expected.BC {
        log.Printf("BC register mismatch: %v != %v", cpu.BC, expected.BC)
    }

    if cpu.DE != expected.DE {
        log.Printf("DE register mismatch: %v != %v", cpu.DE, expected.DE)
    }

    if cpu.HL != expected.HL {
        log.Printf("HL register mismatch: %v != %v", cpu.HL, expected.HL)
    }

    if cpu.F != expected.F {
        log.Printf("F register mismatch: %v != %v", cpu.F, expected.F)
    }

    if cpu.PC != expected.PC {
        log.Printf("PC register mismatch: %v != %v", cpu.PC, expected.PC)
    }

    for i, value := range cpu.Ram {
        if value != expected.Ram[i] {
            log.Printf("Ram mismatch at address 0x%x: %v != %v", i, value, expected.Ram[i])
        }
    }

}

func doTest(path string) error {
    reader, err := os.Open(path)
    if err != nil {
        return err
    }
    defer reader.Close()

    gzReader, err := gzip.NewReader(reader)
    if err != nil {
        return err
    }

    decoder := json.NewDecoder(gzReader)
    var data interface{}
    err = decoder.Decode(&data)
    if err != nil {
        return err
    }

    tests, ok := data.([]interface{})
    if !ok {
        return fmt.Errorf("Invalid data type: %v", reflect.TypeOf(data))
    }

    for _, test := range tests {
        testData, ok := test.(map[string]interface{})
        if !ok {
            log.Printf("Invalid test data type: %v", reflect.TypeOf(test))
        } else {
            runTest(testData)
        }

        break
    }

    return nil
}

func main(){
    path := "test-files/00.json.gz"
    err := doTest(path)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
