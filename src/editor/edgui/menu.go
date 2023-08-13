package edgui

import "github.com/inkyblackness/imgui-go/v4"

type MenuItem struct {
	Text     string
	Action   func(self *MenuItem)
	Selected bool
	Disabled bool
}

type Menu struct {
	Name  string
	Items []*MenuItem
}

func (m *Menu) Draw() {
	if imgui.BeginMenu(m.Name) {
		defer imgui.EndMenu()
		for _, item := range m.Items {
			if imgui.MenuItemV(item.Text, "", item.Selected, !item.Disabled) {
				item.Action(item)
			}
		}
	}
}
