package fruitroids

import (
	"fmt"

	"github.com/bradbev/flatland/src/flat"
	"github.com/deeean/go-vector/vector3"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Ship struct {
	flat.ActorBase
	ticks        int
	RotationRate float64
	Acceleration float64
	MaxVelocity  float64

	velocity vector3.Vector3
}

func (s *Ship) BeginPlay() {
	s.ActorBase.BeginPlay()
	s.velocity = vector3.Vector3{}
}

func (s *Ship) Tick(deltaseconds float64) {
	s.ActorBase.Tick(deltaseconds)
	s.ticks++
	isDown := func(key ebiten.Key) bool { return inpututil.KeyPressDuration(key) > 0 }
	if isDown(ebiten.KeyArrowLeft) {
		s.Transform.AddRotation(-s.RotationRate * deltaseconds)
	}
	if isDown(ebiten.KeyArrowRight) {
		s.Transform.AddRotation(s.RotationRate * deltaseconds)
	}
	if isDown(ebiten.KeyArrowDown) {
		s.velocity = vector3.Vector3{}
	}
	if isDown(ebiten.KeyArrowUp) {
		g := ebiten.GeoM{}
		g.Rotate(flat.DegToRad(s.Transform.Rotation))
		// Vector pointing along negative Y lets us add to x/y each timestep
		x, y := g.Apply(0, -s.Acceleration)
		s.velocity.X += x
		s.velocity.Y += y
		mag := s.velocity.Magnitude()
		clamped := flat.Clamp(mag, 0, s.MaxVelocity)
		if mag != clamped && mag > 0 {
			s.velocity = *s.velocity.MulScalar(clamped / mag)
		}
	}
	if isDown(ebiten.KeySpace) {
		// fire
	}

	// move along our velocity
	v := s.velocity.MulScalar(deltaseconds)
	s.Transform.Add(*v)
}

func (s *Ship) Draw(screen *ebiten.Image) {
	s.ActorBase.Draw(screen)
	x, y := s.Transform.Location.X, s.Transform.Location.Y
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Ship at %.2f %v %v", s.velocity.Magnitude(), s.velocity, s.Transform.Rotation), int(x), int(y))
}
