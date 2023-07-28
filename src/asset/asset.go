package asset

/*
* Assets on disk have the same format - assetContainer.  This struct holds
* some meta data about the asset and then an Inner that holds the user
* defined data.
 */

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"reflect"
	"strings"

	"golang.org/x/exp/slices"
)

type AssetDescriptor struct {
	Name     string
	FullName string
	Create   FactoryFunc
}

type Asset interface{}
type PostLoadingAsset interface {
	PostLoad()
}

type PreSavingAsset interface {
	PreSave()
}

type FactoryFunc func() (Asset, error)

func RegisterFileSystem(filesystem fs.FS, priority int) error {
	wrapped := &fsWrapper{FileSystem: filesystem, Priority: priority}
	return assetManager.AddFS(wrapped)
}

func RegisterWritableFileSystem(filesystem WriteableFileSystem) error {
	assetManager.WriteFS = filesystem
	return nil
}

func RegisterAssetFactory(zeroAsset any, factoryFunction FactoryFunc) {
	assetManager.RegisterAssetFactory(zeroAsset, factoryFunction)
}

func RegisterAsset(zeroAsset any) bool {
	zeroType := reflect.TypeOf(zeroAsset)

	assetManager.RegisterAssetFactory(zeroAsset, func() (Asset, error) {
		zero := reflect.New(zeroType)
		return zero.Interface().(Asset), nil
	})
	return true
}

func ReadFile(assetPath string) ([]byte, error) {
	return assetManager.ReadFile(assetPath)
}

func Load(assetPath string) (Asset, error) {
	return assetManager.Load(assetPath)
}

func Save(path string, toSave Asset) error {
	return assetManager.Save(path, toSave)
}

func Reset() {
	assetManager = newAssetManagerImpl()
}

func ListAssets() []*AssetDescriptor {
	return assetManager.AssetDescriptorList
}

func ObjectTypeName(obj any) (name string, fullname string) {
	return TypeName(reflect.TypeOf(obj))
}

func TypeName(t reflect.Type) (name string, fullname string) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t.Name(), t.PkgPath() + "." + t.Name()
}

type assetContainer struct {
	Type  string
	Inner json.RawMessage
}

type assetManagerImpl struct {
	FileSystems         []*fsWrapper
	AssetDescriptors    map[string]*AssetDescriptor
	AssetDescriptorList []*AssetDescriptor
	WriteFS             WriteableFileSystem

	// This map tracks in-memory assets to their load path
	// it is needed when assets are saved to convert the Asset
	// to a path
	AssetToLoadPath map[Asset]string

	// The opposite mapping
	LoadPathToAsset map[string]Asset
}

type fsWrapper struct {
	FileSystem fs.FS
	Priority   int
}

var assetManager = newAssetManagerImpl()

func newAssetManagerImpl() *assetManagerImpl {
	return &assetManagerImpl{
		FileSystems:      []*fsWrapper{},
		AssetDescriptors: map[string]*AssetDescriptor{},
		AssetToLoadPath:  map[Asset]string{},
		LoadPathToAsset:  map[string]Asset{},
	}
}

func (a *assetManagerImpl) AddFS(wrapper *fsWrapper) error {
	a.FileSystems = append(a.FileSystems, wrapper)
	slices.SortFunc(a.FileSystems, func(a, b *fsWrapper) bool {
		return a.Priority < b.Priority
	})
	return nil
}

