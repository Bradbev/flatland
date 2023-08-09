package editor

import (
	"flatland/src/editor/edgui"

	"github.com/inkyblackness/imgui-go/v4"
)

type menuManager struct {
	menus []edgui.Menu
}

func (m *menuManager) AddMenu(menu edgui.Menu) {
	m.menus = append(m.menus, menu)
}

func (m *menuManager) Draw() {
	if imgui.BeginMainMenuBar() {
		defer imgui.EndMainMenuBar()
		for _, menu := range m.menus {
			menu.Draw()
		}
	}
}
