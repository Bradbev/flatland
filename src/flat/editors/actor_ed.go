package editors

import (
	"image/color"
	"reflect"

	"github.com/bradbev/flatland/src/asset"
	"github.com/bradbev/flatland/src/editor"
	"github.com/bradbev/flatland/src/editor/edgui"
	"github.com/bradbev/flatland/src/flat"

	"github.com/inkyblackness/imgui-go/v4"
)

type actorEdContext struct {
	componentTreeRoot *componentTreeNode
	addDialog         *AddComponentDialog
	valueToEdit       reflect.Value
}

func actorEd(context *editor.TypeEditContext, value reflect.Value) error {
	actor := value.Addr().Interface().(flat.Actor)
	c, firstTime := editor.GetContext[actorEdContext](context, value)
	if firstTime {
		c.componentTreeRoot = buildComponentTree(actor)
		c.addDialog = &AddComponentDialog{Id: "addComponent"}
		c.valueToEdit = value
	}

	flags := imgui.TableFlagsSizingStretchSame |
		imgui.TableFlagsResizable |
		imgui.TableFlagsBordersOuter |
		imgui.TableFlagsBordersV |
		imgui.TableFlagsContextMenuInBody
	if imgui.BeginTableV(context.ID("ActorEd##"), 3, flags, imgui.Vec2{}, 0) {
		defer imgui.EndTable()
		imgui.TableNextRow()

		imgui.TableSetColumnIndex(0)
		componentTree(actor, c, context, value)

		imgui.TableSetColumnIndex(1)
		editor.StructEd(context, c.valueToEdit)

		imgui.TableSetColumnIndex(2)
		renderActor(context, value)
	}
	return nil
}

func renderActor(context *editor.TypeEditContext, value reflect.Value) error {
	imgui.Text("Rendered View")
	w := imgui.ColumnWidth()
	id, img := context.Ed.GetImguiTexture(value, int(w), int(w))
	img.Fill(color.Black)
	_ = id
	if tickable, ok := value.Addr().Interface().(flat.Tickable); ok {
		tickable.Tick(1 / 60.0)
	}
	if drawable, ok := value.Addr().Interface().(flat.Drawable); ok {
		drawable.Draw(img)
	}

	_ = id
	// tell imgui to draw the ebiten.Image
	imgui.Image(id, imgui.Vec2{X: f32(w), Y: f32(w)})
	return nil
}

func componentTree(actor flat.Actor, actorEd *actorEdContext, context *editor.TypeEditContext, value reflect.Value) error {
	// temp
	actorEd.componentTreeRoot = buildComponentTree(actor)

	if imgui.Button("Add Component") {
		actorEd.addDialog.Open()
	}
	if actorEd.addDialog.Draw() {
		desc := actorEd.addDialog.assetToCreate
		if desc != nil {
			comp, _ := desc.Create()
			actor.SetComponents(append(actor.GetComponents(), comp.(flat.Component)))
			actorEd.componentTreeRoot = buildComponentTree(actor)
			actorEd.valueToEdit = reflect.ValueOf(comp.(flat.Component)).Elem()
		}
	}
	edgui.DrawTree(actorEd.componentTreeRoot, nil)
	return nil
}

type componentTreeNode struct {
	name     string
	children []edgui.TreeNode
}

var _ edgui.TreeNode = (*componentTreeNode)(nil)

func (c *componentTreeNode) Name() string               { return c.name }
func (c *componentTreeNode) Children() []edgui.TreeNode { return c.children }
func (c *componentTreeNode) Leaf() bool                 { return len(c.children) == 0 }
func (c *componentTreeNode) Expanded() bool             { return true }

func buildComponentTree(actor flat.Actor) *componentTreeNode {
	root := &componentTreeNode{name: "root"}
	toNode := map[any]*componentTreeNode{nil: root}

	for _, c := range actor.GetComponents() {
		flat.WalkComponents(c, func(target, parent flat.Component) {
			name, _ := asset.ObjectTypeName(target)
			node := &componentTreeNode{name: name}
			toNode[target] = node
			pnode := toNode[parent]
			pnode.children = append(pnode.children, node)
		})
	}

	return root
}

type AddComponentDialog struct {
	Id            string
	assetToCreate *asset.AssetDescriptor
}

func (a *AddComponentDialog) Open() {
	imgui.OpenPopup(a.Id)
}

func (a *AddComponentDialog) Draw() bool {
	open := true
	if imgui.BeginPopupModalV(a.Id, &open, imgui.WindowFlagsNone) {
		defer imgui.EndPopup()
		imgui.Text("Create Component")
		imgui.Separator()
		for _, desc := range asset.GetAssetDescriptors() {
			imgui.TreeNodeV(desc.Name, imgui.TreeNodeFlagsLeaf|imgui.TreeNodeFlagsNoTreePushOnOpen)
			if imgui.IsItemClicked() {
				a.assetToCreate = desc
				imgui.CloseCurrentPopup()
				return true
			}
		}
	}
	return false
}
