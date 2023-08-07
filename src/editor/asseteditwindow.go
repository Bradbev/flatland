package editor

import (
	"flatland/src/asset"
	"reflect"

	"github.com/inkyblackness/imgui-go/v4"
)

type assetEditWindow struct {
	target  asset.Asset
	path    string
	context *TypeEditContext
}

func (a *assetEditWindow) Draw() error {
	defer imgui.End()
	open := true
	if imgui.BeginV(a.path, &open, 0) {
		imgui.SameLineV(0, imgui.WindowWidth()-80)
		if imgui.Button("Save") {
			asset.Save(asset.Path(a.path), a.target)
		}
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
