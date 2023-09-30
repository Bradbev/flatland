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

	lastTextTemplate string
	tmpl             *template.Template
	lastEval         strings.Builder
	op               ebiten.DrawImageOptions
}

func (t *TextComponent) updateCachedValues() {
	if t.lastTextTemplate == t.TextTemplate {
		return
	}
	t.lastEval.Reset()
	tmpl, err := template.New(t.Name).Parse(t.TextTemplate)
	if err != nil {
		// malformed templates return errors, which we will display instead of the real text
		t.lastEval.WriteString("Error:" + err.Error())
		return
	}
	t.tmpl = tmpl
	t.tmpl.Execute(&t.lastEval, nil)
	t.op = ebiten.DrawImageOptions{}
	t.op.ColorScale.ScaleWithColor(t.Color)
	t.lastTextTemplate = t.TextTemplate
}

func (t *TextComponent) SetValues(data any) {
	t.updateCachedValues()
	t.lastEval.Reset()
	t.tmpl.Execute(&t.lastEval, data)
}

func (t *TextComponent) Draw(screen *ebiten.Image) {
	if t.Font == nil {
		return
	}
	t.updateCachedValues()

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
