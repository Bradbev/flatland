package flat

import (
	"github.com/bradbev/flatland/src/asset"
	"github.com/hajimehoshi/ebiten/v2"
)

type World struct {
	ActorBase
	tickables        []Tickable
	drawables        []Drawable
	PersistentActors []Actor
}

func NewWorld() *World {
	w := &World{}
	w.reset()
	return w
}

func (w *World) reset() {
	w.tickables = nil
	w.drawables = nil
}

func (w *World) PostLoad() {
}

func (w *World) BeginPlay() {
	w.reset()
	for _, actor := range w.PersistentActors {
		instance, _ := asset.NewInstance(actor)
		w.AddToWorld(instance.(Actor))
	}
}

func (w *World) AddToWorld(actor Actor) {
	if tickable, ok := actor.(Tickable); ok {
		w.tickables = append(w.tickables, tickable)
	}
	if drawable, ok := actor.(Drawable); ok {
		w.drawables = append(w.drawables, drawable)
	}
	if playable, ok := actor.(Playable); ok {
		playable.BeginPlay()
	}
}

func (w *World) Tick(deltaseconds float64) {
	for _, tickable := range w.tickables {
		tickable.Tick(deltaseconds)
	}
}

func (w *World) Draw(screen *ebiten.Image) {
	for _, drawable := range w.drawables {
		drawable.Draw(screen)
	}
}
