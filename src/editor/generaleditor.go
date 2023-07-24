package editor

// this file is the general editor implementation
// in theory you should be able to use some other
// GUI implementation without changing this file

import (
	"flatland/src/asset"
	"log"
	"reflect"
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
