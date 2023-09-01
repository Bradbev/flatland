package editors

import (
	"reflect"

	"github.com/bradbev/flatland/src/editor"
	"github.com/bradbev/flatland/src/editor/edgui"
	"github.com/bradbev/flatland/src/flat"
)

func transformEd(context *editor.TypeEditContext, value reflect.Value) error {
	t := value.Addr().Interface().(*flat.Transform)
	if edgui.DragFloat3("Location   ",
		"X", &t.Location.X,
		"Y", &t.Location.Y,
		"Z", &t.Location.Z) {
		context.SetChanged()
	}

	if edgui.DragFloat64("Rotation     ", &t.Rotation) {
		context.SetChanged()
	}

	if edgui.DragFloat2("Scale      ",
		"X", &t.ScaleX,
		"Y", &t.ScaleY) {
		context.SetChanged()
	}
	return nil
}
