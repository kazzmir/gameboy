.PHONY: gameboy gameboy.wasm

all: gameboy

gameboy:
	go mod tidy
	go build -o gameboy ./emulator

wasm: gameboy.wasm

gameboy.wasm:
	env GOOS=js GOARCH=wasm go build -o gameboy.wasm ./emulator

itch.io: gameboy.wasm
	cp gameboy.wasm itch.io
	butler push itch.io kazzmir/gameboy:html

vet:
	go vet ./...
