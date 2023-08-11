package editor

import (
	"flatland/src/asset"
	"flatland/src/editor/edgui"
	"flatland/src/flat"
	"reflect"

	"github.com/inkyblackness/imgui-go/v4"
)

type assetEditWindow struct {
	target  asset.Asset
	path    string
	context *TypeEditContext
}

func newAssetEditWindow(path string, target asset.Asset, context *TypeEditContext) *assetEditWindow {
	return &assetEditWindow{
		path:    path,
		target:  target,
		context: context,
	}
}

func (a *assetEditWindow) Draw() error {
	defer imgui.End()
	open := true
	if imgui.BeginV(a.path, &open, 0) {
		enabled := a.context.hasChanged
		edgui.WithDisabled(!enabled, func() {
			imgui.SameLineV(0, imgui.WindowWidth()-120)
			reload := false
			if imgui.Button("Save") && enabled {
				asset.Save(asset.Path(a.path), a.target)
				a.context.hasChanged = false
				reload = true
			}
			imgui.SameLine()
			if (imgui.Button("Refresh") || reload) && enabled {
				if postloader, ok := a.target.(asset.PostLoadingAsset); ok {
					postloader.PostLoad()
				}
				if playable, ok := a.target.(flat.Playable); ok {
					playable.BeginPlay()
				}
			}
		})
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
