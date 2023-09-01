package fruitroids

import (
	"github.com/bradbev/flatland/src/flat"
	"github.com/deeean/go-vector/vector3"
	"github.com/hajimehoshi/ebiten/v2"
)

type Bullet struct {
	flat.ActorBase
	Velocity     float64
	FireDelay    float64
	RotationRate float64

	velocity vector3.Vector3
}

func (b *Bullet) BeginPlay() {
	b.ActorBase.BeginPlay(b)
	b.velocity = vector3.Vector3{}
}

func (b *Bullet) SetDirection(angle float64) {
	op := ebiten.GeoM{}
	op.Rotate(flat.DegToRad(angle))
	x, y := op.Apply(0, -b.Velocity)
	b.velocity.Set(x, y, 0)
}

func (b *Bullet) Update() {
	b.ActorBase.Update()
	deltaseconds := flat.FrameTime()
	v := b.velocity.MulScalar(deltaseconds)
	b.Transform.Location = *b.Transform.Location.Add(v)
	b.Transform.Rotation += b.RotationRate * deltaseconds
}

func (b *Bullet) Draw(screen *ebiten.Image) {
	pos := b.Transform.Location
	bounds := screen.Bounds()
	if pos.X < 0 || pos.Y < 0 || pos.Y > float64(bounds.Max.Y) || pos.X > float64(bounds.Max.X) {
		ActiveWorld.World.RemoveFromWorld(b)
	}

	b.ActorBase.Draw(screen)
}
