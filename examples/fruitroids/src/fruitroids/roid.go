package fruitroids

import (
	"github.com/bradbev/flatland/src/flat"
	"github.com/deeean/go-vector/vector3"
	"github.com/hajimehoshi/ebiten/v2"
)

type Roid struct {
	flat.ActorBase
	velocity vector3.Vector3
}

func (r *Roid) Tick(deltaseconds float64) {
	r.ActorBase.Tick(deltaseconds)
	r.updatePhysics(deltaseconds)
}

func (r *Roid) updatePhysics(deltaseconds float64) {
	// move along our velocity
	v := r.velocity.MulScalar(deltaseconds)
	r.Transform.Location = *r.Transform.Location.Add(v)
}

func (r *Roid) Draw(screen *ebiten.Image) {
	pos := r.Transform.Location
	b := screen.Bounds()
	v := &r.velocity
	if (pos.X < 0 && v.X < 0) || (pos.X > float64(b.Max.X) && v.X > 0) {
		v.X *= -1
	}
	if (pos.Y < 0 && v.Y < 0) || (pos.Y > float64(b.Max.Y) && v.Y > 0) {
		v.Y *= -1
	}

	r.ActorBase.Draw(screen)
}