func (a *assetManagerImpl) ReadFile(path string) ([]byte, error) {
	for _, fsys := range a.FileSystems {
		data, err := fs.ReadFile(fsys.FileSystem, path)
		if err == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("Unable to find path (%s) in any registered FS ", path)
}

func (a *assetManagerImpl) RegisterAssetFactory(zeroAsset any, factoryFunction FactoryFunc) {
	name, typeName := ObjectTypeName(zeroAsset)
	println("Registered asset ", typeName)
	descriptor := &AssetDescriptor{
		Name:     name,
		FullName: typeName,
		Create:   factoryFunction,
	}
	a.AssetDescriptors[typeName] = descriptor
	a.AssetDescriptorList = append(a.AssetDescriptorList, descriptor)
	slices.SortFunc(a.AssetDescriptorList, func(a, b *AssetDescriptor) bool {
		return strings.Compare(a.Name, b.Name) == -1
	})
}

type saveableAssetContainer struct {
	Type  string
	Inner interface{}
}

func (a *assetManagerImpl) Save(path string, toSave Asset) error {
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
	if !strings.HasSuffix(path, ".json") {
		path = path + ".json"
	}
	a.AssetToLoadPath[toSave] = path
	return assetManager.WriteFS.WriteFile(path, data)
}

type assetLoadPath struct {
	Type string
	Path string
}

func (a *assetManagerImpl) buildJsonToSave(obj any) any {
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
		{
			m := map[string]any{}
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				if !field.IsExported() {
					continue
				}
				m[field.Name] = a.buildJsonToSave(reflect.ValueOf(obj).Field(i).Interface())
			}
			return m
		}
	default:
		return obj
	}
}

func (a *assetManagerImpl) Load(assetPath string) (Asset, error) {
	data, err := assetManager.ReadFile(assetPath)
	if err != nil {
		return nil, err
	}

	container := assetContainer{}
	err = json.Unmarshal(data, &container)

	assetDescriptor, ok := assetManager.AssetDescriptors[container.Type]
	if !ok {
		return nil, fmt.Errorf("Unknown asset '%s' - is type registered?", container.Type)
	}
	obj, err := assetDescriptor.Create()

	if err != nil {
		return nil, err
	}

	_, TType := ObjectTypeName(obj)
	//println("TType ", TType)
	if TType != container.Type {
		return nil, fmt.Errorf("Load type mismatch.  Wanted %s, loaded %s", TType, container.Type)
	}

	err = json.Unmarshal(container.Inner, obj)
	//fmt.Printf("%v %#v\n", reflect.TypeOf(obj).Name(), obj)
	if postLoad, ok := obj.(PostLoadingAsset); ok {
		postLoad.PostLoad()
	}
	return obj, err
}

func (a *assetManagerImpl) unmarshalFromAny(data any, v any) error {
	t := reflect.TypeOf(v)
	fmt.Printf("t %s\n", t.String())
	// deref the pointer
	t = t.Elem()
	fmt.Printf("*t %s\n", t.String())

	val := reflect.ValueOf(v).Elem()
	fmt.Printf("val %s\n", val.String())
	if !val.CanSet() {
		panic("val not settable")
	}
	switch t.Kind() {
	case reflect.Struct:
		rmap := data.(map[string]interface{})
		for i := 0; i < t.NumField(); i++ {
			name := t.Field(i).Name
			fmt.Printf("name %s\n", name)
			mapVal := rmap[name]
			fmt.Printf("mapVal %s\n", mapVal)

			vField := val.Field(i)
			switch vField.Kind() {
			case reflect.Pointer:
				// There will be a serialized assetLoadPath in mapVal
				lp := mapVal.(map[string]any)
				path := lp["Path"].(string)
				asset := a.LoadPathToAsset[path]
				fmt.Printf("Pointer %#v\n", mapVal)
				fmt.Printf("asset %#v\n", asset)
				vField.Set(reflect.ValueOf(asset))

			case reflect.Struct:
				target := val.Field(i).Addr().Interface()
				fmt.Printf("target %#v\n", target)
				a.unmarshalFromAny(mapVal, target)

			case reflect.String:
				vField.SetString(mapVal.(string))
			default:
				fmt.Printf("Default %#v %#v\n", vField, mapVal)
			}
		}
	default:
		dataVal := reflect.ValueOf(data)
		fmt.Printf("dataVal %s\n", dataVal)
		reflect.ValueOf(v).Set(dataVal)
	}
	return nil
}
