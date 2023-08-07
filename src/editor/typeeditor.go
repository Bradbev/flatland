package editor

// this file is the general editor implementation
import (
	"flatland/src/asset"
	"flatland/src/editor/edgui"
	"fmt"
	"io/fs"
	"log"
	"reflect"
	"strings"
	"unsafe"

	"github.com/inkyblackness/imgui-go/v4"
)

var logger = log.Default()

type TypeEditorFn func(*TypeEditContext, reflect.Value) error

type TypeEditContext struct {
	Ed *ImguiEditor

	// editContext is where GetContext saves its data
	editContext map[unsafe.Pointer]map[any]any

	// the stack of struct fields so nested editors
	// can see what their field description is
	structFieldStack []*reflect.StructField
}

func NewTypeEditContext(ed *ImguiEditor) *TypeEditContext {
	return &TypeEditContext{
		Ed:               ed,
		editContext:      map[unsafe.Pointer]map[any]any{},
		structFieldStack: []*reflect.StructField{},
	}
}

// GetContext exists just to help you find editor.GetContext[T](*TypeEditContext, reflect.Value)
func (c *TypeEditContext) GetContext(key reflect.Value) {
	panic("")
}

// EditValue allows the custom type editors to edit sub parts
// of themselves without needing to re-implement the editors
func (c *TypeEditContext) EditValue(value reflect.Value) {
	c.Ed.typeEditor.EditValue(c, value)
}

// Edit calls EditValue.  Custom type editors can use this to
// edit fields.
func (c *TypeEditContext) Edit(obj any) {
	c.Ed.typeEditor.Edit(c, obj)
}

func (c *TypeEditContext) PushStructField(sf *reflect.StructField) {
	c.structFieldStack = append(c.structFieldStack, sf)
}

func (c *TypeEditContext) PopStructField() {
	c.structFieldStack = c.structFieldStack[:len(c.structFieldStack)-1]
}

func (c *TypeEditContext) StructField() *reflect.StructField {
	if len(c.structFieldStack) == 0 {
		return nil
	}
	return c.structFieldStack[len(c.structFieldStack)-1]
}

// if a type stored by using GetContext implements Disposable
// then Dispose will be called when the asset editor window is closed
type Disposable interface {
	Dispose(context *TypeEditContext)
}

// GetContext returns a *T from the TypeEditContext.  Custom editors should
// use this function to save off context during edits
// returns true if this is the first time the context has been created
func GetContext[T any](context *TypeEditContext, key reflect.Value) (*T, bool) {
	ptr := key.Addr().UnsafePointer()
	contexts, contextsExists := context.editContext[ptr]
	if !contextsExists {
		// first level map doesn't exist yet
		contexts = map[any]any{}
		context.editContext[ptr] = contexts
	}
	var zeroT T

	// found the type/value context
	if context, exists := contexts[zeroT]; exists {
		return context.(*T), false
	}

	// second level map doesn't exist
	ret := new(T)
	contexts[zeroT] = ret
	return ret, true
}

func DisposeContext(context *TypeEditContext, key reflect.Value) {
	ptr := key.UnsafePointer()
	if contexts, exists := context.editContext[ptr]; exists {
		for _, value := range contexts {
			if dispose, ok := value.(Disposable); ok {
				dispose.Dispose(context)
			}
		}

		delete(context.editContext, ptr)
	}
}

