package asset

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type saveableAssetContainer struct {
	Type  string
	Inner interface{}
}

func (a *assetManagerImpl) Save(path Path, toSave Asset) error {
	if assetManager.WriteFS == nil {
		return fmt.Errorf("Can't Save asset - no writable FS")
	}
	// toSave must be a pointer, but the top level needs to be
	// saved as a struct
	if reflect.TypeOf(toSave).Kind() != reflect.Pointer {
		panic("Fatal, not a pointer being saved")
	}

	// go from the pointer to the real struct
	structToSave := reflect.ValueOf(toSave).Elem()

	fixedRefs := a.buildJsonToSave(structToSave.Interface())
	_, fullname := ObjectTypeName(toSave)
	container := saveableAssetContainer{
		Type:  fullname,
		Inner: fixedRefs,
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

	return nil
}

type assetLoadPath struct {
	Type string
	Path Path
}

func (a *assetManagerImpl) buildJsonToSave(obj any) any {
	if obj == nil {
		return nil
	}
	t := reflect.TypeOf(obj)
	switch t.Kind() {
	case reflect.Pointer:
		_, fullname := ObjectTypeName(obj)
		path := a.AssetToLoadPath[obj]
		return assetLoadPath{
			Type: fullname,
			Path: path,
		}
	case reflect.Struct:
		m := map[string]any{}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}
			m[field.Name] = a.buildJsonToSave(reflect.ValueOf(obj).Field(i).Interface())
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
				s[i] = a.buildJsonToSave(index.Interface())
			}
			return s
		}
	default:
		return obj
	}
}
