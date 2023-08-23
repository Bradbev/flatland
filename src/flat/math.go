package flat

import (
	"math"

	"github.com/deeean/go-vector/vector3"
	"golang.org/x/exp/constraints"
)

type Transform struct {
	Location vector3.Vector3
	Rotation float64
	ScaleX   float64
	ScaleY   float64
}

func (t *Transform) Add(v vector3.Vector3) {
	t.Location = *t.Location.Add(&v)
}

func (t *Transform) AddRotation(deg float64) {
	t.Rotation += deg
	if t.Rotation > 360 || t.Rotation < 80 {
		// the first mod brings us to the range -360..360
		// add on another to get to 0..720, and the second mod to get to 0..360
		t.Rotation = math.Mod(math.Mod(t.Rotation, 360)+360.0, 360.0)
	}
}

func DegToRad(deg float64) float64 {
	return deg * math.Pi / 180
}

func Clamp[T constraints.Float | constraints.Integer](current, min, max T) T {
	if current < min {
		return min
	}
	if current > max {
		return max
	}
	return current
}
