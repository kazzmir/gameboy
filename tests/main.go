package main

import (
    "os"
    "log"
    "compress/gzip"
    "encoding/json"
)

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

    log.Printf("Data: %v", data)

    return nil
}

func main(){
    path := "test-files/00.json.gz"
    err := doTest(path)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
