package editors

import (
	"image/color"

	"github.com/bradbev/flatland/src/editor"
	"github.com/bradbev/flatland/src/flat"
	"golang.org/x/image/font"
)

func RegisterAllFlatEditors(ed *editor.ImguiEditor) {
	// EXAMPLE: You can add your own custom editors for any type you choose,
	// including primitive types.
	ed.AddType(new(flat.Image), imageEd)

	// Editors for interfaces are supported
	ed.AddType(new(flat.Actor), actorEd)
	// For the case or ActorBase, it will also match the interface, and we do not want that
	ed.AddType(new(flat.ActorBase), actorBaseEd)
	ed.AddType(new(flat.Transform), transformEd)
	ed.AddType(new(flat.Font), fontEd)
	ed.AddType(new(color.RGBA), colorRGBAEd)
	ed.AddType(new(flat.World), worldEd)

	ed.RegisterEnum(map[any]string{
		font.HintingNone:     "None",
		font.HintingVertical: "Vertical",
		font.HintingFull:     "Full",
	})
}
