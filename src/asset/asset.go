/*
* Assets on disk have the same format - assetContainer.  This struct holds
* some meta data about the asset and then an Inner that holds the user
* defined data.
*
* The Rules of Assets
* Every Asset is a singleton.  When an Asset is in memory there is only one copy of
* it.  This means that it is easy to tune live assets, but also means that you cannot
* use Assets as a way to store runtime data.  Runtime data should be stored and updated
* in Actors.
*
* Pointers to other Assets within an Assets are treated specially.  The Asset loading system
* will load the dependency.  This means that loading a single Asset may result in a load
* chain that loads nearly everything.
 */

package asset

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	systemLog "log"
	"os"
	"reflect"
	"strings"

	"golang.org/x/exp/slices"
)

var log = systemLog.New(os.Stderr, "Asset", systemLog.Ltime)

// Path is a distinct type from string so that the editor package
// can present an autocomplete window.
// The editor also understands the `filter` tag when used with Path,
// for example
//
//	struct {
//	  P Path `filter:"png,jpg"`
//	}
//
// will only show files that contain the text 'png' or 'jpg'
type Path string

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

// NamedAsset allows assets to provide a different Name
// The editor will use this Name instead of the struct name
type NamedAsset interface {
	Name() string
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

func ReadFile(assetPath Path) ([]byte, error) {
	return assetManager.ReadFile(assetPath)
}

func Load(assetPath Path) (Asset, error) {
	return assetManager.Load(assetPath)
}

func Save(path Path, toSave Asset) error {
	return assetManager.Save(path, toSave)
}

// WalkFiles is like fs.WalkDir, but it will walk all the readable file systems
// registered with asset.RegisterFileSystem
func WalkFiles(fn fs.WalkDirFunc) error {
	return assetManager.WalkFiles(fn)
}

func Reset() {
	assetManager = newAssetManagerImpl()
}

func GetAssetDescriptors() []*AssetDescriptor {
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
	AssetToLoadPath map[Asset]Path

	// Maps a Path to an already loaded asset
	LoadPathToAsset map[Path]Asset
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
		AssetToLoadPath:  map[Asset]Path{},
		LoadPathToAsset:  map[Path]Asset{},
	}
}

func (a *assetManagerImpl) AddFS(wrapper *fsWrapper) error {
	a.FileSystems = append(a.FileSystems, wrapper)
	slices.SortFunc(a.FileSystems, func(a, b *fsWrapper) bool {
		return a.Priority < b.Priority
	})
	return nil
}

func (a *assetManagerImpl) ReadFile(path Path) ([]byte, error) {
	for _, fsys := range a.FileSystems {
		data, err := fs.ReadFile(fsys.FileSystem, string(path))
		if err == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("Unable to find path (%s) in any registered FS ", path)
}

func (a *assetManagerImpl) WalkFiles(fn fs.WalkDirFunc) error {
	var e error
	for _, fsys := range a.FileSystems {
		err := fs.WalkDir(fsys.FileSystem, ".", func(path string, d fs.DirEntry, err error) error {
			e = fn(path, d, err)
			return e
		})
		if err != nil {
			return err
		}
		// if fn has requested SkipAll, then we early out
		if e == fs.SkipAll {
			return nil
		}
	}
	return e
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

func (a *assetManagerImpl) Load(assetPath Path) (Asset, error) {
	if asset, loaded := a.LoadPathToAsset[assetPath]; loaded {
		return asset, nil
	}

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

	var anyInner any
	err = json.Unmarshal(container.Inner, &anyInner)
	if err != nil {
		return nil, err
	}

	err = a.unmarshalFromAny(anyInner, obj)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("%v %#v\n", reflect.TypeOf(obj).Name(), obj)
	if postLoad, ok := obj.(PostLoadingAsset); ok {
		postLoad.PostLoad()
	}
	if err == nil {
		// save the references to these assets to prevent future loading
		a.AssetToLoadPath[obj] = assetPath
		a.LoadPathToAsset[assetPath] = obj
	}
	return obj, err
}

func (a *assetManagerImpl) unmarshalFromAny(data any, v any) error {
	return a.unmarshalFromValues(reflect.ValueOf(data), reflect.ValueOf(v).Elem())
}

func (a *assetManagerImpl) unmarshalFromValues(data reflect.Value, v reflect.Value) error {
	//fmt.Printf("v:%#v settable? %v kind %s\n", v, v.CanSet(), v.Kind())
	//fmt.Printf("data:%#v kind %s\n", data, data.Kind())
	t := v.Type()
	switch v.Kind() {
	case reflect.Pointer:
		fmt.Printf("Handle ptr")
		// There will be a serialized assetLoadPath in data
		lp := data.Interface().(map[string]any)
		path := lp["Path"].(string)
		asset, err := a.Load(Path(path))
		fmt.Printf("Pointer %#v\n", data)
		fmt.Printf("asset %#v\n", asset)
		if err != nil {
			return fmt.Errorf("unable to load asset at path %s", path)
		}
		v.Set(reflect.ValueOf(asset))

	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if !t.Field(i).IsExported() {
				continue
			}
			fieldToSet := v.Field(i)
			key := reflect.ValueOf(t.Field(i).Name)
			dataToRead := data.MapIndex(key)
			//fmt.Printf("Key %v Data %v\n", key, dataToRead)

			if dataToRead.Kind() == reflect.Invalid {
				log.Printf("dataToRead for key (%s) is missing, skipping", key)
				continue
			}
			a.unmarshalFromValues(dataToRead.Elem(), fieldToSet)
		}
	case reflect.Slice:
		v.Set(reflect.MakeSlice(v.Type(), data.Len(), data.Len()))
		fallthrough
	case reflect.Array:
		if data.Len() > 0 && v.Index(0).Kind() == reflect.Uint8 {
			// byte slices are uuencoded into a string
			encoded := data.String()
			decoded, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				return err
			}
			//fmt.Printf("%v %v\n", decoded, string(decoded))
			v.SetBytes(decoded)
		} else {
			for i := 0; i < data.Len(); i++ {
				indexToSet := v.Index(i)
				dataToRead := data.Index(i)
				//fmt.Printf("Array Data %v\n", dataToRead)
				a.unmarshalFromValues(dataToRead.Elem(), indexToSet)
			}
		}
	default:
		if data.CanFloat() {
			// Json treats all numerics as floats, so we must handle float
			// to int conversion
			if v.CanFloat() {
				v.SetFloat(data.Float())
			} else if v.CanInt() {
				v.SetInt(int64(data.Float()))
			} else if v.CanUint() {
				v.SetUint(uint64(data.Float()))
			}
		} else if data.Kind() == reflect.String {
			// the asset.Path type is a string, but we can't just do
			// v.Set, instead we need to use SetString.  Other types
			// that are aliased like this might also break
			v.SetString(data.String())
		} else if data.Kind() == reflect.Bool {
			v.SetBool(data.Bool())
		} else {
			v.Set(data)
		}
	}

	return nil
}
