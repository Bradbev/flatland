package editor

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"golang.org/x/exp/slices"
)

type pieManager struct {
	ed        *ImguiEditor
	pieGames  []ebiten.Game
	startGame func() ebiten.Game
}

func (p *pieManager) StartGameCallback(startGame func() ebiten.Game) {
	p.startGame = startGame
}

func (p *pieManager) StartGame() {
	game := p.startGame()
	p.pieGames = append(p.pieGames, game)
}

func (p *pieManager) Update(deltaseconds float32) {
	closedIndex := -1
	for i, game := range p.pieGames {
		open := true
		if imgui.BeginV(fmt.Sprintf("PIE %d", i), &open, 0) {
			game.Update()
			winSize := imgui.WindowSize()
			game.Layout(int(winSize.X), int(winSize.Y))
			id, img := p.ed.GetImguiTexture(game, int(winSize.X), int(winSize.Y))
			img.Fill(color.Black)
			game.Draw(img)
			imgui.Image(id, winSize)
		}
		imgui.End()
		if !open {
			closedIndex = i
		}
	}
	if closedIndex >= 0 {
		p.pieGames = slices.Delete(p.pieGames, closedIndex, closedIndex+1)
	}
}
