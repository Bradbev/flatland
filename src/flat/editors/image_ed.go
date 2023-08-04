package editors

import (
	"flatland/src/editor"
	"flatland/src/editor/edgui"
	"flatland/src/flat"
	"image/color"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
)

// shorter type aliases
type f64 = float64
type f32 = float32

func RegisterAllFlatEditors(edit *editor.ImguiEditor) {
	edit.AddType(new(flat.Image), imageEd)
}

type imageEdContext struct {
	flt float32
}

// imageEd is a simple custom editor for flat.Image
func imageEd(ed *editor.ImguiEditor, value reflect.Value) error {
	// If we need to store temp info while editing, use GetContext
	c, firstTime := editor.GetContext[imageEdContext](ed, value)
	if firstTime {
		c.flt = 90
	}

	image := value.Addr().Interface().(*flat.Image)
	{
		// To edit the same underlying value using the standard struct editor
		// make a new type so this editor isn't opened
		type alias flat.Image
		a := (*alias)(image)
		ed.Edit(a)
	}

	// Custom editor below here
	winSize := imgui.WindowSize()
	edgui.Text("Window size (%d, %d) %v", int(winSize.X), int(winSize.Y), imgui.CursorPos())

	if image.GetImage() == nil {
		return nil
	}
	size := image.GetImage().Bounds().Size()
	edgui.Text("Image size (%d, %d)", int(size.X), int(size.Y))

	// use the context so that the c.flt value is preserved
	imgui.SliderFloat("Rotation", &c.flt, 0, 360)

	w := winSize.X
	scale := f64(w) / f64(size.X)
	h := f64(size.Y) * scale

	// Use GetImguiTexture/imgui.Image(id) to put ebitengine Images
	// into an imgui context
	id, img := ed.GetImguiTexture(image, int(w), int(h))

	{
		// scale and draw the image
		op := ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(-f64(w)/2, -h/2)
		op.GeoM.Rotate(flat.DegToRad(f64(c.flt)))
		op.GeoM.Translate(f64(w)/2, h/2)
		img.Fill(color.Black)
		img.DrawImage(image.GetImage(), &op)
	}

	// tell imgui to draw the ebiten.Image
	imgui.Image(id, imgui.Vec2{X: w, Y: f32(h)})
	return nil
}
