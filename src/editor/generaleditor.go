package editor

import (
	"flatland/src/asset"
	"fmt"
	"log"
	"reflect"

	"github.com/inkyblackness/imgui-go/v4"
)

var logger = log.Default()

type TypeEditorFn func(*TypeEditor, reflect.Value) error

type TypeEditor struct {
	typeEditors map[string]TypeEditorFn
	impl        EditorImpl
}

type EditorImpl interface {
	BeginStruct(name string)
	EndStruct()
	FieldName(name string)
}

func NewTypeEditor(impl EditorImpl) *TypeEditor {
	return &TypeEditor{
		typeEditors: map[string]TypeEditorFn{},
		impl:        impl,
	}
}

func Default() *TypeEditor {
	ed := NewTypeEditor(&imguiEditorImpl{})
	ed.AddType(new(float32), float32Edit)
	ed.AddType(new(float64), float64Edit)
	ed.AddType(new(bool), boolEdit)
	ed.AddType(new(string), stringEdit)
	ed.AddType(new(int), intEdit)
	return ed
}

func (t *TypeEditor) AddType(typeToAdd any, edit TypeEditorFn) {
	toAdd := asset.ObjectTypeName(typeToAdd)
	t.typeEditors[toAdd] = edit
}

func (t *TypeEditor) Edit(obj any) {
	value := reflect.ValueOf(obj)
	t.EditValue(value)
}

func (t *TypeEditor) EditValue(value reflect.Value) {
	valueType := asset.TypeName(value.Type())
	if value.Kind() != reflect.Pointer {
		logger.Panicf("Value %v is not a pointer, this is a programming error", value)
	}
	// Get at the value being pointed to
	value = value.Elem()
	edFn := t.typeEditors[valueType]
	if edFn == nil && value.Kind() == reflect.Struct {
		edFn = structEd
	}

	if edFn == nil {
		logger.Printf("No editor for %s", valueType)
		return
	}
	if !value.CanSet() {
		logger.Panicf("Value %v is not settable, this is a programming error", value)
	}
	edFn(t, value)
}

func Edit(obj any, editors *TypeEditor) {
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
