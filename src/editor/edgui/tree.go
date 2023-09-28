package edgui

import (
	"strings"

	"github.com/inkyblackness/imgui-go/v4"
)

type TreeNode interface {
	Name() string
	Children() []TreeNode
	Leaf() bool
	Expanded() bool
}

// TreeNodeSelected is an extended interface from the basic TreeNode interface
// Nodes must implement this interface if they wish to be selectable
type TreeNodeSelected interface {
	Selected() bool
}

type TreeNodeActionHandler interface {
	Clicked(node TreeNode)
}

type TreeDropFlag int

const (
	NotDropped TreeDropFlag = iota
	DroppedOnNode
	DroppedBeforeNode
	DroppedAfterNode
)

// TreeNodeDragDropHandler is an extended interface that can be used by clients
// that wish to respond to drag/drop events
type TreeNodeDragDropHandler interface {
	// Context must be stored by the client
	Context() *TreeDragContext
	DragSource(node TreeNode)
	DragTarget(node TreeNode, dropFlag TreeDropFlag)
}

type TreeDragContext struct {
	dragStartNode  TreeNode
	lastTargetNode TreeNode
	lastMousePos   imgui.Vec2
	ydelta         float32
}

type treeContext struct {
	dragging bool
}

func WalkTree(root TreeNode, context any, callback func(node TreeNode, context any)) {
	callback(root, context)
	for _, n := range root.Children() {
		WalkTree(n, context, callback)
	}
}

func TreeNodeIsDescendantOf(parent TreeNode, target TreeNode) bool {
	if len(parent.Children()) == 0 {
		return false
	}
	for _, c := range parent.Children() {
		if c == target {
			return true
		}
		if TreeNodeIsDescendantOf(c, target) {
			return true
		}
	}
	return false
}

func DrawTree(root TreeNode, handler TreeNodeActionHandler) {
	context := &treeContext{}
	drawTree(root, handler, context)

	if dragger, ok := handler.(TreeNodeDragDropHandler); ok {
		if context.dragging == false && dragger.Context().dragStartNode != nil {
			treeDropped(dragger, nil, true)
		}
	}
}

func drawTree(root TreeNode, handler TreeNodeActionHandler, context *treeContext) {
	if imgui.TreeNodeV(root.Name(), treeFlags(root)) {
		if handler != nil {
			treeClicked(root, handler)
			treeDragged(root, handler, context)
		}

		for _, child := range root.Children() {
			drawTree(child, handler, context)
		}
		imgui.TreePop()
	}
}

func treeFlags(node TreeNode) imgui.TreeNodeFlags {
	nodeFlags := imgui.TreeNodeFlagsOpenOnArrow
	if node.Leaf() {
		nodeFlags |= imgui.TreeNodeFlagsLeaf
	}
	if node.Expanded() {
		nodeFlags |= imgui.TreeNodeFlagsDefaultOpen
	}
	if sel, ok := node.(TreeNodeSelected); ok && sel.Selected() {
		nodeFlags |= imgui.TreeNodeFlagsSelected
	}

	return nodeFlags
}

func treeClicked(node TreeNode, handler TreeNodeActionHandler) {
	if imgui.IsItemClicked() {
		handler.Clicked(node)
	}
}

func treeDragged(node TreeNode, handler TreeNodeActionHandler, context *treeContext) {
	if dragger, ok := handler.(TreeNodeDragDropHandler); ok {
		if imgui.BeginDragDropSource(imgui.DragDropFlagsNone) {
			context.dragging = true
			imgui.SetDragDropPayload("_TREENODE", []byte("fake"), imgui.ConditionNone)

			dragger.Context().dragStartNode = node
			dragger.DragSource(node)
			nameParts := strings.Split(node.Name(), "##")
			imgui.Text(nameParts[0])
			imgui.EndDragDropSource()
		}
		if imgui.BeginDragDropTarget() {
			data := imgui.AcceptDragDropPayload("_TREENODE", imgui.DragDropFlagsNone)
			treeDropped(dragger, node, len(data) != 0)
			imgui.EndDragDropTarget()
		}
	}
}

func treeDropped(dragger TreeNodeDragDropHandler, node TreeNode, dropped bool) {
	context := dragger.Context()
	m := imgui.MousePos()
	mouseDelta := m.Minus(context.lastMousePos)
	if mouseDelta.Y != 0 {
		context.ydelta = mouseDelta.Y
	}
	context.lastMousePos = m
	if node != nil {
		context.lastTargetNode = node
	}

	flag := NotDropped
	if dropped {
		flag = DroppedOnNode
		if node == nil {
			flag = DroppedBeforeNode
			if context.ydelta > 0 {
				flag = DroppedAfterNode
			}
		}
	}

	dragger.DragTarget(context.lastTargetNode, flag)

	// clear the context
	if dropped {
		*context = TreeDragContext{}
	}
}
