package editor

import (
	"flatland/src/asset"
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/inkyblackness/imgui-go/v4"
	"golang.org/x/exp/slices"
)

// This file contains the imgui specific workings of editor

type imguiEditorImpl struct{}

func (e *imguiEditorImpl) BeginStruct(name string) {
	imgui.BeginTable(name, 2)
}

func (e *imguiEditorImpl) EndStruct() {
	imgui.EndTable()
}

func (e *imguiEditorImpl) FieldName(name string) {
	imgui.TableNextRow()
	imgui.TableNextColumn()
	imgui.Text(name)
	imgui.TableNextColumn()
}

func ImguiTypeEditor() *CommonEditor {
	ed := NewCommonEditor(&imguiEditorImpl{})
	ed.AddType(new(float32), float32Edit)
	ed.AddType(new(float64), float64Edit)
	ed.AddType(new(bool), boolEdit)
	ed.AddType(new(string), stringEdit)
	ed.AddType(new(int), intEdit)
	return ed
}

func structEd(types *CommonEditor, value reflect.Value) error {
	t := value.Type()
	if t.Kind() != reflect.Struct {
		logger.Fatalf("Not a struct - %v", t.Kind())
	}
	types.impl.BeginStruct("Need to find struct name")
	for i := 0; i < t.NumField(); i++ {
		field := value.Field(i)
		structField := t.Field(i)
		if structField.IsExported() {
			types.impl.FieldName(structField.Name)
			types.EditValue(field.Addr())
		}
	}
	types.impl.EndStruct()

	return nil
}

func withID(value reflect.Value, body func()) {
	addr := fmt.Sprintf("%x", value.UnsafeAddr())
	imgui.PushID(addr)
	defer imgui.PopID()
	body()
}

func float32Edit(types *CommonEditor, value reflect.Value) error {
	withID(value, func() {
		addr := value.Addr().Interface().(*float32)
		imgui.DragFloat("", addr)
	})
	return nil
}

func float64Edit(types *CommonEditor, value reflect.Value) error {
	withID(value, func() {
		f32 := float32(value.Float())
		imgui.DragFloat("", &f32)
		value.SetFloat(float64(f32))
	})
	return nil
}

func boolEdit(types *CommonEditor, value reflect.Value) error {
	withID(value, func() {
		addr := value.Addr().Interface().(*bool)
		imgui.Checkbox("", addr)
	})
	return nil
}

func stringEdit(types *CommonEditor, value reflect.Value) error {
	withID(value, func() {
		addr := value.Addr().Interface().(*string)
		imgui.InputText("", addr)
	})
	return nil
}

func intEdit(types *CommonEditor, value reflect.Value) error {
	withID(value, func() {
		i32 := int32(value.Int())
		imgui.InputInt("", &i32)
		value.SetInt(int64(i32))
	})
	return nil
}

const errorModalID = "ErrorModal##unique"

type ImguiEditor struct {
	commonEditor    *CommonEditor
	contentPath     string
	fsys            fs.FS
	errorModalText  string
	content         *contentWindow
	assetsUnderEdit []*assetEditWindow
}

type contentWindow struct {
	ed                   *ImguiEditor
	OpenDirs             map[string]bool
	SelectedDir          string
	ContentItems         []string
	SelectedContentItems map[string]bool
}

type assetEditWindow struct {
	target asset.Asset
	path   string
}

func NewImguiEditor() *ImguiEditor {
	ret := &ImguiEditor{
		commonEditor: ImguiTypeEditor(),
		content: &contentWindow{
			OpenDirs: map[string]bool{},
		},
	}
	ret.content.ed = ret

	return ret
}

type fswalk struct {
	path  string
	dirs  []*fswalk
	files []string
}

func (e *ImguiEditor) Update(deltaseconds float32) error {
	e.content.draw()

	var toRemove *assetEditWindow
	for _, toEdit := range e.assetsUnderEdit {
		func(editing *assetEditWindow) {
			defer imgui.End()
			open := true
			if imgui.BeginV(editing.path, &open, 0) {
				e.commonEditor.Edit(editing.target)
			}
			if !open {
				toRemove = editing
			}
		}(toEdit)
	}
	e.assetsUnderEdit = slices.DeleteFunc(e.assetsUnderEdit, func(aew *assetEditWindow) bool {
		return toRemove == aew
	})
	return nil
}

func (e *ImguiEditor) EditAsset(path string) {
	// don't add already open windows
	if slices.ContainsFunc(e.assetsUnderEdit, func(underEdit *assetEditWindow) bool {
		return underEdit.path == path
	}) {
		return
	}

	loaded, err := asset.Load(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	e.assetsUnderEdit = append(e.assetsUnderEdit, &assetEditWindow{
		path:   path,
		target: loaded,
	})
}

func (c *contentWindow) draw() error {
	defer imgui.End()
	if !imgui.Begin("Content Browser") {
		return nil
	}
	if imgui.Button("Add") {
		imgui.OpenPopup("AddAssetModal")
	}
	c.drawAddAssetModal()

	cache := c.ed.buildFileCache()
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
						c.ed.EditAsset(filepath.Join(c.SelectedDir, text))
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

func (e *ImguiEditor) raiseError(err error) {
	e.errorModalText = err.Error()
	imgui.OpenPopup(errorModalID)
}

func (e *ImguiEditor) displayErrorModal() {
	open := true
	if imgui.BeginPopupModalV(errorModalID, &open, imgui.WindowFlagsAlwaysAutoResize) {
		defer imgui.EndPopup()
		imgui.Text(e.errorModalText)
		if imgui.Button("Dismiss") {
			e.errorModalText = ""
			imgui.CloseCurrentPopup()
		}
	}
}

func (e *ImguiEditor) buildFileCache() *fswalk {
	// todo - optimise this and cache state, filling it
	// as folders are opened
	stack := []*fswalk{{path: ""}}
	peek := func() *fswalk {
		return stack[len(stack)-1]
	}

	fs.WalkDir(e.commonEditor.fsysRead, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}
		if d.IsDir() {
			next := &fswalk{path: path}
			peek().dirs = append(peek().dirs, next)
			stack = append(stack, next)
			return nil
		}
		for {
			if strings.HasPrefix(path, peek().path) {
				break
			}
			stack = stack[:len(stack)-1]
		}
		peek().files = append(peek().files, path)

		return nil
	})
	return stack[0]
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
		defer c.ed.displayErrorModal()

		/*
			if imgui.Button("another") {
				imgui.OpenPopup("Test")
			e.raiseError(fmt.Errorf("test err"))
			}
			if imgui.BeginPopupModal("Test") {
				imgui.Text("inner modal")
				imgui.EndPopup()
			}
		*/

		for _, a := range asset.ListAssets() {
			imgui.TreeNodeV(a.Name,
				imgui.TreeNodeFlagsLeaf|imgui.TreeNodeFlagsNoTreePushOnOpen)
			if imgui.IsItemClicked() {
				obj, err := a.Create()
				_ = err
				err = asset.Save("test", obj)
				if err != nil {
					c.ed.raiseError(err)
				}
			}
		}
	}
}
