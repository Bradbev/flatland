package flat

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Actor interface {
	Transformer
	Component
	IsActor()
}

// Component
type Component interface {
	Transformer
	SetComponents([]Component)
	GetComponents() []Component
	SetOwner(owner Component)
	GetOwner() Component
}

type Transformer interface {
	GetTransform() Transform
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
	Transform Transform
	Owner     Component
	Children  []Component
}

var _ Component = (*ComponentBase)(nil)

func (c *ComponentBase) SetOwner(owner Component)        { c.Owner = owner }
func (c *ComponentBase) GetOwner() Component             { return c.Owner }
func (c *ComponentBase) SetComponents(comps []Component) { c.Children = comps }
func (c *ComponentBase) GetComponents() []Component      { return c.Children }
func (c *ComponentBase) GetTransform() Transform         { return c.Transform }

type ActorBase struct {
	Transform          Transform
	Components         []Component `flat:"inline"`
	tickableComponents []Tickable
	drawableComponents []Drawable
}

// "static assert" that ActorBase implements Actor
var _ Actor = (*ActorBase)(nil)
var _ Component = (*ActorBase)(nil)

func (a *ActorBase) reset() {
	a.tickableComponents = nil
	a.drawableComponents = nil
}
func (a *ActorBase) SetOwner(Component)              { panic("Cannot SetOwner on an Actor") }
func (a *ActorBase) GetOwner() Component             { return nil }
func (a *ActorBase) SetComponents(comps []Component) { a.Components = comps }
func (a *ActorBase) GetComponents() []Component      { return a.Components }

func (a *ActorBase) BeginPlay() {
	a.reset()
	for _, component := range a.Components {
		if component == nil {
			continue
		}
		WalkComponents(component, func(component, parent Component) {
			component.SetOwner(a)
			if playable, ok := component.(Playable); ok {
				playable.BeginPlay()
			}
			if tickable, ok := component.(Tickable); ok {
				a.tickableComponents = append(a.tickableComponents, tickable)
			}
			if drawable, ok := component.(Drawable); ok {
				a.drawableComponents = append(a.drawableComponents, drawable)
			}
		})
	}
}

func (a *ActorBase) IsActor() {}

func (a *ActorBase) GetTransform() Transform {
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

func WalkComponents(c Component, callback func(target, parent Component)) {
	walkComponents(c, nil, callback)
}

func walkComponents(target, parent Component, callback func(target, parent Component)) {
	callback(target, parent)
	for _, child := range target.GetComponents() {
		walkComponents(child, target, callback)
	}
}
