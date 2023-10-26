package flat

import (
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
)

type MouseEventComponent struct {
	ComponentBase

	EventHandler MouseEvent `flat:"inline"`
}

func (m *MouseEventComponent) BeginPlay() {
	if m.EventHandler != nil {
		m.EventHandler.OnEnter(nil)
	}
}

type MouseEvent interface {
	OnEnter(owner Actor)
	OnExit(owner Actor)
	OnMouseButton(owner Actor, button ebiten.MouseButton)
}

type codeBlock struct {
	name        string
	description string
	fn          reflect.Value
}

type codeBlocks struct {
	blocks map[reflect.Type][]codeBlock
}

func CodeBlock[T any](name string, description string, factory func() *T) {

}

type TestMouseHandler struct {
	ToPrint string
}

func (t TestMouseHandler) OnEnter(owner Actor) {
	print("enter", t.ToPrint, "\n")

}
func (t TestMouseHandler) OnExit(owner Actor)                                   {}
func (t TestMouseHandler) OnMouseButton(owner Actor, button ebiten.MouseButton) {}
