Gameboy emulator written in Golang and Ebiten

# Keys

 * up/down/left/right
 * A: gameboy A
 * S: gameboy B
 * enter: gameboy start
 * space: gameboy select
 * P: pause/unpause
 * R: restart

# Online demo

Player in a browser:
https://kazzmir.itch.io/gameboy

# Build

Extra packages needed for ebiten
https://ebitengine.org/en/documents/install.html

```
$ go mod tidy
$ go build -o gameboy ./emulator
```
or
```
$ make
```

# Screenshots
![megaman](./images/screenshot.png)
