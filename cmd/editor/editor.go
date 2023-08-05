package main

import (
	"flatland/src/asset"
	"flatland/src/editor"
	"flatland/src/flat"
	"flatland/src/flat/editors"
	"fmt"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/inkyblackness/imgui-go/v4"
)

// Example with the main Demo window and ClipMask

func main() {
	mgr := renderer.New(nil)

	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	gg := &G{
		mgr:    mgr,
		dscale: ebiten.DeviceScaleFactor(),
		ed:     editor.New("./content", mgr),
	}

	asset.RegisterAsset(editTest{})
	flat.RegisterAllFlatTypes()
	editors.RegisterAllFlatEditors(gg.ed)

	ebiten.RunGame(gg)
}

type nestedIndirect struct {
	NestedStr string
}

type editTest struct {
	Flt    float32
	Flt64  float64
	Bool   bool
	String string
	Int    int
	hidden float32

	// Path is filtered using the tag "filter" to files containing the text "json"
	Path asset.Path `filter:"json"`

	// NestedImmediate is renamed using the tag "flat"
	NestedImmediate struct {
		NestedFloat  float32
		NestedFloat2 float32
	} `flat:"Override field name from Nested Immediate"`
	NestedIndirect             nestedIndirect
	LastOne                    bool
	SupportNestedCustomEditors flat.Image
}

var test editTest

type G struct {
	mgr *renderer.Manager
	// demo members:
	showDemoWindow bool
	dscale         float64
	retina         bool
	w, h           int

	ed *editor.ImguiEditor
}

func (g *G) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %.3f\nFPS: %.2f\n", ebiten.ActualTPS(), ebiten.ActualFPS()), 11, 2)
	g.mgr.Draw(screen)
}

var autoS string

func (g *G) Update() error {
	g.mgr.Update(1.0 / 60.0)
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.mgr.ClipMask = !g.mgr.ClipMask
	}
	g.mgr.BeginFrame()
	{
		g.ed.Update(1.0 / float32(ebiten.ActualTPS()))
		g.debugWindow()
		func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Println(err)
					fmt.Println("Recovered")
				}
			}()

			g.ed.Edit(&test)
		}()
	}
	g.mgr.EndFrame()
	return nil
}

func (g *G) debugWindow() {
	defer imgui.End()
	if !imgui.Begin("Debug") {
		return
	}
	imgui.Checkbox("Retina", &g.retina)              // Edit bools storing our window open/close state
	imgui.Checkbox("Demo Window", &g.showDemoWindow) // Edit bools storing our window open/close state

	if g.showDemoWindow {
		imgui.ShowDemoWindow(&g.showDemoWindow)
	}
}

func (g *G) Layout(outsideWidth, outsideHeight int) (int, int) {
	if g.retina {
		m := ebiten.DeviceScaleFactor()
		g.w = int(float64(outsideWidth) * m)
		g.h = int(float64(outsideHeight) * m)
	} else {
		g.w = outsideWidth
		g.h = outsideHeight
	}
	g.mgr.SetDisplaySize(float32(g.w), float32(g.h))
	return g.w, g.h
}
