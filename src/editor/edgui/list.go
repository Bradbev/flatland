package edgui

import (
	"strings"

	"github.com/inkyblackness/imgui-go/v4"
)

type ListNode interface {
	Name() string
}

type ListNodeSelected interface {
	Selected() bool
}

type ListNodeActionHandler interface {
	Clicked(node ListNode, index int)
	DoubleClicked(node ListNode, index int)
}

func DrawList[T ListNode](id string, list []T, handler ListNodeActionHandler) {
	DrawListWithFilter(id, list, nil, handler)
}

func DrawListWithFilter[T ListNode](id string, list []T, passFilter func(item T) bool, handler ListNodeActionHandler) {
	flags := imgui.TreeNodeFlagsLeaf

	for i, node := range list {
		nodeFlags := flags
		isSelected := false
		if sel, ok := any(node).(ListNodeSelected); ok && sel.Selected() {
			nodeFlags |= imgui.TreeNodeFlagsSelected
			isSelected = true
		}
		if !isSelected && passFilter != nil && !passFilter(node) {
			continue
		}

		if imgui.TreeNodeV(node.Name(), nodeFlags) {
			if handler != nil {
				if imgui.IsItemClicked() {
					handler.Clicked(node, i)
				}
				if imgui.IsItemHovered() && imgui.IsMouseDoubleClicked(0) {
					handler.DoubleClicked(node, i)
				}
			}
			imgui.TreePop()
		}
	}
}

type FilteredList[T ListNode] struct {
	List       []T
	filterText string
}

func (f *FilteredList[T]) Draw(label string, handler ListNodeActionHandler) {
	InputText(label, &f.filterText)
	filterText := strings.ToLower(f.filterText)
	filter := func(item T) bool {
		return strings.Contains(strings.ToLower(item.Name()), filterText)
	}
	imgui.BeginChild(label + "scrollbox")
	DrawListWithFilter(label, f.List, filter, handler)
	imgui.EndChild()
}
