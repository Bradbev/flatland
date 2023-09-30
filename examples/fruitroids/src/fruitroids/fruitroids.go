package fruitroids

import (
	"fmt"

	"github.com/bradbev/flatland/src/asset"
	"github.com/bradbev/flatland/src/flat"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Fruitroids struct {
	w, h     int
	gameFlow *GameFlow

	score          int
	scoreComponent *flat.TextComponent
}

var mainGame *Fruitroids

func (f *Fruitroids) BeginPlay(flow *GameFlow) {
	mainGame = f
	f.gameFlow = flow
	flow.BeginPlay()
}

func (f *Fruitroids) NextWorld() {
	f.gameFlow.NextWorld()
}

func (f *Fruitroids) IncScore(amount int) {
	f.score += amount
	if f.scoreComponent != nil {
		f.scoreComponent.SetValues(f.score)
	}
}

func getMainGame() *Fruitroids {
	return mainGame
}

type WorldType int

const (
	WorldType_Pre WorldType = iota
	WorldType_Main
	WorldType_Post
)

type GameFlow struct {
	Worlds []*flat.World

	index       WorldType
	activeWorld *flat.World
}

func (g *GameFlow) BeginPlay() {
	g.index = -1
	g.NextWorld()
}

func (g *GameFlow) Draw(screen *ebiten.Image) {
	if g.activeWorld != nil {
		g.activeWorld.Draw(screen)
	}
}

func (g *GameFlow) Update() {
	if g.activeWorld != nil {
		g.activeWorld.Update()
		if g.index != WorldType_Main {
			if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
				g.NextWorld()
			}
		}
	}
}

func (g *GameFlow) EndMainPlay() {
	g.NextWorld()
}

func (g *GameFlow) NextWorld() {
	if g.activeWorld != nil {
		g.activeWorld.EndPlay()
	}

	g.index = (g.index + 1) % (WorldType_Post + 1)
	g.activeWorld = g.Worlds[g.index]
	ActiveWorld = g.activeWorld
	g.activeWorld.BeginPlay()

	if f := getMainGame(); f != nil {
		f.scoreComponent = nil
		if f.gameFlow.index == WorldType_Main {
			flat.ForEachActorByType[*flat.EmptyActor](g.activeWorld, func(actor *flat.EmptyActor) error {
				for _, textComponent := range flat.FindComponentsByType[*flat.TextComponent](actor) {
					if textComponent.Name == "Score" {
						f.scoreComponent = textComponent
						return flat.StopIterating
					}
				}
				return nil
			})
			f.score = 0
			f.IncScore(0)
		}
	}
}

var ActiveWorld *flat.World

func (g *Fruitroids) Draw(screen *ebiten.Image) {
	if g.gameFlow != nil {
		g.gameFlow.Draw(screen)
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %.3f\nFPS: %.2f\n", ebiten.ActualTPS(), ebiten.ActualFPS()), 11, 2)
	ebitenutil.DebugPrintAt(screen, "FRUITROIDS", 11, 30)
}

func (g *Fruitroids) Update() error {
	if g.gameFlow != nil {
		g.gameFlow.Update()
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
	asset.RegisterAsset(Roid{})
	asset.RegisterAsset(SpawnConfig{})
	asset.RegisterAsset(LevelSpawn{})
	asset.RegisterAsset(CircleCollisionComponent{})
	asset.RegisterAsset(PhysicsCollisionManager{})
	asset.RegisterAsset(Bullet{})
	asset.RegisterAsset(GameFlow{})
}
