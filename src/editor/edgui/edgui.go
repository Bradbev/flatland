// edgui provides some helper functions for the editor UI
// The functions here should follow imgui style.

package edgui

import (
	"fmt"
	"reflect"

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

func WithDisabled(applyDisable bool, body func()) {
	if applyDisable {
		imgui.PushStyleColor(imgui.StyleColorButton, imgui.CurrentStyle().Color(imgui.StyleColorTextDisabled))
		defer imgui.PopStyleColor()
	}
	body()
}
