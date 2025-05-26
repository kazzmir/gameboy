.PHONY: gameboy

all: gameboy

gameboy:
	go mod tidy
	go build -o gameboy ./emulator

wasm: gameboy.wasm

gameboy.wasm:
	env GOOS=js GOARCH=wasm go build -o gameboy.wasm ./emulator

vet:
	go vet ./...
