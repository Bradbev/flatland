package editor

import (
	"flatland/src/asset"
	"fmt"
	"path/filepath"

	"github.com/inkyblackness/imgui-go/v4"
)

type contentWindow struct {
	editor               *ImguiEditor
	OpenDirs             map[string]bool
	SelectedDir          string
	ContentItems         []string
	SelectedContentItems map[string]bool
}

func (c *contentWindow) Draw() error {
	defer imgui.End()
	if !imgui.Begin("Content Browser") {
		return nil
	}
	if imgui.Button("Add") {
		imgui.OpenPopup("AddAssetModal")
	}
	c.drawAddAssetModal()

	cache := c.editor.buildFileCache()
	avail := imgui.ContentRegionAvail()
	size := imgui.Vec2{X: avail.X * 0.5, Y: avail.Y}
	imgui.BeginChildV("TreeViewChild", size, false, imgui.WindowFlagsAlwaysAutoResize)
	c.drawDirectoryTreeFromCache(cache)
	imgui.EndChild()

	imgui.SameLine()
	imgui.BeginChildV("ContentChild", size, false, imgui.WindowFlagsAlwaysAutoResize)
	if imgui.BeginTable("content items", 2) {
		for index := 0; index < len(c.ContentItems); {
			imgui.TableNextRow()
			for col := 0; col < 2; col++ {
				imgui.TableSetColumnIndex(col)
				text := c.ContentItems[index]
				if imgui.SelectableV(
					text,
					c.SelectedContentItems[text],
					imgui.SelectableFlagsAllowDoubleClick, imgui.Vec2{}) {

					// if ctrl down
					//c.SelectedContentItems[text] = !c.SelectedContentItems[text]
					if imgui.IsMouseDoubleClicked(0) {
						c.editor.EditAsset(filepath.Join(c.SelectedDir, text))
					}
				}
				index++
			}
		}
		imgui.EndTable()
	}
	imgui.EndChild()

	return nil
}

func (c *contentWindow) drawDirectoryTreeFromCache(node *fswalk) {
	basepath := filepath.Base(node.path)
	if basepath == "." {
		basepath = "content"
	}
	flags := imgui.TreeNodeFlagsOpenOnArrow | imgui.TreeNodeFlagsOpenOnDoubleClick
	if c.OpenDirs[node.path] {
		flags |= imgui.TreeNodeFlagsDefaultOpen
	}
	if len(node.dirs) == 0 {
		flags |= imgui.TreeNodeFlagsLeaf
	}
	if node.path == c.SelectedDir {
		flags |= imgui.TreeNodeFlagsSelected
	}
	c.OpenDirs[node.path] = false

	isOpened := imgui.TreeNodeV(basepath, flags)
	if imgui.IsItemClicked() {
		c.SelectedDir = node.path
		c.ContentItems = node.files
		c.SelectedContentItems = map[string]bool{}
	}
	if isOpened {
		c.OpenDirs[node.path] = true
		for _, n := range node.dirs {
			c.drawDirectoryTreeFromCache(n)
		}
		imgui.TreePop()
	}
}

func (c *contentWindow) drawAddAssetModal() {
	open := true
	if imgui.BeginPopupModalV("AddAssetModal", &open, imgui.WindowFlagsAlwaysAutoResize) {
		defer imgui.EndPopup()

		for _, a := range asset.ListAssets() {
			imgui.TreeNodeV(a.Name,
				imgui.TreeNodeFlagsLeaf|imgui.TreeNodeFlagsNoTreePushOnOpen)
			if imgui.IsItemClicked() {
				obj, err := a.Create()
				_ = err
				err = asset.Save("test", obj)
				if err != nil {
					fmt.Printf("error %v\n", err)
					//c.editor.raiseError(err)
				}
			}
		}
	}
}
