package flat

import (
	"math"

	"github.com/deeean/go-vector/vector2"
	"github.com/deeean/go-vector/vector3"
	"github.com/hajimehoshi/ebiten/v2"
)

type Component interface {
	setOwner(owner *ActorBase)
}

type Actor interface {
}

type Tickable interface {
	Tick(deltaseconds float64)
}

type Drawable interface {
	Draw(screen *ebiten.Image)
}

type ComponentBase struct {
	Owner *ActorBase
}

func (c *ComponentBase) setOwner(owner *ActorBase) {
	c.Owner = owner
}

type Transform2D struct {
	Location vector3.Vector3
	Rotation float64
}

type ActorBase struct {
	Transform          Transform2D
	tickableComponents []Tickable
	drawableComponents []Drawable
	components         []Component
}

func (a *ActorBase) isActor() {}

func (a *ActorBase) AddComponent(component Component) {
	component.setOwner(a)
	a.components = append(a.components, component)
	if tickable, ok := component.(Tickable); ok {
		a.tickableComponents = append(a.tickableComponents, tickable)
	}
	if drawable, ok := component.(Drawable); ok {
		a.drawableComponents = append(a.drawableComponents, drawable)
	}
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

type ImageComponent struct {
	ComponentBase
	image      *ebiten.Image
	dimensions vector2.Vector2
	op         ebiten.DrawImageOptions
	geoM       ebiten.GeoM
}

func (a *ImageComponent) SetImage(image *ebiten.Image) {
	a.image = image
	bounds := image.Bounds()
	x, y := bounds.Dx(), bounds.Dy()
	a.dimensions.Set(float64(x), float64(y))
	a.op = ebiten.DrawImageOptions{
		Filter: ebiten.FilterLinear,
	}
	a.geoM = ebiten.GeoM{}
	a.geoM.Translate(-a.dimensions.X/2.0, -a.dimensions.Y/2.0)
}

func (a *ImageComponent) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM = a.geoM
	op.GeoM.Scale(0.25, 0.25)
	t := &a.Owner.Transform
	op.GeoM.Rotate(t.Rotation * math.Pi / 180.0)
	op.GeoM.Translate(t.Location.X, t.Location.Y)
	screen.DrawImage(a.image, op)
}
