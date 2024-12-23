package core

import (
    "io"
    "os"
)

type GameboyFile struct {
    Data []byte
}

func LoadGameboy(reader io.Reader) (*GameboyFile, error) {
    data, err := io.ReadAll(reader)

    return &GameboyFile{Data: data}, err
}

func LoadGameboyFromFile(filename string) (*GameboyFile, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    return LoadGameboy(file)
}
