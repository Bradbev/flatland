package fruitroids

import (
	"flatland/src/flat"
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Ship struct {
	flat.ActorBase
	ticks int
	Image *flat.Image
}

func (s *Ship) Tick(deltaseconds float64) {
	s.ActorBase.Tick(deltaseconds)
	s.ticks++
}

func (s *Ship) Draw(screen *ebiten.Image) {
	s.ActorBase.Draw(screen)
	x, y := s.Transform.Location.X, s.Transform.Location.Y
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Ship at %v has ticked %v", s.Transform.Location, s.ticks), int(x), int(y))
	if s.Image != nil {
		t := s.Transform
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(t.Scale, t.Scale)
		op.GeoM.Rotate(t.Rotation * math.Pi / 180.0)
		op.GeoM.Translate(t.Location.X, t.Location.Y)
		screen.DrawImage(s.Image.GetImage(), op)
	}
}
