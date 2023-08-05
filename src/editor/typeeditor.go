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

	"github.com/inkyblackness/imgui-go/v4"
)

var logger = log.Default()

type TypeEditorFn func(*TypeEditContext, reflect.Value) error

type TypeEditContext struct {
	Ed *ImguiEditor
}

func (c *TypeEditContext) EditValue(value reflect.Value) {
	c.Ed.typeEditor.EditValue(c, value)
}
func (c *TypeEditContext) Edit(obj any) {
	c.Ed.typeEditor.Edit(c, obj)
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

type fieldEditContext struct {
	fieldNameOverride string
}

func structEd(context *TypeEditContext, value reflect.Value) error {
	t := value.Type()
	if t.Kind() != reflect.Struct {
		logger.Fatalf("Not a struct - %v", t.Kind())
	}

	// select the name for this struct edit
	// Typename, NamedAsset, FieldNameOverride
	name, _ := asset.TypeName(t)
	if value.CanAddr() {
		iface := value.Addr().Interface()
		if namedAsset, ok := iface.(asset.NamedAsset); ok {
			name = namedAsset.Name()
		}
	}
	// If this is a nested type, the higher stack level might have
	// wanted to override the name
	ctx, _ := GetContext[fieldEditContext](context, value)
	if ctx.fieldNameOverride != "" {
		name = ctx.fieldNameOverride
	}
	edgui.TreeNodeWithPop(name, imgui.TreeNodeFlagsDefaultOpen, func() {
		imgui.BeginTable(name+"##table", 2)
		for i := 0; i < t.NumField(); i++ {
			field := value.Field(i)
			structField := t.Field(i)
			if structField.IsExported() {
				sfContext, _ := GetContext[*reflect.StructField](context, field)
				*sfContext = &structField

				if structField.Type.Kind() == reflect.Struct {
					// structs are a new TreeNode
					// disable the current table, edit the value
					// in a new tree node and then restart the table
					imgui.EndTable()
					// set the name for the new tree
					ctx, _ := GetContext[fieldEditContext](context, field)
					ctx.fieldNameOverride = structField.Name
					if name, ok := structField.Tag.Lookup("flat"); ok {
						ctx.fieldNameOverride = name
					}
					context.EditValue(field.Addr())
					imgui.BeginTable(name+"##table", 2)
					continue
				}

				imgui.TableNextRow()

				imgui.TableNextColumn()
				imgui.Text(structField.Name)

				imgui.TableNextColumn()
				context.EditValue(field.Addr())
			}
		}
		imgui.EndTable()
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
		structFieldPtr, _ := GetContext[*reflect.StructField](context, value)
		structField := *structFieldPtr
		val, _ := structField.Tag.Lookup("filter")
		filters := strings.Split(strings.ToLower(val), ",")

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
