package asset

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func (a *assetManagerImpl) Save(path Path, toSave Asset) error {
	container, err := a.makeSavedAssetContainer(toSave)
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

	a.reloadChildAssets(path)
	return nil
}

func (a *assetManagerImpl) makeSavedAssetContainer(toSave Asset) (*savedAssetContainer, error) {
	if assetManager.WriteFS == nil {
		return nil, fmt.Errorf("Can't Save asset - no writable FS")
	}
	// toSave must be a pointer, but the top level needs to be
	// saved as a struct
	if reflect.TypeOf(toSave).Kind() != reflect.Pointer {
		panic("Fatal, not a pointer being saved")
	}

	// go from the pointer to the real struct
	structToSave := reflect.ValueOf(toSave).Elem()
	fixedRefs := a.buildJsonToSave(structToSave.Interface())

	var parentJson any
	if parentPath, ok := a.ChildToParent[toSave]; ok {
		parent, err := a.Load(parentPath)
		if err != nil {
			return nil, err
		}
		if parent != nil {
			parentConcrete := reflect.ValueOf(parent).Elem().Interface()
			parentJson = a.buildJsonToSave(parentConcrete)
		}
	}

	diffsFromParent := findDiffsFromParentJson(parentJson, fixedRefs)
	_, fullname := ObjectTypeName(toSave)
	if _, ok := a.AssetDescriptors[fullname]; !ok {
		return nil, fmt.Errorf("Type %s is not registered with the asset system", fullname)
	}

	container := savedAssetContainer{
		Type:   fullname,
		Inner:  diffsFromParent,
		Parent: a.ChildToParent[toSave],
	}

	return &container, nil
}

func (a *assetManagerImpl) reloadChildAssets(path Path) {
	for childAsset, parentPath := range a.ChildToParent {
		if path == parentPath {
			childPath := a.AssetToLoadPath[childAsset]
			a.LoadWithOptions(childPath, LoadOptions{ForceReload: true})
		}
	}
}

type assetLoadPath struct {
	Type string
	Path Path
}

type buildJsonContext struct {
	stack []*reflect.StructField
}

func (b *buildJsonContext) Push(sf *reflect.StructField) {
	b.stack = append(b.stack, sf)
}

func (b *buildJsonContext) Pop() {
	b.stack = b.stack[:len(b.stack)-1]
}

func (b *buildJsonContext) Peek() *reflect.StructField {
	if len(b.stack) > 0 {
		return b.stack[len(b.stack)-1]
	}
	return nil
}

func (a *assetManagerImpl) buildJsonToSave(obj any) any {
	return a.buildJsonToSaveInternal(obj, &buildJsonContext{})
}

func (a *assetManagerImpl) buildJsonToSaveInternal(obj any, context *buildJsonContext) any {
	if obj == nil {
		return nil
	}
	t := reflect.TypeOf(obj)
	switch t.Kind() {
	case reflect.Pointer:
		if sf := context.Peek(); sf != nil {
			if _, inline := GetFlatTag(sf, "inline"); inline {
				container, err := a.makeSavedAssetContainer(obj)
				if err != nil {
					panic(err)
				}
				return container
			}
		}
		// save references to known assets
		if path, ok := a.AssetToLoadPath[obj]; ok {
			_, fullname := ObjectTypeName(obj)
			return assetLoadPath{
				Type: fullname,
				Path: path,
			}
		}
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
			m[structField.Name] = a.buildJsonToSaveInternal(field.Interface(), context)
			context.Pop()
		}
		return m
	case reflect.Slice:
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
				s[i] = a.buildJsonToSaveInternal(index.Interface(), context)
			}
			return s
		}
	default:
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
		parentConcrete := reflect.ValueOf(parent).Elem().Interface()
		parentJson = a.buildJsonToSave(parentConcrete)
	}
	childConcrete := reflect.ValueOf(child).Elem().Interface()
	childJson := a.buildJsonToSave(childConcrete)
	return findDiffsFromParentJson(parentJson, childJson)
}

// findDiffsFromParentJson expects to revceive output from buildJsonToSave
// ie, all structs have been converted to maps.  The input objects are
// compared recursively.  If parent and child are the same, then nil is
// returned.
// Maps recursively apply this function to each key.
// The results is a map that contains only key/value pairs where the child
// is different from the Parent.  In the degenerate case, a copy of child will
// be returned.
func findDiffsFromParentJson(parent, child any) any {
	//jsp("Parent ----", parent)
	//jsp("Child ----", child)

	childType := reflect.TypeOf(child)
	parentValue := reflect.ValueOf(parent)
	childValue := reflect.ValueOf(child)
	if parent != nil {
		parentType := reflect.TypeOf(parent)
		if parentType != childType {
			// types differ, keep the child (unless it's the zero value)
			if childValue.IsZero() {
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
			diff := findDiffsFromParentJson(pv, cv)
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

			diffs := findDiffsFromParentJson(parentMapValue, v.Interface())
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
