package flat

import (
	"bytes"
	"flatland/src/asset"
	"fmt"
	"image"
	_ "image/png"
	"log"

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
