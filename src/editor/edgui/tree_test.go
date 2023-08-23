package edgui

import (
	"fmt"
	"testing"
)

type componentTreeNode struct {
	name     string
	children []TreeNode
	selected bool
}

var _ TreeNode = (*componentTreeNode)(nil)

func (c *componentTreeNode) Name() string         { return c.name }
func (c *componentTreeNode) Children() []TreeNode { return c.children }
func (c *componentTreeNode) Leaf() bool           { return len(c.children) == 0 }
func (c *componentTreeNode) Expanded() bool       { return true }
func (c *componentTreeNode) Selected() bool       { return c.selected }

func TestTreeWalk(t *testing.T) {
	a := &componentTreeNode{name: "a", children: []TreeNode{}}
	b := &componentTreeNode{name: "b", children: []TreeNode{}}
	c := &componentTreeNode{name: "c", children: []TreeNode{}}

	s := []TreeNode{a, b, c}
	root := &componentTreeNode{children: s}

	WalkTree(root, nil, func(node TreeNode, context any) {
		fmt.Printf("node %v\n", node)
		for _, e := range node.(*componentTreeNode).children {
			fmt.Println(a == e)
		}
	})

	t.Fail()
}
