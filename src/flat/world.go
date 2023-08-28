package flat

import (
	"github.com/bradbev/flatland/src/asset"
	"github.com/hajimehoshi/ebiten/v2"
)

type World struct {
	ActorBase
	updateables      []Updateable
	drawables        []Drawable
	PersistentActors []Actor
}

func NewWorld() *World {
	w := &World{}
	w.reset()
	return w
}

func (w *World) reset() {
	w.updateables = nil
	w.drawables = nil
}

func (w *World) PostLoad() {
}

func (w *World) BeginPlay() {
	w.reset()
	for _, actor := range w.PersistentActors {
		if actor == nil {
			continue
		}
		instance, _ := asset.NewInstance(actor)
		w.AddToWorld(instance.(Actor))
	}
}

func (w *World) AddToWorld(actor Actor) {
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
