package editor

import (
	"flatland/src/asset"
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/inkyblackness/imgui-go/v4"
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
	ed := NewTypeEditor(&imguiEditorImpl{})
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
	commonEditor   *CommonEditor
	contentPath    string
	fsys           fs.FS
	errorModalText string
}

func NewImguiEditor() *ImguiEditor {
	ret := &ImguiEditor{
		commonEditor: ImguiTypeEditor(),
	}

	return ret
}

type fswalk struct {
	path  string
	dirs  []*fswalk
	files []string
}

func (e *ImguiEditor) Update(deltaseconds float32) error {
	e.contentWindow()
	return nil
}

func (e *ImguiEditor) contentWindow() error {
	defer imgui.End()
	if !imgui.Begin("Content Browser") {
		return nil
	}
	if imgui.Button("Add") {
		imgui.OpenPopup("AddAssetModal")
	}
	e.drawAddAssetModal()

	cache := e.buildFileCache()
	e.walkCache(cache)

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

func (e *ImguiEditor) walkCache(node *fswalk) {
	path := filepath.Base(node.path)
	if path == "." {
		path = "content"
	}
	if imgui.TreeNodeV(path, imgui.TreeNodeFlagsDefaultOpen) {
		for _, n := range node.dirs {
			e.walkCache(n)
		}
		for _, f := range node.files {
			imgui.TreeNodeV(filepath.Base(f),
				imgui.TreeNodeFlagsLeaf|imgui.TreeNodeFlagsNoTreePushOnOpen)
		}
		imgui.TreePop()
	}
}

func (e *ImguiEditor) drawAddAssetModal() {
	open := true
	if imgui.BeginPopupModalV("AddAssetModal", &open, imgui.WindowFlagsAlwaysAutoResize) {
		defer imgui.EndPopup()
		defer e.displayErrorModal()

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
					e.raiseError(err)
				}
			}
		}
	}
}
