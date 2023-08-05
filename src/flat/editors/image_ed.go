package editors

import (
	"flatland/src/asset"
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
	lastPathTried asset.Path
	img           *flat.Image
}

func (i *imageEdContext) Dispose(ed *editor.ImguiEditor) {
	ed.DisposeImguiTexture(i.img)
}

type aliasImage flat.Image

// The editor will try to call the Name function on an asset
// if it exists
func (a *aliasImage) Name() string { return "Custom Image Editor" }

// imageEd is a simple custom editor for flat.Image
// HOW DO I CLEAN UP RESOURCES FOR THIS EDITOR?
func imageEd(ed *editor.ImguiEditor, value reflect.Value) error {
	// use this idiom to get a pointer to the underlying value
	image := value.Addr().Interface().(*flat.Image)

	// If we need to store temp info while editing, use GetContext
	c, firstTime := editor.GetContext[imageEdContext](ed, value)
	if firstTime {
		c.lastPathTried = image.Path
		c.img = image
	}

	{
		// To edit the same underlying value using the standard struct editor
		// make a new type so this editor isn't opened
		a := (*aliasImage)(image)
		ed.Edit(a)
	}

	if image.Path != c.lastPathTried {
		c.lastPathTried = image.Path
		img := flat.Image{Path: image.Path}
		img.PostLoad()
		if img.GetImage() != nil {
			*image = img
		}
	}

	// Custom editor below here
	winSize := imgui.WindowSize()
	edgui.Text("Window size (%d, %d) %v", int(winSize.X), int(winSize.Y), imgui.CursorPos())

	if image.GetImage() == nil {
		return nil
	}
	size := image.GetImage().Bounds().Size()

	w := winSize.X - 50
	scale := f64(w) / f64(size.X)
	h := f64(size.Y) * scale
	edgui.Text("Image size (%d, %d) scale %f", int(size.X), int(size.Y), scale)
	edgui.Text("Scaled size (%v, %v)", w, h)

	// Use GetImguiTexture/imgui.Image(id) to put ebitengine Images
	// into an imgui context
	id, img := ed.GetImguiTexture(image, int(w), int(h))
	{
		// scale and draw the image
		op := ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		img.Fill(color.Black)
		img.DrawImage(image.GetImage(), &op)
	}

	// tell imgui to draw the ebiten.Image
	imgui.Image(id, imgui.Vec2{X: w, Y: f32(h)})
	return nil
}
