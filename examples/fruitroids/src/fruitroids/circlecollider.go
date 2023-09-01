package fruitroids

import (
	"image/color"

	"github.com/bradbev/flatland/src/flat"
	"github.com/deeean/go-vector/vector2"
	"github.com/deeean/go-vector/vector3"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type CircleCollisionComponent struct {
	flat.ComponentBase

	ShowDebug      bool
	Radius         float64
	CollisionGroup string
}

type f32 = float32

func (c *CircleCollisionComponent) BeginPlay() {
	if gPhysicsManager != nil {
		gPhysicsManager.Add(c)
	}
}

func (c *CircleCollisionComponent) Draw(screen *ebiten.Image) {
	if !c.ShowDebug {
		return
	}
	geom := ebiten.GeoM{}
	flat.ApplyComponentTransforms(c, &geom)

	// radius is in screen space, this is probably all busted for scaled cameras
	p := vector2.New(geom.Apply(0, 0))
	antialias := true
	r := c.Radius
	vector.StrokeCircle(screen, f32(p.X), f32(p.Y), f32(r), 2, color.White, antialias)
}

func (c *CircleCollisionComponent) CheckOverlap(other *CircleCollisionComponent) {
	gC, gO := ebiten.GeoM{}, ebiten.GeoM{}
	flat.ApplyComponentTransforms(c, &gC)
	flat.ApplyComponentTransforms(other, &gO)

	vC := vector2.New(gC.Apply(0, 0))
	vO := vector2.New(gO.Apply(0, 0))
	dist := vC.Distance(vO)

	if dist < c.Radius+other.Radius {
		if r, ok := c.Owner().(*Ship); ok {
			r.velocity = vector3.Vector3{}
		}
		if r, ok := other.Owner().(*Roid); ok {
			if main, ok := c.Owner().(*Bullet); ok {
				ActiveWorld.World.RemoveFromWorld(r)
				gPhysicsManager.Remove(other)

				ActiveWorld.World.RemoveFromWorld(main)
				gPhysicsManager.Remove(c)
			}
		}
	}
}

var gPhysicsManager *PhysicsCollisionManager = nil

type BucketPair struct {
	Bucket1 string
	Bucket2 string
}

type PhysicsCollisionManager struct {
	flat.ActorBase

	// Only collide against pairs of buckets, not everything
	Buckets []BucketPair

	buckets map[string]map[*CircleCollisionComponent]struct{}
}

func (p *PhysicsCollisionManager) Remove(c *CircleCollisionComponent) {
	bucket := p.buckets[c.CollisionGroup]
	delete(bucket, c)
}

func (p *PhysicsCollisionManager) Add(c *CircleCollisionComponent) {
	bucket, ok := p.buckets[c.CollisionGroup]
	if !ok {
		bucket = map[*CircleCollisionComponent]struct{}{}
		p.buckets[c.CollisionGroup] = bucket
	}
	bucket[c] = struct{}{}
}

func (p *PhysicsCollisionManager) BeginPlay() {
	// this is really yuck, need a better solution than a global and relying on BeginPlay order
	gPhysicsManager = p
	p.buckets = map[string]map[*CircleCollisionComponent]struct{}{}
}

func (p *PhysicsCollisionManager) Update() {
	// this is clearly a terrible way to check collisions
	for _, pair := range p.Buckets {
		for a := range p.buckets[pair.Bucket1] {
			for b := range p.buckets[pair.Bucket2] {
				a.CheckOverlap(b)
			}
		}
	}
}
