package flat

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"log"

	"github.com/bradbev/flatland/src/asset"

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
	dimensions vector2.Vector2
	op         ebiten.DrawImageOptions
}

var _ = Component((*ImageComponent)(nil))

func (c *ImageComponent) String() string {
	if c.Image != nil && c.Image.Path != "" {
		return "Image - " + string(c.Image.Path)
	}
	return "Image"
}

func (c *ImageComponent) BeginPlay() {
	if c.Image == nil {
		return
	}
	if c.Image.GetImage() != nil {
		bounds := c.Image.GetImage().Bounds()
		x, y := bounds.Dx(), bounds.Dy()
		c.dimensions.Set(float64(x), float64(y))
	}
	c.op = ebiten.DrawImageOptions{
		Filter: ebiten.FilterLinear,
	}
	c.op.GeoM = ebiten.GeoM{}
	// images start with the centre of the image at 0,0
	c.op.GeoM.Translate(-c.dimensions.X/2.0, -c.dimensions.Y/2.0)
}

func (c *ImageComponent) Draw(screen *ebiten.Image) {
	op := c.op
	ApplyComponentTransforms(c, &op.GeoM)

	if c.Image != nil && c.Image.GetImage() != nil {
		screen.DrawImage(c.Image.GetImage(), &op)
	}
}
