package flat

import (
	"reflect"

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
type Playable interface {
	BeginPlay()
}

type ComponentBase struct {
	Transform Transform
	owner     Component
	Children  []Component
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
	beginPlayWasCalled   bool
}

// "static assert" that ActorBase implements Actor
var _ Actor = (*ActorBase)(nil)
var _ Component = (*ActorBase)(nil)

func (a *ActorBase) reset() {
	a.updateableComponents = nil
	a.drawableComponents = nil
	a.beginPlayWasCalled = false
}
func (a *ActorBase) SetOwner(o Component) {
	if o != nil {
		panic("Cannot SetOwner on an Actor")
	}
}
func (a *ActorBase) Owner() Component                { return nil }
func (a *ActorBase) SetComponents(comps []Component) { a.Components = comps }
func (a *ActorBase) GetComponents() []Component      { return a.Components }

func (a *ActorBase) BeginPlay(rootParent Actor) {
	a.reset()
	a.beginPlayWasCalled = true
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
	Assert(a.beginPlayWasCalled, "BeginPlay(actor) must be called before update (usually in your own BeginPlay) if you wish to use ActorBase")
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
	callback(target, parent)
	for _, child := range target.GetComponents() {
		walkComponents(child, target, callback)
	}
}

func FindComponents[T Component](parent Component) []T {
	result := []T{}
	var zeroT T
	typ := reflect.TypeOf(zeroT)
	WalkComponents(parent, func(target, _ Component) {
		if reflect.TypeOf(target) == typ {
			result = append(result, target.(T))
		}
	})
	return result
}
