package editors

import (
	"image/color"
	"reflect"

	"github.com/bradbev/flatland/src/asset"
	"github.com/bradbev/flatland/src/editor"
	"github.com/bradbev/flatland/src/editor/edgui"
	"github.com/bradbev/flatland/src/flat"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
)

// shorter type aliases
type f64 = float64
type f32 = float32

func RegisterAllFlatEditors(edit *editor.ImguiEditor) {
	// EXAMPLE: You can add your own custom editors for any type you choose,
	// including primitive types.
	edit.AddType(new(flat.Image), imageEd)

	// Editors for interfaces are supported
	edit.AddType(new(flat.Actor), actorEd)
	// For the case or ActorBase, it will also match the interface, and we do not want that
	edit.AddType(new(flat.ActorBase), editor.StructEd)
}

type imageEdContext struct {
	lastPathTried asset.Path
	img           *flat.Image
}

func (i *imageEdContext) Dispose(context *editor.TypeEditContext) {
	context.Ed.DisposeImguiTexture(i.img)
}

type aliasImage flat.Image

// The editor will try to call the Name function on an asset
// if it exists
func (a *aliasImage) Name() string { return "Custom Image Editor" }

// imageEd is a simple custom editor for flat.Image
// custom editors should not be in the same package as the asset they
// edit, otherwise these functions could be pulled into the game binary
func imageEd(context *editor.TypeEditContext, value reflect.Value) error {
	// use this idiom to get a pointer to the underlying value
	image := value.Addr().Interface().(*flat.Image)

	// If we need to store temp info while editing, use GetContext
	c, firstTime := editor.GetContext[imageEdContext](context, value)
	if firstTime {
		c.lastPathTried = image.Path
		c.img = image
	}

	{
		// To edit the same underlying value using the standard struct editor
		// make a new type so this editor isn't opened
		a := (*aliasImage)(image)
		context.Edit(a)
	}

	if image.Path != c.lastPathTried {
		c.lastPathTried = image.Path
		img := flat.Image{Path: image.Path}
		img.PostLoad()
		if img.GetImage() != nil {
			*image = img
		}
		if image.Path == "" {
			image.Reset()
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
	id, img := context.Ed.GetImguiTexture(image, int(w), int(h))
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
