package editors

import (
	"image/color"
	"reflect"

	"github.com/bradbev/flatland/src/editor"
	"github.com/inkyblackness/imgui-go/v4"
)

type colorRGBAEdContext struct {
	RGBA [4]float32
}

func colorRGBAEd(context *editor.TypeEditContext, value reflect.Value) error {
	col := value.Addr().Interface().(*color.RGBA)
	c, firstTime := editor.GetContext[colorRGBAEdContext](context, value)
	if firstTime {
		c.RGBA[0] = float32(col.R) / 0xFF
		c.RGBA[1] = float32(col.G) / 0xFF
		c.RGBA[2] = float32(col.B) / 0xFF
		c.RGBA[3] = float32(col.A) / 0xFF
	}

	imgui.Text(context.StructField().Name)
	imgui.SameLine()
	if imgui.ColorEdit4("", &c.RGBA) {
		col.R = uint8(0xFF * c.RGBA[0])
		col.G = uint8(0xFF * c.RGBA[1])
		col.B = uint8(0xFF * c.RGBA[2])
		col.A = uint8(0xFF * c.RGBA[3])
		context.SetChanged()
	}
	return nil
}
