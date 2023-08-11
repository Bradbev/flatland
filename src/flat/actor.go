package flat

import (
	"github.com/deeean/go-vector/vector3"
	"github.com/hajimehoshi/ebiten/v2"
)

type Component interface {
	SetOwner(owner Actor)
	Owner() Actor
}

type Actor interface {
	GetTransform() Transform2D
}

type Tickable interface {
	Tick(deltaseconds float64)
}

type Drawable interface {
	Draw(screen *ebiten.Image)
}

type Playable interface {
	BeginPlay()
}

type ComponentBase struct {
	owner Actor
}

func (c *ComponentBase) SetOwner(owner Actor) {
	c.owner = owner
}
func (c *ComponentBase) Owner() Actor {
	return c.owner
}

type Transform2D struct {
	Location vector3.Vector3
	Rotation float64
	Scale    float64
}

type ActorBase struct {
	Transform          Transform2D
	Components         []Component
	tickableComponents []Tickable
	drawableComponents []Drawable
}

// "static assert" that ActorBase implements Actor
var _ = Actor((*ActorBase)(nil))

func (a *ActorBase) reset() {
	a.tickableComponents = nil
	a.drawableComponents = nil
}

func (a *ActorBase) BeginPlay() {
	a.reset()
	for _, component := range a.Components {
		component.SetOwner(a)
		if tickable, ok := component.(Tickable); ok {
			a.tickableComponents = append(a.tickableComponents, tickable)
		}
		if drawable, ok := component.(Drawable); ok {
			a.drawableComponents = append(a.drawableComponents, drawable)
		}
	}
}

func (a *ActorBase) GetTransform() Transform2D {
	return a.Transform
}

func (a *ActorBase) Tick(deltaseconds float64) {
	for _, tickable := range a.tickableComponents {
		tickable.Tick(deltaseconds)
	}
}

func (a *ActorBase) Draw(screen *ebiten.Image) {
	for _, drawable := range a.drawableComponents {
		drawable.Draw(screen)
	}
}

/*
func FindComponent[T Component](owner Actor) T {
	actor := owner.(*ActorBase)
	target := reflect.TypeOf(owner)
	for _, component := range actor.components {
		if reflect.TypeOf(component) == target {
			return component.(T)
		}
	}
	return *new(T) // noooo
}
*/