type typeEditor struct {
	// typeEditFuncs map an asset type string to the function
	// that will be called when that type needs to be edited
	typeEditFuncs map[string]TypeEditorFn
	ed            *ImguiEditor
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
func (e *typeEditor) Edit(context *TypeEditContext, obj any) {
	value := reflect.ValueOf(obj)
	e.EditValue(context, value)
}

// EditValue accepts a reflect.Value and draws an editor window for that value
func (e *typeEditor) EditValue(context *TypeEditContext, value reflect.Value) {
	_, fullName := asset.TypeName(value.Type())
	if value.Kind() != reflect.Pointer {
		logger.Panicf("Value %v is not a pointer, this is a programming error", value)
	}
	// Get at the value being pointed to
	value = value.Elem()
	edFn := e.typeEditFuncs[fullName]
	if edFn == nil {
		switch value.Kind() {
		case reflect.Struct:
			edFn = structEd
		case reflect.Array:
			fallthrough
		case reflect.Slice:
			edFn = sliceEd
		}
	}

	if edFn == nil {
		logger.Printf("No editor for %s (Kind: %s)", fullName, value.Kind().String())
		return
	}
	if !value.CanSet() {
		logger.Panicf("Value %v is not settable, this is a programming error", value)
	}
	edFn(context, value)
}

func (e *typeEditor) addPrimitiveTypes() {
	e.AddType(new(float32), float32Edit)
	e.AddType(new(float64), float64Edit)
	e.AddType(new(bool), boolEdit)
	e.AddType(new(string), stringEdit)
	e.AddType(new(int), intEdit)
	e.AddType(new(asset.Path), pathEd)
}

// primitive type handler funcs below here

func structEd(context *TypeEditContext, value reflect.Value) error {
	t := value.Type()
	if t.Kind() != reflect.Struct {
		logger.Fatalf("Not a struct - %v", t.Kind())
	}

	treeNodeName := getNodeName(context, value)
	edgui.TreeNodeWithPop(treeNodeName+"##structEd", imgui.TreeNodeFlagsDefaultOpen, func() {
		imgui.BeginTable(treeNodeName+"##table", 2)
		for i := 0; i < t.NumField(); i++ {
			field := value.Field(i)
			structField := t.Field(i)
			func() {
				context.PushStructField(&structField)
				defer context.PopStructField()
				if structField.IsExported() {
					switch structField.Type.Kind() {
					case reflect.Array:
						fallthrough
					case reflect.Slice:
						fallthrough
					case reflect.Struct:
						// structs are a new TreeNode
						// End the current table, edit the value
						// in a new tree node and then Begin the table again
						imgui.EndTable()
						context.EditValue(field.Addr())
						imgui.BeginTable(treeNodeName+"##table", 2)

					default:
						imgui.TableNextRow()
						imgui.TableNextColumn()

						{ // Handle field name overrides
							fieldName := structField.Name
							if override, ok := structField.Tag.Lookup("flat"); ok {
								fieldName = override
							}
							imgui.Text(fieldName)
						}

						imgui.TableNextColumn()
						context.EditValue(field.Addr())
					}
				}
			}()
		}
		imgui.EndTable()
	})

	return nil
}

func getNodeName(context *TypeEditContext, value reflect.Value) string {
	t := value.Type()
	// select the nodeName for this slice edit
	// Typename, NamedAsset, FieldNameOverride
	nodeName, _ := asset.TypeName(t)
	if value.CanAddr() {
		iface := value.Addr().Interface()
		if namedAsset, ok := iface.(asset.NamedAsset); ok {
			nodeName = namedAsset.Name()
		}
	}
	// If this is a nested type, the higher stack level might have
	// wanted to override the name
	if sf := context.StructField(); sf != nil {
		nodeName = sf.Name
		if override, ok := sf.Tag.Lookup("flat"); ok {
			nodeName = override
		}
	}
	return nodeName
}

func sliceEd(context *TypeEditContext, value reflect.Value) error {
	t := value.Type()
	isSlice := t.Kind() == reflect.Slice
	isArray := t.Kind() == reflect.Array
	if !isArray && !isSlice {
		logger.Fatalf("Not a slice or array - %v", t.Kind())
	}

	treeNodeName := getNodeName(context, value)
	sliceLen := value.Len()
	treeNodeName = fmt.Sprintf("%s (%d)", treeNodeName, sliceLen)
	edgui.TreeNodeWithPop(treeNodeName+"##slicdEd", imgui.TreeNodeFlagsDefaultOpen, func() {
		edgui.WithID(value, func() {
			if isSlice {
				imgui.SameLine()
				imgui.Text("   ")
				imgui.SameLine()
				if imgui.Button("+") {
					value.Grow(1)
					value.SetLen(sliceLen + 1)
					// default init the new item
					value.Index(sliceLen).Set(reflect.New(value.Index(0).Type()).Elem())
					sliceLen = value.Len()
				}
				imgui.SameLine()
				if imgui.Button("Clear") {
					value.SetLen(0)
					sliceLen = 0
				}
			}
			toDelete := -1
			for i := 0; i < sliceLen; i++ {
				edgui.Text("%d", i)
				imgui.SameLine()
				index := value.Index(i)
				edgui.WithID(index, func() {
					context.EditValue(index.Addr())
					if isSlice {
						imgui.SameLine()
						if imgui.Button("X") {
							toDelete = i
						}
					}
				})
			}
			if toDelete != -1 {
				for i := toDelete; i < sliceLen-1; i++ {
					value.Index(i).Set(value.Index(i + 1))
				}
				value.SetLen(sliceLen - 1)
			}
		})
	})

	return nil
}

func withID(value reflect.Value, body func()) {
	addr := fmt.Sprintf("%x", value.UnsafeAddr())
	imgui.PushID(addr)
	defer imgui.PopID()
	body()
}

func float32Edit(context *TypeEditContext, value reflect.Value) error {
	withID(value, func() {
		addr := value.Addr().Interface().(*float32)
		imgui.DragFloat("", addr)
	})
	return nil
}

func float64Edit(context *TypeEditContext, value reflect.Value) error {
	withID(value, func() {
		f32 := float32(value.Float())
		imgui.DragFloat("", &f32)
		value.SetFloat(float64(f32))
	})
	return nil
}

func boolEdit(context *TypeEditContext, value reflect.Value) error {
	withID(value, func() {
		addr := value.Addr().Interface().(*bool)
		imgui.Checkbox("", addr)
	})
	return nil
}

func stringEdit(context *TypeEditContext, value reflect.Value) error {
	withID(value, func() {
		addr := value.Addr().Interface().(*string)
		imgui.InputText("", addr)
	})
	return nil
}

func intEdit(context *TypeEditContext, value reflect.Value) error {
	withID(value, func() {
		i32 := int32(value.Int())
		imgui.InputInt("", &i32)
		value.SetInt(int64(i32))
	})
	return nil
}

type pathEdContext struct {
	auto *edgui.AutoComplete
}

func pathEd(context *TypeEditContext, value reflect.Value) error {
	onActivated := func() []string {
		filters := []string{}
		if sf := context.StructField(); sf != nil {
			val, _ := sf.Tag.Lookup("filter")
			filters = strings.Split(strings.ToLower(val), ",")

		}

		var items []string
		asset.WalkFiles(func(path string, d fs.DirEntry, err error) error {
			// do not include directories
			if d != nil && d.IsDir() {
				return nil
			}
			// include everything if there is no filter
			if len(filters) == 0 {
				items = append(items, path)
				return nil
			}
			// otherwise only show files that contain the filter
			for _, filter := range filters {
				if strings.Contains(path, filter) {
					items = append(items, path)
					return nil
				}
			}
			return nil
		})
		return items
	}

	c, firstTime := GetContext[pathEdContext](context, value)
	if firstTime {
		c.auto = &edgui.AutoComplete{}
	}
	path := value.Addr().Interface().(*asset.Path)
	s := (*string)(path)
	c.auto.InputText("", s, onActivated)
	return nil
}
