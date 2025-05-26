.PHONY: gameboy

all: gameboy

gameboy:
	go mod tidy
	go build -o gameboy ./emulator

vet:
	go vet ./...
