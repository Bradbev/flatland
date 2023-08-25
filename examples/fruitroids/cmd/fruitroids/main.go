package main

import (
	"fmt"
	"os"

	"github.com/bradbev/flatland/examples/fruitroids/src/fruitroids"
	"github.com/bradbev/flatland/src/asset"
	"github.com/bradbev/flatland/src/flat"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	gg := &fruitroids.Fruitroids{}

	fsysRead := os.DirFS("./content")
	asset.RegisterFileSystem(fsysRead, 0)

	flat.RegisterAllFlatTypes()
	fruitroids.RegisterFruitroidTypes()
	world, err := asset.Load("world.json")
	fruitroids.ActiveWorld = gg
	fmt.Println(err)
	gg.World = world.(*flat.World)
	gg.World.BeginPlay()

	ebiten.RunGame(gg)
}
