package editors

import (
	"image/color"
	"reflect"

	"github.com/bradbev/flatland/src/editor"
	"github.com/bradbev/flatland/src/flat"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/inkyblackness/imgui-go/v4"
	"golang.org/x/image/font/opentype"
)

type fontEdContext struct {
	x, y        int32
	lastOptions opentype.FaceOptions
	lastPath    string
	testText    string
}

type aliasFont flat.Font

func fontEd(context *editor.TypeEditContext, value reflect.Value) error {
	// use this idiom to get a pointer to the underlying value
	fnt := value.Addr().Interface().(*flat.Font)

	// If we need to store temp info while editing, use GetContext
	c, firstTime := editor.GetContext[fontEdContext](context, value)
	if firstTime {
		c.x = 50
		c.y = 50
	}

	{
		// To edit the same underlying value using the standard struct editor
		// make a new type so this editor isn't opened
		f := (*aliasFont)(fnt)
		context.Edit(f)
	}

	imgui.DragInt("X", &c.x)
	imgui.DragInt("Y", &c.y)
	imgui.InputText("Text", &c.testText)

	if c.lastOptions != fnt.Options || c.lastPath != string(fnt.TtfFile) {
		c.lastOptions = fnt.Options
		c.lastPath = string(fnt.TtfFile)
		fnt.PostLoad()
	}

	// Custom editor below here
	winSize := imgui.WindowSize()

	if fnt.Face() == nil {
		return nil
	}
	w := winSize.X - 50

	// Use GetImguiTexture/imgui.Image(id) to put ebitengine Images
	// into an imgui context
	id, img := context.Ed.GetImguiTexture(fnt, int(w), int(w))
	{
		img.Fill(color.White)
		text.Draw(img, "ABCDEFGHIJKLMNOPQRSTUVWXYZ", fnt.Face(), int(c.x), int(c.y), color.Black)
		text.Draw(img, "abcdefghijklmnopqrstuvwxyz", fnt.Face(), int(c.x), int(c.y+1*int32(fnt.Options.Size+5)), color.Black)
		text.Draw(img, "0123456789!@#$%^&*()_+-/;:", fnt.Face(), int(c.x), int(c.y+2*int32(fnt.Options.Size+5)), color.Black)
		text.Draw(img, ",.<> /?'\"[]{}\\|`~", fnt.Face(), int(c.x), int(c.y+3*int32(fnt.Options.Size+5)), color.Black)
		text.Draw(img, c.testText, fnt.Face(), int(c.x), int(c.y+5*int32(fnt.Options.Size+5)), color.Black)
	}

	// tell imgui to draw the ebiten.Image
	imgui.Image(id, imgui.Vec2{X: w, Y: f32(w)})
	return nil
}
