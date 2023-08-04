package editor

// this file is the general editor implementation
import (
	"flatland/src/asset"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/inkyblackness/imgui-go/v4"
)

var logger = log.Default()

type TypeEditorFn func(*ImguiEditor, reflect.Value) error

type typeEditor struct {
	// typeEditFuncs map an asset type string to the function
	// that will be called when that type needs to be edited
	typeEditFuncs map[string]TypeEditorFn
	ed            *ImguiEditor
}

type editorWriteFS struct {
	base string
}

func (e *editorWriteFS) WriteFile(path string, data []byte) error {
	return os.WriteFile(filepath.Join(e.base, path), data, 0777)
}

func WriteFS(base string) asset.WriteableFileSystem {
	return &editorWriteFS{base: base}
}

func newTypeEditor() *typeEditor {
	ret := &typeEditor{
		typeEditFuncs: map[string]TypeEditorFn{},
	}
	ret.addPrimitiveTypes()

	return ret
}

func (e *typeEditor) AddType(typeToAdd any, edit TypeEditorFn) {
	_, fullName := asset.ObjectTypeName(typeToAdd)
	e.typeEditFuncs[fullName] = edit
}

// Edit accepts any object and draws an editor window for it
func (e *typeEditor) Edit(obj any) {
	value := reflect.ValueOf(obj)
	e.EditValue(value)
}

// EditValue accepts a reflect.Value and draws an editor window for that value
func (e *typeEditor) EditValue(value reflect.Value) {
	_, fullName := asset.TypeName(value.Type())
	if value.Kind() != reflect.Pointer {
		logger.Panicf("Value %v is not a pointer, this is a programming error", value)
	}
	// Get at the value being pointed to
	value = value.Elem()
	edFn := e.typeEditFuncs[fullName]
	if edFn == nil && value.Kind() == reflect.Struct {
		edFn = structEd
	}

	if edFn == nil {
		logger.Printf("No editor for %s", fullName)
		return
	}
	if !value.CanSet() {
		logger.Panicf("Value %v is not settable, this is a programming error", value)
	}
	edFn(e.ed, value)
}

func (e *typeEditor) addPrimitiveTypes() {
	e.AddType(new(float32), float32Edit)
	e.AddType(new(float64), float64Edit)
	e.AddType(new(bool), boolEdit)
	e.AddType(new(string), stringEdit)
	e.AddType(new(int), intEdit)
}

// primitive type handler funcs below here

func structEd(types *ImguiEditor, value reflect.Value) error {
	t := value.Type()
	if t.Kind() != reflect.Struct {
		logger.Fatalf("Not a struct - %v", t.Kind())
	}
	name, _ := asset.TypeName(t)
	imgui.BeginTable(name, 2)
	for i := 0; i < t.NumField(); i++ {
		field := value.Field(i)
		structField := t.Field(i)
		if structField.IsExported() {
			imgui.TableNextRow()

			imgui.TableNextColumn()
			imgui.Text(structField.Name)

			imgui.TableNextColumn()
			types.EditValue(field.Addr())
		}
	}
	imgui.EndTable()

	return nil
}

func withID(value reflect.Value, body func()) {
	addr := fmt.Sprintf("%x", value.UnsafeAddr())
	imgui.PushID(addr)
	defer imgui.PopID()
	body()
}

func float32Edit(types *ImguiEditor, value reflect.Value) error {
	withID(value, func() {
		addr := value.Addr().Interface().(*float32)
		imgui.DragFloat("", addr)
	})
	return nil
}

func float64Edit(types *ImguiEditor, value reflect.Value) error {
	withID(value, func() {
		f32 := float32(value.Float())
		imgui.DragFloat("", &f32)
		value.SetFloat(float64(f32))
	})
	return nil
}

func boolEdit(types *ImguiEditor, value reflect.Value) error {
	withID(value, func() {
		addr := value.Addr().Interface().(*bool)
		imgui.Checkbox("", addr)
	})
	return nil
}

func stringEdit(types *ImguiEditor, value reflect.Value) error {
	withID(value, func() {
		addr := value.Addr().Interface().(*string)
		imgui.InputText("", addr)
	})
	return nil
}

func intEdit(types *ImguiEditor, value reflect.Value) error {
	withID(value, func() {
		i32 := int32(value.Int())
		imgui.InputInt("", &i32)
		value.SetInt(int64(i32))
	})
	return nil
}
