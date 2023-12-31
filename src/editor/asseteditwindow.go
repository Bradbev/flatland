package editor

import (
	"reflect"

	"github.com/bradbev/flatland/src/asset"
	"github.com/bradbev/flatland/src/editor/edgui"
	"github.com/bradbev/flatland/src/flat"

	"github.com/inkyblackness/imgui-go/v4"
)

type assetEditWindow struct {
	target      asset.Asset
	path        string
	context     *TypeEditContext
	selectModal edgui.SelectAssetModal
}

func callEditorBeginPlay(a any) {
	if editorPlayable, ok := a.(flat.EditorPlayable); ok {
		editorPlayable.EditorBeginPlay()
	} else if playable, ok := a.(flat.Playable); ok {
		playable.BeginPlay()
	}
}

func newAssetEditWindow(path string, target asset.Asset, context *TypeEditContext) *assetEditWindow {
	// TODO - this needs a different lifecycle hook
	callEditorBeginPlay(target)
	return &assetEditWindow{
		path:    path,
		target:  target,
		context: context,
		selectModal: edgui.SelectAssetModal{
			Title: "Select Parent",
			Type:  reflect.TypeOf(target),
		},
	}
}

func (a *assetEditWindow) Draw() error {
	defer imgui.End()
	open := true
	if imgui.BeginV(a.path, &open, 0) {
		enabled := a.context.hasChanged
		reload := false
		edgui.WithDisabled(!enabled, func() {
			if imgui.Button("Save") && enabled {
				asset.Save(asset.Path(a.path), a.target)
				a.context.hasChanged = false
				reload = true
			}
			imgui.SameLine()
			if imgui.Button("Revert") {
				asset.LoadWithOptions(asset.Path(a.path), asset.LoadOptions{ForceReload: true})
				callEditorBeginPlay(a.target)
			}
		})
		imgui.SameLine()
		if imgui.Button("Refresh") || reload {
			if postloader, ok := a.target.(asset.PostLoadingAsset); ok {
				postloader.PostLoad()
			}
			if comp, ok := a.target.(flat.Component); ok {
				flat.WalkComponents(comp, func(target, _ flat.Component) {
					if postloader, ok := target.(asset.PostLoadingAsset); ok {
						postloader.PostLoad()
					}
				})
			}
			callEditorBeginPlay(a.target)
		}
		imgui.SameLine()
		if imgui.Button("Set Parent") {
			a.selectModal.Open()
		}
		if a.selectModal.DrawWithExtraHeaderUI(func() {
			imgui.SameLine()
			if imgui.Button("Set No Parent") {
				asset.SetParent(a.target, nil)
				imgui.CloseCurrentPopup()
			}
		}) {
			newParentPath := a.selectModal.SelectedPath()
			currentParentPath := asset.GetParent(a.target)
			if a.path != newParentPath && newParentPath != string(currentParentPath) {
				newParent, err := asset.Load(asset.Path(newParentPath))
				flat.Check(err)
				asset.SetParent(a.target, newParent)
				a.context.SetChanged()
			}
		}

		if parent := string(asset.GetParent(a.target)); parent != "" {
			imgui.SameLine()
			imgui.Text("Parent: " + parent)
		}
		imgui.Separator()

		value := reflect.ValueOf(a.target)
		a.context.EditValue(value)
	}
	if !open {
		// If there is context from the editor, delete it
		DisposeContext(a.context, reflect.ValueOf(a.target))
		return closeDrawable
	}
	return nil
}
