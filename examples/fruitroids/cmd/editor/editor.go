package main

import (
	"flatland/examples/fruitroids/src/fruitroids"
	"flatland/src/editor"
	"flatland/src/flat"
	"flatland/src/flat/editors"
	"fmt"

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
	game := &fruitroids.Fruitroids{}
	gg.ed.StartGame(game)

	ebiten.RunGame(gg)
}

type G struct {
	mgr  *renderer.Manager
	w, h int

	ed *editor.ImguiEditor
}

func (g *G) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %.3f\nFPS: %.2f\n", ebiten.ActualTPS(), ebiten.ActualFPS()), 11, 2)
	g.mgr.Draw(screen)
}

func (g *G) Update() error {
	g.mgr.Update(1.0 / 60.0)
	g.mgr.BeginFrame()
	{
		g.ed.Update(1.0 / float32(ebiten.ActualTPS()))
	}
	g.mgr.EndFrame()
	return nil
}

func (g *G) Layout(outsideWidth, outsideHeight int) (int, int) {
	/*
		if g.retina {
			m := ebiten.DeviceScaleFactor()
			g.w = int(float64(outsideWidth) * m)
			g.h = int(float64(outsideHeight) * m)
		} else {
			g.w = outsideWidth
			g.h = outsideHeight
		}
	*/
	g.w = outsideWidth
	g.h = outsideHeight
	g.mgr.SetDisplaySize(float32(g.w), float32(g.h))
	return g.w, g.h
}
