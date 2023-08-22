package fruitroids

import (
	"fmt"

	"github.com/bradbev/flatland/src/flat"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Ship struct {
	flat.ActorBase
	ticks        int
	RotationRate float64
}

func (s *Ship) Tick(deltaseconds float64) {
	s.ActorBase.Tick(deltaseconds)
	s.ticks++
	isDown := func(key ebiten.Key) bool { return inpututil.KeyPressDuration(key) > 0 }
	if isDown(ebiten.KeyArrowLeft) {
		s.Transform.AddRotation(s.RotationRate * deltaseconds)
	}
	if isDown(ebiten.KeyArrowRight) {
		s.Transform.AddRotation(-s.RotationRate * deltaseconds)
	}
	if isDown(ebiten.KeySpace) {
		// fire
	}
}

func (s *Ship) Draw(screen *ebiten.Image) {
	s.ActorBase.Draw(screen)
	x, y := s.Transform.Location.X, s.Transform.Location.Y
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Ship at %v has ticked %v", s.Transform.Location, s.ticks), int(x), int(y))
	if len(s.Components) > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("C0x %v", s.Components[0].GetTransform().Location.X), int(x), int(y+20))
	}
}
