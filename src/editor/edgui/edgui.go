// edgui provides some helper functions for the editor UI
// The functions here should follow imgui style.

package edgui

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/inkyblackness/imgui-go/v4"
)

func Text(format string, args ...any) {
	t := fmt.Sprintf(format, args...)
	imgui.Text(t)
}

func InputText(label string, text *string) bool {
	imgui.Text(label)
	imgui.SameLine()
	return imgui.InputText("", text)
}

func TreeNodeWithPop(label string, flags imgui.TreeNodeFlags, body func()) {
	if imgui.TreeNodeV(label, flags) {
		defer imgui.TreePop()
		body()
	}
}

func BeginTableWithEnd(label string, columns int, body func()) {
	imgui.BeginTable(label, columns)
	defer imgui.EndTable()
	body()
}

// WithID is used to prevent two inputs from being treated as the
// same input within imgui.  Essentially, you should use this whenever
// you need to edit a reflect.Value
func WithID(value reflect.Value, body func()) {
	addr := fmt.Sprintf("%x", value.UnsafeAddr())
	imgui.PushID(addr)
	defer imgui.PopID()
	body()
}

func WithIDPtr[T any](ptr *T, body func()) {
	addr := fmt.Sprintf("%x", ptr)
	imgui.PushID(addr)
	defer imgui.PopID()
	body()
}

type AutoComplete struct {
	items []string
	// lowerItems is the lowercase version of items
	lowerItems []string
}

func (a *AutoComplete) SetItems(items []string) {
	lower := make([]string, 0, len(items))
	for _, item := range items {
		lower = append(lower, strings.ToLower(item))
	}
	a.items = items
	a.lowerItems = lower
}

func (a *AutoComplete) InputText(label string, s *string, onActivated func() []string) {
	// see https://github.com/ocornut/imgui/issues/718#issuecomment-1249822993
	// for implementation reference
	var enterPressed bool
	WithIDPtr(s, func() {
		enterPressed = imgui.InputTextV("", s, imgui.InputTextFlagsEnterReturnsTrue, nil)
	})
	isActive := imgui.IsItemActive()
	isActivated := imgui.IsItemActivated()

	if isActivated {
		imgui.OpenPopup("##popup")
		a.SetItems(onActivated())
	}

	{
		lowerInput := strings.ToLower(*s)
		imgui.SetNextWindowPos(imgui.Vec2{X: imgui.ItemRectMin().X, Y: imgui.ItemRectMax().Y})
		ImGuiWindowFlags_ChildWindow := imgui.WindowFlags(1 << 24) // don't use (internal)
		flags := imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoMove | imgui.WindowFlagsNoResize | ImGuiWindowFlags_ChildWindow
		if imgui.BeginPopupV("##popup", flags) {
			for _, item := range a.lowerItems {
				if strings.Contains(item, lowerInput) {
					if imgui.Selectable(item) {
						imgui.ClearActiveID()
						*s = item
					}
				}
			}
			if enterPressed || (!isActive && !imgui.IsWindowFocused()) {
				imgui.CloseCurrentPopup()
			}
			imgui.EndPopup()
		}
	}
}
