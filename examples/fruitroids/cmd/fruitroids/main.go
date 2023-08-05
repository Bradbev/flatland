package main

import (
	"flatland/examples/fruitroids/src/fruitroids"
	"flatland/src/flat"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	gg := &fruitroids.Fruitroids{}

	flat.RegisterAllFlatTypes()

	ebiten.RunGame(gg)
}
