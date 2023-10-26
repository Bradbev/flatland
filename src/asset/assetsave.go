package asset

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func (a *assetManagerImpl) Save(path Path, toSave Asset) error {
	container, err := a.toDiskFormat(toSave)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(container, "", "  ")
	if err != nil {
		return err
	}

	if !strings.HasSuffix(string(path), ".json") {
		path = path + ".json"
	}
	err = assetManager.WriteFS.WriteFile(path, data)

	if err != nil {
		return err
	}
	a.AssetToLoadPath[toSave] = path
	a.LoadPathToAsset[path] = toSave

	a.reloadChildAssets(path, toSave)
	return nil
}

// toDiskFormat converts toSave into Common format and wraps it with
// the on-disk format
func (a *assetManagerImpl) toDiskFormat(toSave Asset) (*onDiskSaveFormat, error) {
	// toSave must be a pointer, but the top level needs to be
	// saved as a struct
	if reflect.TypeOf(toSave).Kind() != reflect.Pointer {
		panic("Fatal, not a pointer being saved")
	}

	// go from the pointer to the real struct
	structToSave := reflect.ValueOf(toSave).Elem()
	commonFormat := a.toCommonFormat(structToSave.Interface())

	var parentCommonFormat any
	if parentPath, ok := a.ChildToParent[toSave]; ok {
		parent, err := a.Load(parentPath)
		if err != nil {
			return nil, err
		}
		if parent != nil {
			parentConcrete := reflect.ValueOf(parent).Elem().Interface()
			parentCommonFormat = a.toCommonFormat(parentConcrete)
		}
	}

	diffsFromParent := findDiffsFromParentCommonFormat(parentCommonFormat, commonFormat)
	_, fullname := ObjectTypeName(toSave)
	if _, ok := a.AssetDescriptors[fullname]; !ok {
		return nil, fmt.Errorf("Type %s is not registered with the asset system", fullname)
	}

	container := onDiskSaveFormat{
		Type:   fullname,
		Inner:  diffsFromParent,
		Parent: a.ChildToParent[toSave],
	}

	return &container, nil
}

// reloadChildAssets walks all children assets and force-reload
// Intended to be called after a parent asset has been saved.
func (a *assetManagerImpl) reloadChildAssets(path Path, parent Asset) {
	for childAsset, parentPath := range a.ChildToParent {
		if path == parentPath {
			a.refreshParentValuesForChild(childAsset, parentPath)
		}
	}
}

func (a *assetManagerImpl) refreshParentValuesForChild(childAsset Asset, parentPath Path) {
	overrides := a.ChildAssetOverrides[childAsset]

	// we don't to convert the *interface* to commonFormat, we want the
	// struct that the asset is really referring to in commonFormat
	c := reflect.ValueOf(childAsset).Elem().Interface()
	child := a.toCommonFormatInternal(c, &commonFormatContext{overrides: overrides})

	desc := a.GetAssetDescriptor(childAsset)
	a.loadFromCommonFormat(desc.FullName, parentPath, child, childAsset)
}

type assetLoadPath struct {
	Type string
	Path Path
}

type commonFormatContext struct {
	stack     []*reflect.StructField
	overrides *childOverrides
}

func (b *commonFormatContext) Push(sf *reflect.StructField) {
	b.stack = append(b.stack, sf)
}

func (b *commonFormatContext) Pop() {
	b.stack = b.stack[:len(b.stack)-1]
}

func (b *commonFormatContext) Peek() *reflect.StructField {
	if len(b.stack) > 0 {
		return b.stack[len(b.stack)-1]
	}
	return nil
}

func (b *commonFormatContext) StackPath() string {
	var result strings.Builder
	for i, field := range b.stack {
		result.WriteString(field.Name)
		if i < len(b.stack)-1 {
			result.WriteString(".")
		}
	}
	return result.String()
}

// toCommonFormat takes an object and converts it to the internal Common format
// The Common format erases type information so must be unmarshalled back into
// concrete Go types that do have type information.  Common format mostly exists
// to facilitate the replacing of pointers with other structures that are completely
// unrelated to the original pointer's type.
//
// The Common format is
//   - pointers are replaced by either
//     a) savedAssetContainers for "inline" members OR
//     b) assetLoadPath so normal assets can be loaded
//   - structs are replaced with map[string]any for all exported fields
//   - slices of bytes are uuencoded to strings
//   - everything else remains the same
func (a *assetManagerImpl) toCommonFormat(obj any) any {
	return a.toCommonFormatInternal(obj, &commonFormatContext{})
}

