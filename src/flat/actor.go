package flat

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Actor interface {
	Transformer
	Component
	Playable
	IsActor()
}

// Component
type Component interface {
	Transformer
	SetComponents([]Component)
	GetComponents() []Component
	SetOwner(owner Component)
	Owner() Component
}

type Transformer interface {
	GetTransform() *Transform
}

type Updateable interface {
	Update()
}

type Drawable interface {
	Draw(screen *ebiten.Image)
}

// TODO - Need some other lifecycle hooks.
// - Editor Construction?
// - EndPlay?

// Playable will be called right before a world is ready to run
// and start ticking
type Playable interface {
	BeginPlay()
}

// EditorPlayable will be called by the editor instead of Begin play
// in certain editor situations
type EditorPlayable interface {
	EditorBeginPlay()
}

type ComponentBase struct {
	Transform Transform
	owner     Component
	Children  []Component `flat:"inline"`
}

var _ Component = (*ComponentBase)(nil)

func (c *ComponentBase) SetOwner(owner Component)        { c.owner = owner }
func (c *ComponentBase) Owner() Component                { return c.owner }
func (c *ComponentBase) SetComponents(comps []Component) { c.Children = comps }
func (c *ComponentBase) GetComponents() []Component      { return c.Children }
func (c *ComponentBase) GetTransform() *Transform        { return &c.Transform }

type ActorBase struct {
	Transform            Transform
	Components           []Component `flat:"inline"`
	updateableComponents []Updateable
	drawableComponents   []Drawable
}

// EmptyActor can be used when you need an actor that is entirely defined
// by its components
type EmptyActor struct {
	ActorBase
}

func (a *EmptyActor) BeginPlay() {
	a.ActorBase.BeginPlay(a)
}

var _ Actor = (*EmptyActor)(nil)

// "static assert" that ActorBase implements Actor
// var _ Actor = (*ActorBase)(nil)
var _ Component = (*ActorBase)(nil)

func (a *ActorBase) reset() {
	a.updateableComponents = nil
	a.drawableComponents = nil
}
func (a *ActorBase) SetOwner(o Component) {
	if o != nil {
		panic("Cannot SetOwner on an Actor")
	}
}
func (a *ActorBase) Owner() Component                { return nil }
func (a *ActorBase) SetComponents(comps []Component) { a.Components = comps }
func (a *ActorBase) GetComponents() []Component      { return a.Components }

// BeginPlay must be called
func (a *ActorBase) BeginPlay(rootParent Actor) {
	a.reset()
	for _, component := range a.Components {
		if component == nil {
			continue
		}
		WalkComponents(component, func(component, parent Component) {
			if parent == nil {
				parent = rootParent
			}
			component.SetOwner(parent)
			if playable, ok := component.(Playable); ok {
				playable.BeginPlay()
			}
			if updateable, ok := component.(Updateable); ok {
				a.updateableComponents = append(a.updateableComponents, updateable)
			}
			if drawable, ok := component.(Drawable); ok {
				a.drawableComponents = append(a.drawableComponents, drawable)
			}
		})
	}
}

func (a *ActorBase) IsActor() {}

func (a *ActorBase) GetTransform() *Transform {
	return &a.Transform
}

func (a *ActorBase) Update() {
	for _, updateable := range a.updateableComponents {
		updateable.Update()
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
	if target == nil {
		return
	}
	callback(target, parent)
	for _, child := range target.GetComponents() {
		walkComponents(child, target, callback)
	}
}

func FindComponentsByType[T Component](parent Component) []T {
	result := []T{}
	WalkComponents(parent, func(target, _ Component) {
		if t, ok := target.(T); ok {
			result = append(result, t)
		}
	})
	return result
}
