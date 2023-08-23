package editors

import (
	"fmt"
	"image/color"
	"reflect"

	"github.com/bradbev/flatland/src/asset"
	"github.com/bradbev/flatland/src/editor"
	"github.com/bradbev/flatland/src/editor/edgui"
	"github.com/bradbev/flatland/src/flat"
	"golang.org/x/exp/slices"

	"github.com/inkyblackness/imgui-go/v4"
)

type actorEdContext struct {
	componentTreeRoot *componentTreeNode
	handler           componentTreeNodeHandler
	addDialog         *AddComponentDialog
	valueToEdit       reflect.Value
}

func actorEd(context *editor.TypeEditContext, value reflect.Value) error {
	actor := value.Addr().Interface().(flat.Actor)
	c, firstTime := editor.GetContext[actorEdContext](context, value)
	if firstTime {
		c.componentTreeRoot = buildComponentTree(actor)
		c.handler = componentTreeNodeHandler{
			root:    c.componentTreeRoot,
			context: c,
			actor:   actor,
		}
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
	if tickable, ok := value.Addr().Interface().(flat.Tickable); ok {
		tickable.Tick(1 / 60.0)
	}
	if drawable, ok := value.Addr().Interface().(flat.Drawable); ok {
		drawable.Draw(img)
	}
	// tell imgui to draw the ebiten.Image
	imgui.Image(id, imgui.Vec2{X: f32(w), Y: f32(w)})
	return nil
}

func componentTree(actor flat.Actor, actorEd *actorEdContext, context *editor.TypeEditContext, value reflect.Value) error {
	if imgui.Button("Add Component") {
		actorEd.addDialog.Open()
	}
	imgui.SameLine()
	if imgui.Button("Delete") {
		edgui.WalkTree(actorEd.componentTreeRoot, nil, func(node edgui.TreeNode, context any) {
			if node.(*componentTreeNode).selected {
				removeNode(actorEd.componentTreeRoot, node)
			}
		})
		copyTree(actorEd.handler.actor, actorEd.componentTreeRoot)

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
	edgui.DrawTree(actorEd.componentTreeRoot, &actorEd.handler)
	return nil
}

type componentTreeNode struct {
	name      string
	children  []edgui.TreeNode
	selected  bool
	component flat.Component
}

var _ edgui.TreeNode = (*componentTreeNode)(nil)

func (c *componentTreeNode) Name() string {
	name, _ := asset.ObjectTypeName(c.component)
	if s, ok := c.component.(fmt.Stringer); ok {
		name = s.String()
	}
	name = fmt.Sprintf("%s##%x", name, c.component)
	return name
}
func (c *componentTreeNode) Children() []edgui.TreeNode { return c.children }
func (c *componentTreeNode) Leaf() bool                 { return len(c.children) == 0 }
func (c *componentTreeNode) Expanded() bool             { return true }
func (c *componentTreeNode) Selected() bool             { return c.selected }

type componentTreeNodeHandler struct {
	context       *actorEdContext
	root          *componentTreeNode
	dragContext   edgui.TreeDragContext
	dragStartNode edgui.TreeNode
	actor         flat.Actor
}

func (t *componentTreeNodeHandler) Clicked(node edgui.TreeNode) {
	n := node.(*componentTreeNode)
	edgui.WalkTree(t.root, nil, func(node edgui.TreeNode, context any) {
		node.(*componentTreeNode).selected = false
	})
	n.selected = !n.selected
	t.context.valueToEdit = reflect.ValueOf(n.component).Elem()
}

func (t *componentTreeNodeHandler) Context() *edgui.TreeDragContext {
	return &t.dragContext
}

func (t *componentTreeNodeHandler) DragSource(node edgui.TreeNode) {
	if t.dragStartNode == nil {
		edgui.WalkTree(t.root, nil, func(node edgui.TreeNode, context any) {
			node.(*componentTreeNode).selected = false
		})
	}
	t.dragStartNode = node
}

func (t *componentTreeNodeHandler) DragTarget(node edgui.TreeNode, dropFlag edgui.TreeDropFlag) {
	if dropFlag == edgui.NotDropped || t.dragStartNode == t.root {
		return
	}
	// can't reparent to a child
	if edgui.TreeNodeIsDescendantOf(t.dragStartNode, node) {
		return
	}

	indexOffset := 0
	switch dropFlag {
	case edgui.DroppedOnNode:
		if removeNode(t.root, t.dragStartNode) {
			n := node.(*componentTreeNode)
			n.children = append(n.children, t.dragStartNode)
		}

	case edgui.DroppedAfterNode:
		indexOffset++
		fallthrough
	case edgui.DroppedBeforeNode:
		p := findParentOfNode(t.root, node)
		if p != nil {
			if removeNode(t.root, t.dragStartNode) {
				index := slices.Index(p.children, node) + indexOffset
				p.children = slices.Insert(p.children, index, t.dragStartNode)
			}
		}
	}

	copyTree(t.actor, t.root)

	t.dragStartNode = nil
}

func copyTree(dest flat.Component, src edgui.TreeNode) {
	dstChildren := make([]flat.Component, len(src.Children()))
	dest.SetComponents(dstChildren)
	for i := 0; i < len(dstChildren); i++ {
		dstChildren[i] = src.Children()[i].(*componentTreeNode).component
		copyTree(dstChildren[i], src.Children()[i])
	}
}

func findParentOfNode(root *componentTreeNode, child edgui.TreeNode) (ret *componentTreeNode) {
	edgui.WalkTree(root, nil, func(node edgui.TreeNode, context any) {
		if slices.Contains(node.(*componentTreeNode).children, child) {
			ret = node.(*componentTreeNode)
		}
	})
	return ret
}

func removeNode(root *componentTreeNode, node edgui.TreeNode) bool {
	if p := findParentOfNode(root, node); p != nil {
		if slices.Contains(p.children, node) {
			p.children = slices.DeleteFunc(p.children, func(n edgui.TreeNode) bool {
				return node == n
			})
			return true
		}
	}
	return false
}

func buildComponentTree(actor flat.Actor) *componentTreeNode {
	root := &componentTreeNode{
		name:      "root",
		component: actor,
	}
	toNode := map[any]*componentTreeNode{nil: root}

	for _, c := range actor.GetComponents() {
		flat.WalkComponents(c, func(target, parent flat.Component) {
			name, _ := asset.ObjectTypeName(target)
			if s, ok := target.(fmt.Stringer); ok {
				name = s.String()
			}
			name = fmt.Sprintf("%s_%x", name, target)
			node := &componentTreeNode{
				name:      name,
				component: target,
			}
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
