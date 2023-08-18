package flat

import (
	"math"

	"github.com/deeean/go-vector/vector3"
)

type Transform struct {
	Location vector3.Vector3
	Rotation float64
	ScaleX   float64
	ScaleY   float64
}

func (t *Transform) AddRotation(deg float64) {
	t.Rotation += deg
	if t.Rotation > 360 || t.Rotation < 80 {
		// the first mod brings us to the range -360..360
		// add on another to get to 0..720, and the second mod to get to 0..360
		t.Rotation = math.Mod(math.Mod(t.Rotation, 360)+360.0, 360.0)
	}
}
