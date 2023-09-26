// edgui provides some helper functions for the editor UI
// The functions here should follow imgui style.

package edgui

import (
	"fmt"
	"reflect"

	"github.com/inkyblackness/imgui-go/v4"
)

func DragFloat64(label string, v *float64) bool {
	f32 := float32(*v)
	ret := false
	imgui.Text(label)
	imgui.SameLine()
	WithIDPtr(v, func() {
		imgui.PushItemWidth(50)
		ret = imgui.DragFloat("", &f32)
		imgui.PopItemWidth()
	})
	if ret {
		*v = float64(f32)
	}
	return ret
}

func Text(format string, args ...any) {
	t := fmt.Sprintf(format, args...)
	imgui.Text(t)
}

func InputText(label string, text *string) bool {
	imgui.Text(label)
	imgui.SameLine()

	return imgui.InputText("##"+label, text)
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

func WithItemWidth(width float32, body func()) {
	imgui.PushItemWidth(width)
	defer imgui.PopItemWidth()
	body()
}

func DragFloat2(
	label string,
	label1 string, v1 *float64,
	label2 string, v2 *float64) (ret bool) {

	imgui.Text(label)
	imgui.SameLine()
	ret = DragFloat64(label1, v1) || ret
	imgui.SameLine()
	ret = DragFloat64(label2, v2) || ret
	return ret
}

func DragFloat3(
	label string,
	label1 string, v1 *float64,
	label2 string, v2 *float64,
	label3 string, v3 *float64) (ret bool) {

	imgui.Text(label)
	imgui.SameLine()
	ret = DragFloat64(label1, v1) || ret
	imgui.SameLine()
	ret = DragFloat64(label2, v2) || ret
	imgui.SameLine()
	ret = DragFloat64(label3, v3) || ret
	return ret
}
