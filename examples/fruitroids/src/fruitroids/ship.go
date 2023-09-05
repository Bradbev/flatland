package fruitroids

import (
	"fmt"
	"time"

	"github.com/bradbev/flatland/src/asset"
	"github.com/bradbev/flatland/src/flat"
	"github.com/deeean/go-vector/vector3"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// TOUR:Fruitroids 4
// The Ship struct is our player controlled object.
// It embeds the flat.ActorBase struct.  This mean two things
// 1) The flat.Actor interface is satisfied by Ship
// 2) Ship.BeginPlay *must* call ActorBase.BeginPlay(actor)
// Actors can have sub components added to them.  The ship3.json
// asset has an Image component and a CircleCollisionComponent, those
// components take care of showing an image and colliding into things.
// The exported fields of this struct will be editable in the FruitEditor.
type Ship struct {
	flat.ActorBase
	RotationRate float64
	Acceleration float64
	MaxVelocity  float64
	BulletType   *Bullet

	velocity     vector3.Vector3
	lastFireTime time.Time
}

func (s *Ship) BeginPlay() {
	s.Transform.Location = vector3.Vector3{}
	s.ActorBase.BeginPlay(s)
	s.velocity = vector3.Vector3{}

	//texts := flat.FindComponents[flat.Text](s)
}

func (s *Ship) Update() {
	s.ActorBase.Update()
	s.handleInput()
	s.updatePhysics()
}

func (s *Ship) handleInput() {
	deltaseconds := flat.FrameTime()
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
	if isDown(ebiten.KeySpace) && ActiveWorld != nil {
		if time.Since(s.lastFireTime) < time.Duration(float64(time.Second)*s.BulletType.FireDelay) {
			return
		}
		s.lastFireTime = time.Now()
		// fire
		b, err := asset.NewInstance(s.BulletType)
		if err != nil {
			fmt.Println("terrible")
		}
		bullet := b.(*Bullet)
		bullet.Transform = s.Transform
		ActiveWorld.World.AddToWorld(bullet)
		bullet.SetDirection(s.Transform.Rotation)
	}
}

func (s *Ship) updatePhysics() {
	// move along our velocity
	deltaseconds := flat.FrameTime()
	v := s.velocity.MulScalar(deltaseconds)
	s.Transform.Location = *s.Transform.Location.Add(v)
}

func (s *Ship) Draw(screen *ebiten.Image) {
	pos := s.Transform.Location
	b := screen.Bounds()
	v := &s.velocity
	if (pos.X < 0 && v.X < 0) || (pos.X > float64(b.Max.X) && v.X > 0) {
		v.X *= -1
	}
	if (pos.Y < 0 && v.Y < 0) || (pos.Y > float64(b.Max.Y) && v.Y > 0) {
		v.Y *= -1
	}

	s.ActorBase.Draw(screen)
	//x, y := s.Transform.Location.X, s.Transform.Location.Y
	//	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Ship at %.2f %v %v", s.velocity.Magnitude(), s.velocity, s.Transform.Rotation), int(x), int(y))
}
