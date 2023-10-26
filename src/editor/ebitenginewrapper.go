package editor

import (
	"fmt"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func NewEbitengineWrapper(w, h int) *EbitengineWrapper {
	mgr := renderer.New(nil)

	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	wrapper := &EbitengineWrapper{
		ImguiManager: mgr,
		Editor:       New("./content", mgr),
	}

	return wrapper
}

type EbitengineWrapper struct {
	ImguiManager *renderer.Manager
	Editor       *ImguiEditor
	w, h         int
}

func (g *EbitengineWrapper) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %.3f (%dx%d)\nFPS: %.2f\n", ebiten.ActualTPS(), g.w, g.h, ebiten.ActualFPS()), 11, 20)
	g.ImguiManager.Draw(screen)
}

func (g *EbitengineWrapper) Update() error {
	updateRate := float32(1.0 / 60.0)
	var err error

	g.ImguiManager.Update(updateRate)
	g.ImguiManager.BeginFrame()
	{
		err = g.Editor.Update(updateRate)
	}
	g.ImguiManager.EndFrame()
	return err
}

func (g *EbitengineWrapper) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.w = outsideWidth
	g.h = outsideHeight
	g.ImguiManager.SetDisplaySize(float32(g.w), float32(g.h))
	return g.w, g.h
}
