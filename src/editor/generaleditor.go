package editor

// this file is the general editor implementation
// in theory you should be able to use some other
// GUI implementation without changing this file

import (
	"flatland/src/asset"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
)

var logger = log.Default()

type TypeEditorFn func(*CommonEditor, reflect.Value) error

type CommonEditor struct {
	typeEditors map[string]TypeEditorFn
	impl        EditorImpl
	contentPath string
	fsysRead    fs.FS
	fsysWrite   asset.WriteableFileSystem
}

type EditorImpl interface {
	BeginStruct(name string)
	EndStruct()
	FieldName(name string)
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

func NewTypeEditor(impl EditorImpl) *CommonEditor {
	path := "./content"
	ret := &CommonEditor{
		typeEditors: map[string]TypeEditorFn{},
		impl:        impl,
		contentPath: path,
		fsysRead:    os.DirFS(path),
		fsysWrite:   WriteFS(path),
	}
	asset.RegisterFileSystem(ret.fsysRead, 0)
	asset.RegisterWritableFileSystem(ret.fsysWrite)
	return ret
}

func (t *CommonEditor) AddType(typeToAdd any, edit TypeEditorFn) {
	_, fullName := asset.ObjectTypeName(typeToAdd)
	t.typeEditors[fullName] = edit
}

func (t *CommonEditor) Edit(obj any) {
	value := reflect.ValueOf(obj)
	t.EditValue(value)
}

func (t *CommonEditor) EditValue(value reflect.Value) {
	_, fullName := asset.TypeName(value.Type())
	if value.Kind() != reflect.Pointer {
		logger.Panicf("Value %v is not a pointer, this is a programming error", value)
	}
	// Get at the value being pointed to
	value = value.Elem()
	edFn := t.typeEditors[fullName]
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
	edFn(t, value)
}
