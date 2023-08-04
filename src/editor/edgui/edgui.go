// edgui provides some helper functions for the editor UI
// The functions here should follow imgui style.

package edgui

import (
	"fmt"

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
