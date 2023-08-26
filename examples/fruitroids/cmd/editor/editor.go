package main

import (
	"fmt"

	"github.com/bradbev/flatland/examples/fruitroids/src/fruitroids"
	"github.com/bradbev/flatland/src/asset"
	"github.com/bradbev/flatland/src/editor"
	"github.com/bradbev/flatland/src/flat"
	"github.com/bradbev/flatland/src/flat/editors"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func main() {
	mgr := renderer.New(nil)

	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	gg := &G{
		mgr: mgr,
		ed:  editor.New("./content", mgr),
	}

	flat.RegisterAllFlatTypes()
	editors.RegisterAllFlatEditors(gg.ed)
	// add the game specific types
	fruitroids.RegisterFruitroidTypes()

	gg.ed.StartGameCallback(func() ebiten.Game {
		world, err := asset.Load("world.json")
		fmt.Println(err)
		game := &fruitroids.Fruitroids{}
		game.World = world.(*flat.World)
		fruitroids.ActiveWorld = game
		game.World.BeginPlay()
		return game
	})

	ebiten.RunGame(gg)
}

type G struct {
	mgr  *renderer.Manager
	w, h int

	ed *editor.ImguiEditor
}

func (g *G) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %.3f\nFPS: %.2f\n", ebiten.ActualTPS(), ebiten.ActualFPS()), 11, 20)
	g.mgr.Draw(screen)
}

func (g *G) Update() error {
	updateRate := float32(1.0 / 60.0)
	var err error

	g.mgr.Update(updateRate)
	g.mgr.BeginFrame()
	{
		err = g.ed.Update(updateRate)
	}
	g.mgr.EndFrame()
	return err
}

func (g *G) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.w = outsideWidth
	g.h = outsideHeight
	g.mgr.SetDisplaySize(float32(g.w), float32(g.h))
	return g.w, g.h
}
