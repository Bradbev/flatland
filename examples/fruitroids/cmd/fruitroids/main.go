package main

import (
	"fmt"
	"io/fs"

	content "github.com/bradbev/flatland/examples/fruitroids"
	"github.com/bradbev/flatland/examples/fruitroids/src/fruitroids"
	"github.com/bradbev/flatland/src/asset"
	"github.com/bradbev/flatland/src/flat"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// embed content to the binary so wasm distribution works
	fsysRead, _ := fs.Sub(content.Content, "content")
	asset.RegisterFileSystem(fsysRead, 0)

	flat.RegisterAllFlatTypes()
	fruitroids.RegisterFruitroidTypes()
	world, err := asset.Load("world.json")
	fmt.Println(err)

	gg := &fruitroids.Fruitroids{}
	gg.World = world.(*flat.World)
	fruitroids.ActiveWorld = gg
	gg.World.BeginPlay()

	ebiten.RunGame(gg)
}
