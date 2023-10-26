package flat

import (
	"embed"
	"image"
)

//go:embed content
var Content embed.FS

type Bounder interface {
	Bounds() image.Rectangle
}