func (a *assetManagerImpl) toCommonFormatInternal(obj any, context *commonFormatContext) any {
	if obj == nil {
		return nil
	}
	path := context.StackPath()
	isPathOverridden := context.overrides == nil || context.overrides.PathHasOverride(path)
	t := reflect.TypeOf(obj)
	switch t.Kind() {
	case reflect.Pointer:
		// Save inline assets
		if sf := context.Peek(); sf != nil {
			if _, inline := GetFlatTag(sf, "inline"); inline {
				container, err := a.toDiskFormat(obj)
				if err != nil {
					panic(err)
				}
				return container
			}
		}
		// save references to known assets
		if path, ok := a.AssetToLoadPath[obj]; ok {
			_, fullname := ObjectTypeName(obj)
			return &assetLoadPath{
				Type: fullname,
				Path: path,
			}
		}
		log.Printf("Field '%s' will not be saved.  Pointer is not inline, nor to a known asset", context.StackPath())
		//return a.toCommonFormatInternal(reflect.ValueOf(obj).Elem().Interface(), context)
		return nil
	case reflect.Struct:
		m := map[string]any{}
		v := reflect.ValueOf(obj)
		for i := 0; i < t.NumField(); i++ {
			structField := t.Field(i)
			if !structField.IsExported() { // ignore unexported fields
				continue
			}
			field := v.Field(i)
			context.Push(&structField)
			c := a.toCommonFormatInternal(field.Interface(), context)
			if c != nil {
				m[structField.Name] = c
			}
			context.Pop()
		}
		return m
	case reflect.Slice:
		if !isPathOverridden {
			return nil
		}
		v := reflect.ValueOf(obj)
		l := v.Len()
		if l > 0 && v.Index(0).Kind() == reflect.Uint8 {
			// byte slices are uuencoded into a string (because the json package does it that way)
			bytes := v.Bytes()
			encoded := base64.StdEncoding.EncodeToString(bytes)
			return encoded
		} else {
			s := make([]any, v.Len())
			for i := 0; i < v.Len(); i++ {
				index := v.Index(i)
				s[i] = a.toCommonFormatInternal(index.Interface(), context)
			}
			return s
		}
	default:
		if !isPathOverridden {
			return nil
		}
		return obj
	}
}

func jsp(pref string, o any) {
	fmt.Println(pref)
	fmt.Printf("%#v\n", o)
}

func (a *assetManagerImpl) findDiffsFromParent(parent, child any) any {
	if child == nil {
		return nil
	}
	var parentJson any
	if parent != nil {
		// we don't want the common format for a pointer to the parent, but
		// to the real struct
		parentConcrete := reflect.ValueOf(parent).Elem().Interface()
		parentJson = a.toCommonFormat(parentConcrete)
	}
	childConcrete := reflect.ValueOf(child).Elem().Interface()
	childJson := a.toCommonFormat(childConcrete)
	return findDiffsFromParentCommonFormat(parentJson, childJson)
}

// findDiffsFromParentCommonFormat expects to take output from toCommonFormat
// The input objects are compared recursively.  If parent and child are the same
// then nil is returned.
// Maps recursively apply this function to each key.
// The results is a map that contains only key/value pairs where the child
// is different from the Parent.  In the degenerate case, a copy of child will
// be returned (in Common format).
func findDiffsFromParentCommonFormat(parent, child any) any {
	//jsp("Parent ----", parent)
	//jsp("Child ----", child)

	childType := reflect.TypeOf(child)
	parentValue := reflect.ValueOf(parent)
	childValue := reflect.ValueOf(child)
	if parent != nil {
		parentType := reflect.TypeOf(parent)
		if parentType != childType {
			// types differ, keep the child (unless it's the zero value)
			if child == nil || childValue.IsZero() {
				return nil
			}
			return child
		}
		if !parentValue.IsZero() && childValue.IsZero() {
			// if the parent is not Zero, but the child is, the zero
			// value needs to be serialized.
			return child
		}
	}

	if child == nil || childValue.IsZero() {
		return nil
	}

	switch childType.Kind() {
	default:
		if parent == child {
			return nil
		}
	case reflect.Slice, reflect.Array:
		if childValue.IsZero() || childValue.Len() == 0 {
			return nil
		}
		// can't call IsNil on an array
		if childType.Kind() == reflect.Slice && childValue.IsNil() {
			return nil
		}
		if parent == nil || childValue.Len() != parentValue.Len() {
			return child
		}
		for i := 0; i < childValue.Len(); i++ {
			pv := parentValue.Index(i).Interface()
			cv := childValue.Index(i).Interface()
			diff := findDiffsFromParentCommonFormat(pv, cv)
			if diff != nil {
				return child
			}
		}
		return nil

	case reflect.Map:
		ret := reflect.MakeMap(childType)
		childIter := childValue.MapRange()
		for childIter.Next() {
			k, v := childIter.Key(), childIter.Value()
			var parentMapValue any = nil
			if parentValue.Kind() == reflect.Map {
				if pmv := parentValue.MapIndex(k); pmv.IsValid() {
					parentMapValue = pmv.Interface()
				}
			}

			diffs := findDiffsFromParentCommonFormat(parentMapValue, v.Interface())
			if diffs != nil {
				ret.SetMapIndex(k, reflect.ValueOf(diffs))
			}
		}
		if len(ret.MapKeys()) == 0 {
			// if there are no keys, we found no diffs so can ignore this struct when saving the child
			return nil
		}
		return ret.Interface()
	}
	// if we got to here without earlying out, safest to return the child
	return child
}
