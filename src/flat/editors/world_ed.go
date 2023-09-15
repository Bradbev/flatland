package editors

import (
	"fmt"
	"image/color"
	"reflect"
	"strings"

	"github.com/bradbev/flatland/src/asset"
	"github.com/bradbev/flatland/src/editor"
	"github.com/bradbev/flatland/src/editor/edgui"
	"github.com/bradbev/flatland/src/flat"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/inkyblackness/imgui-go/v4"
	"golang.org/x/exp/slices"
)

func worldEd(context *editor.TypeEditContext, value reflect.Value) error {
	world := value.Addr().Interface().(*flat.World)
	c, firstTime := editor.GetContext[worldEdContext](context, value)
	_ = c
	if firstTime {
		c.world = world
		c.buildWorldTree()
		c.addDialog = &addDialog{Title: "Add Actor##uniqueID"}
		c.addDialog.Context = c
		world.BeginPlay()
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
	world     *flat.World
	root      *worldTreeNode
	addDialog *addDialog
}

func (w *worldEdContext) renderWorld(world *flat.World, context *editor.TypeEditContext, value reflect.Value) {
	width := imgui.ColumnWidth()
	id, img := context.Ed.GetImguiTexture(value, width, width)
	img.Fill(color.RGBA{0x70, 0x70, 0x70, 0xFF})
	w.world.Draw(img)
	for i, a := range w.world.PersistentActors {
		if a == nil {
			continue
		}
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
	if imgui.Button("Add") {
		w.addDialog.Open()
	}
	if w.addDialog.Draw() {
		context.SetChanged()
	}

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

type addDialog struct {
	Title        string
	Context      *worldEdContext
	open         bool
	selectedItem *addAssetItem
	list         edgui.FilteredList[*addAssetItem]
}

func (a *addDialog) Clicked(node edgui.ListNode, index int) {
	assetItem := node.(*addAssetItem)
	if assetItem.selected {
		assetItem.selected = false
		a.selectedItem = nil
		return
	}
	for _, item := range a.list.List {
		item.selected = false
	}
	assetItem.selected = true
	a.selectedItem = assetItem
}
func (a *addDialog) DoubleClicked(node edgui.ListNode, index int) {

}

type addAssetItem struct {
	descriptor *asset.AssetDescriptor
	assetPath  string
	selected   bool
}

func (a *addAssetItem) Name() string {
	if a.descriptor != nil {
		return a.descriptor.Name
	}
	return a.assetPath
}

func (a *addAssetItem) Selected() bool {
	return a.selected
}

func (a *addDialog) Open() {
	a.open = true
	imgui.OpenPopup(a.Title)

	var items []*addAssetItem
	for _, desc := range asset.GetAssetDescriptors() {
		items = append(items, &addAssetItem{
			descriptor: desc,
		})
	}
	paths, _ := asset.FilterFilesByReflectType(reflect.TypeOf(new(flat.Actor)))
	for _, p := range paths {
		items = append(items, &addAssetItem{
			assetPath: p,
		})
	}
	slices.SortFunc(items, func(a, b *addAssetItem) int {
		return strings.Compare(a.Name(), b.Name())
	})
	a.list = edgui.FilteredList[*addAssetItem]{List: items}
}

func (a *addDialog) Draw() bool {
	result := false
	if imgui.BeginPopupModalV(a.Title, &a.open, imgui.WindowFlagsNone) {
		defer imgui.EndPopup()
		edgui.WithDisabled(a.selectedItem == nil, func() {
			if imgui.Button("Add") {
				var actorToAdd flat.Actor
				if a.selectedItem.descriptor == nil {
					parent, err := asset.Load(asset.Path(a.selectedItem.assetPath))
					flat.Check(err)
					instance, err := asset.NewInstance(parent)
					flat.Check(err)
					asset.SetParent(instance, parent)
					actorToAdd = instance.(flat.Actor)
				} else {
					instance, err := a.selectedItem.descriptor.Create()
					flat.Check(err)
					actorToAdd = instance.(flat.Actor)
				}
				world := a.Context.world
				world.AddToWorld(actorToAdd)
				world.PersistentActors = append(world.PersistentActors, actorToAdd)
				imgui.CloseCurrentPopup()
				result = true
			}
		})
		a.list.Draw("AddActorList", a)
	}
	return result
}
