package editor

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bradbev/flatland/src/asset"
	"github.com/bradbev/flatland/src/editor/edgui"

	"github.com/inkyblackness/imgui-go/v4"
)

// contentWindows shows a tree browser of the filesystem and
// a display pane for the current directory
// The model here is Unreal's Content Browser
type contentWindow struct {
	editor               *ImguiEditor
	OpenDirs             map[string]bool
	SelectedDir          string
	ContentItems         []string
	SelectedContentItems map[string]bool

	// used when creating an asset
	newAssetName  string
	assetToCreate *asset.AssetDescriptor
}

func newContentWindow(editor *ImguiEditor) *contentWindow {
	ret := &contentWindow{
		OpenDirs: map[string]bool{},
		editor:   editor,
	}
	cache := editor.buildFileCache()
	ret.selectNode(cache)
	return ret
}

const addModalTitle = "AddAssetModal"

func (c *contentWindow) Draw() error {
	defer imgui.End()
	if !imgui.Begin("Content Browser") {
		return nil
	}
	if imgui.Button("Add") {
		imgui.OpenPopup(addModalTitle)
		c.newAssetName = "default"
	}
	c.drawAddAssetModal()

	cache := c.editor.buildFileCache()
	avail := imgui.ContentRegionAvail()
	size := imgui.Vec2{X: avail.X * 0.5, Y: avail.Y}

	{ // left
		imgui.BeginChildV("TreeViewChild", size, false, imgui.WindowFlagsAlwaysAutoResize)
		c.drawDirectoryTreeFromCache(cache)
		imgui.EndChild()
	}

	imgui.SameLine()
	{ // right
		imgui.BeginChildV("ContentChild", size, false, imgui.WindowFlagsAlwaysAutoResize)
		if imgui.BeginTable("content items", 2) {
			for index := 0; index < len(c.ContentItems); {
				imgui.TableNextRow()
				for col := 0; col < 2 && index < len(c.ContentItems); col++ {
					imgui.TableSetColumnIndex(col)
					text, _ := strings.CutPrefix(c.ContentItems[index], c.SelectedDir+"/")
					if imgui.SelectableV(
						text,
						c.SelectedContentItems[text],
						imgui.SelectableFlagsAllowDoubleClick, imgui.Vec2{}) {

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
	}

	return nil
}

func (c *contentWindow) selectNode(node *fswalk) {
	c.SelectedDir = node.path
	c.ContentItems = node.files
	c.SelectedContentItems = map[string]bool{}
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
		c.selectNode(node)
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
	if imgui.BeginPopupModalV(addModalTitle, &open, imgui.WindowFlagsNone) {
		defer imgui.EndPopup()

		edgui.Text("content/%s", c.SelectedDir)
		edgui.InputText("AssetName", &c.newAssetName)
		invalidName := strings.Contains(c.newAssetName, ".")
		if invalidName {
			edgui.Text("Names may not contain '.'")
		} else if c.assetToCreate == nil {
			edgui.Text("Select Asset Type")
		} else {
			edgui.Text("Asset Type: %s", c.assetToCreate.Name)
			edgui.Text("Path: content/%s.json", filepath.Join(c.SelectedDir, c.newAssetName))
			if imgui.Button("Create") {
				c.createNewAsset(c.assetToCreate)
				imgui.CloseCurrentPopup()
			}
		}

		imgui.Separator()
		func() { // asset box, in a func to defer the disabled style
			if invalidName {
				imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{0.5, 0.5, 0.5, 0.5})
				defer imgui.PopStyleColor()
			}
			imgui.BeginChildV("Assets", imgui.Vec2{}, true, 0)
			defer imgui.EndChild()
			//edgui.Text("Assets")
			for _, a := range asset.GetAssetDescriptors() {
				imgui.TreeNodeV(a.Name, imgui.TreeNodeFlagsLeaf|imgui.TreeNodeFlagsNoTreePushOnOpen)
				if imgui.IsItemClicked() {
					c.assetToCreate = a
				}
			}
		}()
	}
}

func (c *contentWindow) createNewAsset(a *asset.AssetDescriptor) {
	obj, err := a.Create()
	_ = err
	dest := filepath.Join(c.SelectedDir, c.newAssetName) + ".json"
	err = asset.Save(asset.Path(dest), obj)
	if err != nil {
		fmt.Printf("error %v\n", err)
		//c.editor.raiseError(err)
	}
}
