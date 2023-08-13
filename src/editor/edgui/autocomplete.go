package edgui

import (
	"strings"

	"github.com/inkyblackness/imgui-go/v4"
)

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

func (a *AutoComplete) InputText(label string, s *string, onActivated func() []string) bool {
	// see https://github.com/ocornut/imgui/issues/718#issuecomment-1249822993
	// for implementation reference
	var enterPressed bool
	result := false
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
						result = true
					}
				}
			}
			if enterPressed || (!isActive && !imgui.IsWindowFocused()) {
				imgui.CloseCurrentPopup()
			}
			imgui.EndPopup()
		}
	}
	return result || enterPressed
}
