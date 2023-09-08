package flat

import (
	"image/color"
	"strings"
	"text/template"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type TextComponent struct {
	ComponentBase
	Font                  *Font
	Color                 color.RGBA
	Name                  string
	TextTemplate          string
	IgnoreParentRotations bool

	tmpl     *template.Template
	lastEval strings.Builder
	op       ebiten.DrawImageOptions
}

func (t *TextComponent) PostLoad() {
	tmpl, err := template.New(t.Name).Parse(t.TextTemplate)
	Check(err)
	t.lastEval.Reset()
	t.tmpl = tmpl
	t.tmpl.Execute(&t.lastEval, nil)
	t.op = ebiten.DrawImageOptions{}
	t.op.ColorScale.ScaleWithColor(t.Color)
}

func (t *TextComponent) FillTemplate(data any) {
	t.lastEval.Reset()
	t.tmpl.Execute(&t.lastEval, data)
}

func (t *TextComponent) Draw(screen *ebiten.Image) {
	if t.Font == nil {
		return
	}

	t.op.GeoM = ebiten.GeoM{}
	if t.IgnoreParentRotations {
		WalkUpComponentOwners(t, func(comp Component) {
			transform := comp.GetTransform()
			t.op.GeoM.Scale(transform.ScaleX, transform.ScaleY)
			t.op.GeoM.Translate(transform.Location.X, transform.Location.Y)
		})
	} else {
		ApplyComponentTransforms(t, &t.op.GeoM)
	}
	text.DrawWithOptions(screen, t.lastEval.String(), t.Font.face, &t.op)
}
