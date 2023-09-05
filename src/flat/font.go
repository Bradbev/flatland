package flat

import (
	"github.com/bradbev/flatland/src/asset"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type Font struct {
	Name    string
	TtfFile asset.Path `flat:"filter:ttf"`
	Options opentype.FaceOptions

	face font.Face
}

func (f *Font) Face() font.Face {
	return f.face
}

func (f *Font) DefaultInitialize() {
	f.Options.Size = 24
	f.Options.DPI = 72
	f.Options.Hinting = font.HintingFull
}

func (f *Font) PostLoad() {
	d, err := asset.ReadFile(f.TtfFile)
	if err != nil {
		return
	}
	tt, err := opentype.Parse(d)
	if err != nil {
		return
	}
	face, err := opentype.NewFace(tt, &f.Options)
	if err != nil {
		return
	}
	f.face = face
}
