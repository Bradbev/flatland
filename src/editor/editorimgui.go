package editor

import (
	"fmt"
	"io/fs"
	"os"
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

func ImguiTypeEditor() *TypeEditor {
	ed := NewTypeEditor(&imguiEditorImpl{})
	ed.AddType(new(float32), float32Edit)
	ed.AddType(new(float64), float64Edit)
	ed.AddType(new(bool), boolEdit)
	ed.AddType(new(string), stringEdit)
	ed.AddType(new(int), intEdit)
	return ed
}

func structEd(types *TypeEditor, value reflect.Value) error {
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

func float32Edit(types *TypeEditor, value reflect.Value) error {
	withID(value, func() {
		addr := value.Addr().Interface().(*float32)
		imgui.DragFloat("", addr)
	})
	return nil
}

func float64Edit(types *TypeEditor, value reflect.Value) error {
	withID(value, func() {
		f32 := float32(value.Float())
		imgui.DragFloat("", &f32)
		value.SetFloat(float64(f32))
	})
	return nil
}

func boolEdit(types *TypeEditor, value reflect.Value) error {
	withID(value, func() {
		addr := value.Addr().Interface().(*bool)
		imgui.Checkbox("", addr)
	})
	return nil
}

func stringEdit(types *TypeEditor, value reflect.Value) error {
	withID(value, func() {
		addr := value.Addr().Interface().(*string)
		imgui.InputText("", addr)
	})
	return nil
}

func intEdit(types *TypeEditor, value reflect.Value) error {
	withID(value, func() {
		i32 := int32(value.Int())
		imgui.InputInt("", &i32)
		value.SetInt(int64(i32))
	})
	return nil
}

type ImguiEditor struct {
	typeEditor  *TypeEditor
	contentPath string
	fsys        fs.FS
}

func NewImguiEditor() *ImguiEditor {
	ret := &ImguiEditor{
		typeEditor:  ImguiTypeEditor(),
		contentPath: "./content",
		fsys:        os.DirFS("./content"),
	}
	return ret
}

type fswalk struct {
	path  string
	dirs  []*fswalk
	files []string
}

func (e *ImguiEditor) Update(deltaseconds float32) error {
	defer imgui.End()
	if !imgui.Begin("EditorMainWindow") {
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

func (e *ImguiEditor) buildFileCache() *fswalk {
	// todo - optimise this and cache state, filling it
	// as folders are opened
	stack := []*fswalk{{path: ""}}
	peek := func() *fswalk {
		return stack[len(stack)-1]
	}

	fs.WalkDir(e.fsys, ".", func(path string, d fs.DirEntry, err error) error {
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
	if imgui.BeginPopupModal("AddAssetModal") {
		defer imgui.EndPopup()
		imgui.TreeNodeV("List here",
			imgui.TreeNodeFlagsLeaf|imgui.TreeNodeFlagsNoTreePushOnOpen)
		if imgui.IsItemClicked() {
			println("click")
		}
	}
}
