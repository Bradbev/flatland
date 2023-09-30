package flat

import (
	"errors"

	"github.com/bradbev/flatland/src/asset"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/exp/slices"
)

type World struct {
	actors           []Actor
	updateables      []Updateable
	drawables        []Drawable
	PersistentActors []Actor `flat:"inline"`
}

func NewWorld() *World {
	w := &World{}
	w.reset()
	return w
}

func FindActorsByType[T Actor](world *World) []T {
	var result []T
	for _, a := range world.actors {
		if t, ok := a.(T); ok {
			result = append(result, t)
		}
	}
	return result
}

var StopIterating = errors.New("stop iterating actors")

func ForEachActorByType[T Actor](world *World, callback func(actor T) error) {
	for _, a := range world.actors {
		if t, ok := a.(T); ok {
			if callback(t) == StopIterating {
				return
			}
		}
	}
}

func (w *World) reset() {
	w.actors = nil
	w.drawables = nil
	w.updateables = nil
}

func (w *World) PostLoad() {
}

func (w *World) BeginPlay() {
	w.beginPlay(false)
}

func (w *World) EditorBeginPlay() {
	w.beginPlay(true)
}

func (w *World) beginPlay(isEditor bool) {
	w.reset()
	for _, actor := range w.PersistentActors {
		if actor == nil {
			continue
		}
		if isEditor {
			w.AddToWorld(actor)
		} else {
			instance, _ := asset.NewInstance(actor)
			w.AddToWorld(instance.(Actor))
		}
	}
}

func (w *World) EndPlay() {
	w.reset()
}

func (w *World) AddToWorld(actor Actor) {
	w.actors = append(w.actors, actor)
	if updateable, ok := actor.(Updateable); ok {
		w.updateables = append(w.updateables, updateable)
	}
	if drawable, ok := actor.(Drawable); ok {
		w.drawables = append(w.drawables, drawable)
	}
	if playable, ok := actor.(Playable); ok {
		playable.BeginPlay()
	}
}

func (w *World) RemoveFromWorld(actor Actor) {
	w.actors = slices.DeleteFunc(w.actors, func(a Actor) bool {
		return a == actor
	})
	if updateable, ok := actor.(Updateable); ok {
		w.updateables = slices.DeleteFunc(w.updateables, func(u Updateable) bool {
			return u == updateable
		})
	}
	if drawable, ok := actor.(Drawable); ok {
		w.drawables = slices.DeleteFunc(w.drawables, func(d Drawable) bool {
			return d == drawable
		})
	}
}

func (w *World) Update() {
	for _, updateable := range w.updateables {
		updateable.Update()
	}
}

func (w *World) Draw(screen *ebiten.Image) {
	for _, drawable := range w.drawables {
		drawable.Draw(screen)
	}
}
