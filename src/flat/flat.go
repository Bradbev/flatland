package flat

import "github.com/hajimehoshi/ebiten/v2"

type World struct {
	tickables        []Tickable
	drawables        []Drawable
	PersistentActors []Actor
}

func NewWorld() *World {
	return &World{
		tickables: []Tickable{},
		drawables: []Drawable{},
	}
}

func (w *World) PostLoad() {
	for _, a := range w.PersistentActors {
		w.AddActor(a)
	}
}

func (w *World) AddActor(actor Actor) {
	if tickable, ok := actor.(Tickable); ok {
		w.tickables = append(w.tickables, tickable)
	}
	if drawable, ok := actor.(Drawable); ok {
		w.drawables = append(w.drawables, drawable)
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
