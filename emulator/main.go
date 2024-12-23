package main

import (
    "os"
    "log"

    "github.com/kazzmir/gameboy/core"
)

func main(){
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
}
