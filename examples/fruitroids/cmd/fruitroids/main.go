package main

import (
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

	// TOUR:Fruitroids 1
	// Create an embed filesystem and pass it to asset.  All content will
	// come from here.  Any fs.FS can be given to asset.
	// embed content to the binary so wasm distribution works
	fsysRead, _ := fs.Sub(content.Content, "content")
	asset.RegisterFileSystem(fsysRead, 0)

	// TOUR:Fruitroids 2
	// Register the struct and interface types that the game uses.
	// We need to be explicit and opt in to every type that will be
	// used by the asset package
	flat.RegisterAllFlatTypes()
	fruitroids.RegisterFruitroidTypes()

	// TOUR:Fruitroids 3
	// Load the world asset.  You can read world.json to get an idea of what goes into
	// an asset file, but basically the file stores the Go type (flat.World) and a json-ish
	// form of the asset.  Read assets.md for more detail on how assets work.
	// At this point you probably want to run fruitroids/cmd/editor/editor.go and examine
	// world.json.
	// Each actor in World.PersistentActors will be loaded and the their various interface functions
	// added to the World main loop
	world, err := asset.Load("world.json")
	flat.Check(err)

	// TOUR:Fruitroids 4
	// Create an ebiten.Game type, set it up and call RunGame
	gg := &fruitroids.Fruitroids{}
	gg.World = world.(*flat.World)
	fruitroids.ActiveWorld = gg
	gg.World.BeginPlay()

	ebiten.RunGame(gg)
}
