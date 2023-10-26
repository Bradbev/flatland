package main

import (
	"fmt"
	"reflect"

	"github.com/bradbev/flatland/cmd/editor/edtest"
	"github.com/bradbev/flatland/src/asset"
	"github.com/bradbev/flatland/src/editor"
	"github.com/bradbev/flatland/src/editor/edgui"
	"github.com/bradbev/flatland/src/flat"
	"github.com/bradbev/flatland/src/flat/editors"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/inkyblackness/imgui-go/v4"
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

	edtest.TreeTestInit()

	asset.RegisterAsset(EditTest{})
	asset.RegisterAsset(EditTest2{})
	asset.RegisterAsset(actorTest{})
	asset.RegisterAsset(flat.TestMouseHandler{})
	asset.RegisterAsset(edtest.UiTestActor{})
	flat.RegisterAllFlatTypes()
	editors.RegisterAllFlatEditors(gg.ed)

	gg.ed.AddType(new(testInterfaceEditor), func(tec *editor.TypeEditContext, v reflect.Value) error {
		imgui.Text("TestTab")
		return nil
	})

	// load an asset to be edited by the test editor
	//asset.Save("testedit.json", &defaultTestObject)
	//gg.ed.EditAsset("testedit.json")
	gg.ed.EditAsset("edittest2.json")

	defaultTestObjectChild = defaultTestObject
	asset.SetParent(&defaultTestObjectChild, &defaultTestObject)
	//asset.Save("childedit.json", &defaultTestObjectChild)
	//gg.ed.EditAsset("childedit.json")

	//	gg.ed.EditAsset("uitestactor.json")
	//gg.ed.EditAsset("actorTest.json")
	//gg.ed.EditAsset("font.json")
	//gg.ed.EditAsset("world.json")

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

	gg.ed.RegisterEnum(map[any]string{
		First:    "First",
		Second:   "Second",
		Whatever: "Whatever",
	})

	ebiten.RunGame(gg)
}

type nestedIndirect struct {
	NestedStr string
}

type actorTest struct {
	flat.ActorBase
	ActorBaseName string
}

func (a *actorTest) BeginPlay() {
	a.ActorBase.BeginPlay(a)
}

func (a *actorTest) TestTab() {}

type testInterfaceEditor interface {
	TestTab()
}

type TestEnum int8

const (
	First TestEnum = iota
	Second
	Whatever
)

type EditTest2 struct {
	AssetType       asset.Asset
	AssetTypeInline asset.Asset `flat:"inline ; desc:Name Change"`
}

// EditTest demonstrates all the ways that the editor can
// edit types.
type EditTest struct {
	TestEnum        TestEnum
	AssetType       asset.Asset // support setting Assets
	RestrictedTypes *EditTest   // only show assets of type EditTest
	AssetTypeInline asset.Asset `flat:"inline"`
	Flt             float32
	Slice           []int
	Array           [3]float32
	StringSlice     []string
	StructSlice     []nestedIndirect
	Flt64           float64
	Bool            bool
	String          string
	Int             int
	hidden          float32

	// Path is filtered using the tag "filter" to files containing the text "json"
	Path asset.Path `flat:"desc:Path (json filter) ; filter:json" filter:"json"`

	// NestedImmediate is renamed using the tag "flat"
	NestedImmediate struct {
		NestedFloat  float32
		NestedFloat2 float32
	} `flat:"desc:Override field name from Nested Immediate"`

	NestedIndirectField        nestedIndirect
	SupportNestedCustomEditors flat.Image
}

var defaultTestObject = EditTest{
	Slice:  []int{7, 6},
	String: "Default Test Object",
}
var defaultTestObjectChild = EditTest{}

// eventually this struct will vanish and the whole loop
// will live in the editor
type G struct {
	mgr *renderer.Manager
	// demo members:
	showDemoWindow bool
	dscale         float64
	w, h           int
	showTestTree   bool
	showTestList   bool

	ed *editor.ImguiEditor
}

func (g *G) showTestControls() {
	if imgui.Begin("Test Controls") {
		imgui.Checkbox("ShowTestTree", &g.showTestTree)
		imgui.Checkbox("ShowTestList", &g.showTestList)
		imgui.Checkbox("ShowDemoWindow", &g.showDemoWindow)
	}
	imgui.End()
}

func (g *G) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %.3f\nFPS: %.2f\n", ebiten.ActualTPS(), ebiten.ActualFPS()), 11, 20)
	x, y := ebiten.CursorPosition()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d,%d", x, y), 80, 20)
	g.mgr.Draw(screen)
}

func (g *G) Update() error {
	updateRate := float32(1.0 / 60.0)
	g.mgr.Update(updateRate)
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.mgr.ClipMask = !g.mgr.ClipMask
	}
	g.mgr.BeginFrame()
	{
		g.ed.Update(updateRate)

		g.showTestControls()
		if g.showDemoWindow {
			imgui.ShowDemoWindow(&g.showDemoWindow)
		}

		if g.showTestTree {
			edtest.TreeTest()
		}

		if g.showTestList {
			edtest.ListTest()
		}
	}
	g.mgr.EndFrame()
	return nil
}

func (g *G) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.w = outsideWidth
	g.h = outsideHeight
	g.mgr.SetDisplaySize(float32(g.w), float32(g.h))
	return g.w, g.h
}
