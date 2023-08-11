package main

import (
	"flatland/examples/fruitroids/src/fruitroids"
	"flatland/src/asset"
	"flatland/src/flat"
	"fmt"
	"os"

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
	fmt.Println(err)
	gg.World = world.(*flat.World)
	gg.World.BeginPlay()
	//gg.World = flat.NewWorld()
	//ship, err := asset.Load("ship.json")
	//fmt.Println(err)
	//gg.World.AddActor(ship)

	ebiten.RunGame(gg)
}
