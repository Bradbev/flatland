package fruitroids

import (
	"flatland/src/asset"
	"flatland/src/flat"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Fruitroids struct {
	w, h  int
	World *flat.World
}

func (g *Fruitroids) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %.3f\nFPS: %.2f\n", ebiten.ActualTPS(), ebiten.ActualFPS()), 11, 2)
	ebitenutil.DebugPrintAt(screen, "FRUITROIDS", 11, 30)
	if g.World != nil {
		g.World.Draw(screen)
	}
}

func (g *Fruitroids) Update() error {
	if g.World != nil {
		g.World.Tick(1 / 60.0)
	}
	return nil
}

func (g *Fruitroids) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.w = outsideWidth
	g.h = outsideHeight
	return g.w, g.h
}

func RegisterFruitroidTypes() {
	asset.RegisterAsset(Ship{})
}

type Ship struct {
	flat.ActorBase
	ticks int
}

func (s *Ship) Tick(deltaseconds float64) {
	s.ActorBase.Tick(deltaseconds)
	s.ticks++
}

func (s *Ship) Draw(screen *ebiten.Image) {
	s.ActorBase.Draw(screen)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Ship at %v has ticked %v", s.Transform.Location, s.ticks), 11, 50)
}
