package main

import (
	"github.com/bradbev/flatland/src/editor"
	"github.com/bradbev/flatland/src/flat"
	"github.com/bradbev/flatland/src/flat/editors"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	wrapper := editor.NewEbitengineWrapper(1024, 768)

	// add the game specific types
	flat.RegisterAllFlatTypes()
	editors.RegisterAllFlatEditors(wrapper.Editor)

	ebiten.RunGame(wrapper)
}
