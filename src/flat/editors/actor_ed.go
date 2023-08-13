package editors

import (
	"flatland/src/editor"
	"flatland/src/flat"
	"image/color"
	"reflect"

	"github.com/inkyblackness/imgui-go/v4"
)

func actorEd(context *editor.TypeEditContext, value reflect.Value) error {
	//actor := value.Addr().Interface().(*flat.Image)
	flags := imgui.TableFlagsSizingStretchSame |
		imgui.TableFlagsResizable |
		imgui.TableFlagsBordersOuter |
		imgui.TableFlagsBordersV |
		imgui.TableFlagsContextMenuInBody
	if imgui.BeginTableV(context.ID("ActorEd##"), 3, flags, imgui.Vec2{}, 0) {
		defer imgui.EndTable()
		imgui.TableNextRow()

		imgui.TableSetColumnIndex(0)
		imgui.Text("Component View")
		imgui.Text("<TODO - tree view of components here>")

		imgui.TableSetColumnIndex(1)
		editor.StructEd(context, value)

		imgui.TableSetColumnIndex(2)
		renderActor(context, value)
	}
	return nil
}

func renderActor(context *editor.TypeEditContext, value reflect.Value) error {
	imgui.Text("Rendered View")
	w := imgui.ColumnWidth()
	id, img := context.Ed.GetImguiTexture(value, int(w), int(w))
	img.Fill(color.Black)
	_ = id
	if tickable, ok := value.Addr().Interface().(flat.Tickable); ok {
		tickable.Tick(1 / 60.0)
	}
	if drawable, ok := value.Addr().Interface().(flat.Drawable); ok {
		drawable.Draw(img)
	}

	_ = id
	// tell imgui to draw the ebiten.Image
	imgui.Image(id, imgui.Vec2{X: f32(w), Y: f32(w)})
	return nil
}
