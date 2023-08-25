package fruitroids

import (
	"fmt"

	"github.com/bradbev/flatland/src/flat"
	"github.com/deeean/go-vector/vector3"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Roid struct {
	flat.ActorBase
	velocity      vector3.Vector3
	rotationDelta float64
}

func (r *Roid) Tick(deltaseconds float64) {
	r.ActorBase.Tick(deltaseconds)
	r.updatePhysics(deltaseconds)
}

func (r *Roid) updatePhysics(deltaseconds float64) {
	// move along our velocity
	v := r.velocity.MulScalar(deltaseconds)
	r.Transform.Location = *r.Transform.Location.Add(v)
	r.Transform.AddRotation(r.rotationDelta)
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
	x, y := r.Transform.Location.X, r.Transform.Location.Y
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Roid at %v", r.Transform), int(x), int(y))
}
