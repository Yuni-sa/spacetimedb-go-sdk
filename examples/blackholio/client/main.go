package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	game := NewGame()

	ebiten.SetWindowSize(DEFAULT_SCREEN_WIDTH, DEFAULT_SCREEN_HEIGHT)
	ebiten.SetWindowTitle("Blackholio - Space Arena")
	ebiten.SetFullscreen(true)
	ebiten.SetTPS(ebiten.SyncWithFPS) // could be cool, may cause issues

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
