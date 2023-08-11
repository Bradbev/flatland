package flat

import (
	"bytes"
	"flatland/src/asset"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math"

	"github.com/deeean/go-vector/vector2"
	"github.com/hajimehoshi/ebiten/v2"
)

type Image struct {
	Path asset.Path `filter:"png"`
	img  *ebiten.Image
}

func (i *Image) PostLoad() {
	fmt.Printf("Post load for Image %#v\n", i)
	content, err := asset.ReadFile(i.Path)
	if err != nil {
		log.Print(err)
		return
	}

	img, _, err := image.Decode(bytes.NewReader(content))
	if err != nil {
		log.Print(err)
		return
	}
	i.img = ebiten.NewImageFromImage(img)
	fmt.Printf("[Done] Post load for Image %#v\n", i)
}

func (i *Image) GetImage() *ebiten.Image {
	return i.img
}

func (i *Image) Reset() {
	i.Path = ""
	i.img = nil
}

type ImageComponent struct {
	ComponentBase
	Image      *Image
	image      *ebiten.Image
	dimensions vector2.Vector2
	op         ebiten.DrawImageOptions
	geoM       ebiten.GeoM
}

var _ = Component((*ImageComponent)(nil))

func (a *ImageComponent) SetImage(image *ebiten.Image) {
	a.image = image
	bounds := image.Bounds()
	x, y := bounds.Dx(), bounds.Dy()
	a.dimensions.Set(float64(x), float64(y))
	a.op = ebiten.DrawImageOptions{
		Filter: ebiten.FilterLinear,
	}
	a.geoM = ebiten.GeoM{}
	a.geoM.Translate(-a.dimensions.X/2.0, -a.dimensions.Y/2.0)
}

func (a *ImageComponent) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM = a.geoM
	t := a.Owner().GetTransform()
	op.GeoM.Scale(t.Scale, t.Scale)
	op.GeoM.Rotate(t.Rotation * math.Pi / 180.0)
	op.GeoM.Translate(t.Location.X, t.Location.Y)
	if a.Image != nil && a.Image.GetImage() != nil {
		screen.DrawImage(a.Image.GetImage(), op)
	}
}
