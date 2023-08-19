package edtest

import (
	"fmt"

	"github.com/bradbev/flatland/src/editor/edgui"
	"github.com/inkyblackness/imgui-go/v4"
	"golang.org/x/exp/slices"
)

func TreeTest() {
	if imgui.BeginChild("TestTree") {
		edgui.DrawTree(&testTree, treeHandler)
	}
	imgui.EndChild()
}

func TreeTestInit() {
	edgui.WalkTree(&testTree, nil, func(node edgui.TreeNode, context any) {
		node.(*testTreeNode).ExpandedF = true
	})
}

var testTree = testTreeNode{
	name: "root",
	children: []edgui.TreeNode{
		&ttn{name: "child1"},
		&ttn{name: "child2"},
		&ttn{name: "NextItem", children: []edgui.TreeNode{
			&ttn{name: "child3"},
			&ttn{name: "child4"},
		}},
	},
}

var treeHandler = &treeHandlerStruct{root: &testTree}

type testTreeNode struct {
	name      string
	children  []edgui.TreeNode
	ExpandedF bool
	SelectedF bool
}
type ttn = testTreeNode

func (t *testTreeNode) Name() string               { return t.name }
func (t *testTreeNode) Children() []edgui.TreeNode { return t.children }
func (t *testTreeNode) Leaf() bool                 { return len(t.children) == 0 }
func (t *testTreeNode) Expanded() bool             { return t.ExpandedF }
func (t *testTreeNode) Selected() bool             { return t.SelectedF }

type treeHandlerStruct struct {
	root          *testTreeNode
	dragContext   edgui.TreeDragContext
	dragStartNode edgui.TreeNode
}

func (t *treeHandlerStruct) Clicked(node edgui.TreeNode) {
	fmt.Println("Clicked ", node.Name())
	tn := node.(*testTreeNode)
	tn.SelectedF = !tn.SelectedF
}

func (t *treeHandlerStruct) Context() *edgui.TreeDragContext {
	return &t.dragContext
}

func (t *treeHandlerStruct) DragSource(node edgui.TreeNode) {
	if t.dragStartNode == nil {
		edgui.WalkTree(t.root, nil, func(node edgui.TreeNode, context any) {
			node.(*testTreeNode).SelectedF = false
		})
	}
	t.dragStartNode = node
}

func (t *treeHandlerStruct) DragTarget(node edgui.TreeNode, dropFlag edgui.TreeDropFlag) {
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
		remove(t.root, t.dragStartNode)
		ttn := node.(*testTreeNode)
		ttn.children = append(ttn.children, t.dragStartNode)

	case edgui.DroppedAfterNode:
		indexOffset++
		fallthrough
	case edgui.DroppedBeforeNode:
		p := parent(t.root, node)
		if p != nil {
			remove(t.root, t.dragStartNode)
			index := slices.Index(p.children, node) + indexOffset
			p.children = slices.Insert(p.children, index, t.dragStartNode)
		}
	}
	t.dragStartNode = nil
}

func parent(root *testTreeNode, child edgui.TreeNode) (ret *testTreeNode) {
	edgui.WalkTree(root, nil, func(node edgui.TreeNode, context any) {
		if slices.Contains(node.(*testTreeNode).children, child) {
			ret = node.(*testTreeNode)
		}
	})
	return ret
}

func remove(root *testTreeNode, node edgui.TreeNode) {
	if p := parent(root, node); p != nil {
		if slices.Contains(p.children, node) {
			p.children = slices.DeleteFunc(p.children, func(n edgui.TreeNode) bool {
				return node == n
			})
		}
	}
}
