package main

import (
	"flatland/src/asset"
	"flatland/src/editor"
	"flatland/src/editor/edgui"
	"flatland/src/flat"
	"flatland/src/flat/editors"
	"fmt"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// This file should eventually slim down to almost nothing,
// just some code to register the game specific assets and
// the new game creation function

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

	// load an asset to be edited by the test editor
	a, err := asset.Load("apple-98.json")
	fmt.Println(err)
	defaultTestObject.AssetType = a
	asset.Save("testobj.json", &defaultTestObject)
	gg.ed.EditAsset("testobj.json")

	menu := edgui.Menu{
		Name: "Custom Item",
		Items: []*edgui.MenuItem{
			{
				Text: "Show Imgui Demo",
				Action: func(self *edgui.MenuItem) {
					gg.showDemoWindow = !gg.showDemoWindow
					self.Selected = gg.showDemoWindow
				},
			},
		},
	}
	gg.ed.AddMenu(menu)

	ebiten.RunGame(gg)
}

type nestedIndirect struct {
	NestedStr string
}

// editTest demonstrates all the ways that the editor can
// edit types.
type editTest struct {
	AssetType   asset.Asset // support setting Assets
	Flt         float32
	Slice       []int
	Array       [3]float32
	StringSlice []string
	StructSlice []nestedIndirect
	Flt64       float64
	Bool        bool
	String      string
	Int         int
	hidden      float32

	// Path is filtered using the tag "filter" to files containing the text "json"
	Path asset.Path `flat:"Path (json filter)" filter:"json"`

	// NestedImmediate is renamed using the tag "flat"
	NestedImmediate struct {
		NestedFloat  float32
		NestedFloat2 float32
	} `flat:"Override field name from Nested Immediate"`

	NestedIndirectField        nestedIndirect
	SupportNestedCustomEditors flat.Image
}

var defaultTestObject = editTest{
	Slice: []int{7, 4, 5, 6},
}

// eventually this struct will vanish and the whole loop
// will live in the editor
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
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %.3f\nFPS: %.2f\n", ebiten.ActualTPS(), ebiten.ActualFPS()), 11, 20)
	g.mgr.Draw(screen)
}

func (g *G) Update() error {
	g.mgr.Update(1.0 / 60.0)
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.mgr.ClipMask = !g.mgr.ClipMask
	}
	g.mgr.BeginFrame()
	{
		g.ed.Update(1.0 / float32(ebiten.ActualTPS()))
	}
	g.mgr.EndFrame()
	return nil
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
