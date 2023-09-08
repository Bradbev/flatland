package editors

import (
	"fmt"
	"image/color"
	"reflect"

	"github.com/bradbev/flatland/src/asset"
	"github.com/bradbev/flatland/src/editor"
	"github.com/bradbev/flatland/src/editor/edgui"
	"github.com/bradbev/flatland/src/flat"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/inkyblackness/imgui-go/v4"
)

func worldEd(context *editor.TypeEditContext, value reflect.Value) error {
	world := value.Addr().Interface().(*flat.World)
	c, firstTime := editor.GetContext[worldEdContext](context, value)
	_ = c
	if firstTime {
		c.world = world
		c.buildWorldTree()
	}
	flags := imgui.TableFlagsSizingStretchSame |
		imgui.TableFlagsResizable |
		imgui.TableFlagsBordersOuter |
		imgui.TableFlagsBordersV |
		imgui.TableFlagsContextMenuInBody
	if imgui.BeginTableV(context.ID("WorldEd##"), 2, flags, imgui.Vec2{}, 0) {
		defer imgui.EndTable()

		imgui.TableNextRow()
		imgui.TableSetColumnIndex(0)
		c.renderOutliner(world, context, value)

		imgui.TableSetColumnIndex(1)
		c.renderWorld(world, context, value)
	}

	return nil
}

type worldEdContext struct {
	world *flat.World
	root  *worldTreeNode
}

func (w *worldEdContext) renderWorld(world *flat.World, context *editor.TypeEditContext, value reflect.Value) {
	width := imgui.ColumnWidth()
	id, img := context.Ed.GetImguiTexture(value, width, width)
	img.Fill(color.RGBA{0x70, 0x70, 0x70, 0xFF})
	w.world.Draw(img)
	for i, a := range w.world.PersistentActors {
		g := ebiten.GeoM{}
		flat.ApplyTransform(a.GetTransform(), &g)
		x, y := g.Apply(0, 0)
		s := fmt.Sprintf("%v", i)
		ebitenutil.DebugPrintAt(img, s, int(x), int(y))
	}
	imgui.Image(id, imgui.Vec2{X: f32(width), Y: f32(width)})
}

func (w *worldEdContext) renderOutliner(world *flat.World, context *editor.TypeEditContext, value reflect.Value) {
	imgui.Text("Outliner")
	edgui.DrawTree(w.root, nil)

	imgui.Separator()
	editor.StructEd(context, value)
}

func (w *worldEdContext) buildWorldTree() {
	w.root = &worldTreeNode{
		name: "Root",
	}
	for _, c := range w.world.PersistentActors {
		name, _ := asset.ObjectTypeName(c)
		w.root.children = append(w.root.children, &worldTreeNode{name: name})
	}

}

type worldTreeNode struct {
	name     string
	children []edgui.TreeNode
	selected bool
}

var _ edgui.TreeNode = (*worldTreeNode)(nil)

func (c *worldTreeNode) Name() string {
	return c.name
}
func (c *worldTreeNode) Children() []edgui.TreeNode { return c.children }
func (c *worldTreeNode) Leaf() bool                 { return len(c.children) == 0 }
func (c *worldTreeNode) Expanded() bool             { return true }
func (c *worldTreeNode) Selected() bool             { return c.selected }
