package editor

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
)

type pieManager struct {
	ed       *ImguiEditor
	pieGames []ebiten.Game
}

func (p *pieManager) StartGame(game ebiten.Game) {
	p.pieGames = append(p.pieGames, game)
}

func (p *pieManager) Update(deltaseconds float32) {
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
	}
}
