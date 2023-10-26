package edgui

import (
	"reflect"
	"strings"

	"github.com/bradbev/flatland/src/asset"
	"github.com/bradbev/flatland/src/flat"
	"github.com/inkyblackness/imgui-go/v4"
	"golang.org/x/exp/slices"
)

type selectListItem struct {
	assetPath string
	selected  bool
}

func (i *selectListItem) Name() string {
	return i.assetPath
}

func (i *selectListItem) Selected() bool {
	return i.selected
}

type SelectAssetModal struct {
	Title            string
	Type             reflect.Type
	open             bool
	selectedItem     *selectListItem
	list             FilteredList[*selectListItem]
	wasDoubleClicked bool
}

func (s *SelectAssetModal) SelectedPath() string {
	if s.selectedItem == nil {
		return ""
	}
	return s.selectedItem.assetPath
}

func (s *SelectAssetModal) Clicked(node ListNode, index int) {
	assetItem := node.(*selectListItem)
	if assetItem.selected {
		assetItem.selected = false
		s.selectedItem = nil
		return
	}
	for _, item := range s.list.List {
		item.selected = false
	}
	assetItem.selected = true
	s.selectedItem = assetItem
}
func (s *SelectAssetModal) DoubleClicked(node ListNode, index int) {
	assetItem := node.(*selectListItem)
	assetItem.selected = true
	s.selectedItem = assetItem
	s.wasDoubleClicked = true
}

func (s *SelectAssetModal) Open() {
	s.open = true
	imgui.OpenPopup(s.Title)

	var items []*selectListItem
	paths, _ := asset.FilterFilesByReflectType(s.Type)
	for _, p := range paths {
		items = append(items, &selectListItem{
			assetPath: p,
		})
	}
	slices.SortFunc(items, func(a, b *selectListItem) int {
		return strings.Compare(a.Name(), b.Name())
	})
	s.list = FilteredList[*selectListItem]{List: items}
}

func (s *SelectAssetModal) Draw() bool {
	return s.DrawWithExtraHeaderUI(nil)
}

func (s *SelectAssetModal) DrawWithExtraHeaderUI(headerHook func()) bool {
	okPressed := false
	if imgui.BeginPopupModalV(s.Title, &s.open, imgui.WindowFlagsNone) {
		defer imgui.EndPopup()

		if s.wasDoubleClicked {
			s.wasDoubleClicked = false
			imgui.CloseCurrentPopup()
			return true
		}

		WithDisabled(s.selectedItem == nil, func() {
			if imgui.Button("OK") {
				okPressed = true
				imgui.CloseCurrentPopup()
			}
		})
		imgui.SameLine()
		if imgui.Button("Cancel") {
			imgui.CloseCurrentPopup()
		}
		if headerHook != nil {
			headerHook()
		}

		imgui.Separator()
		s.list.Draw(s.Title, s)
	}
	return okPressed
}

type SelectParentModal struct {
	selectModal SelectAssetModal
	target      asset.Asset
	targetPath  asset.Path
}

func NewSelectParentModel(targetPath asset.Path, target asset.Asset) *SelectParentModal {
	return &SelectParentModal{
		selectModal: SelectAssetModal{
			Title: "Select Parent",
			Type:  reflect.TypeOf(target),
		},
		target:     target,
		targetPath: targetPath,
	}
}

func (s *SelectParentModal) Open() {
	s.selectModal.Open()
}

func (s *SelectParentModal) Draw() bool {
	if s.selectModal.DrawWithExtraHeaderUI(func() {
		imgui.SameLine()
		if imgui.Button("Set No Parent") {
			asset.SetParent(s.target, nil)
			imgui.CloseCurrentPopup()
		}
	}) {
		newParentPath := s.selectModal.SelectedPath()
		currentParentPath := asset.GetParent(s.target)
		if string(s.targetPath) != newParentPath && newParentPath != string(currentParentPath) {
			newParent, err := asset.Load(asset.Path(newParentPath))
			flat.Check(err)
			asset.SetParent(s.target, newParent)
			return true
		}
	}
	return false
}
