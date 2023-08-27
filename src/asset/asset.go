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
	"encoding/json"
	"fmt"
	"io/fs"
	systemLog "log"
	"os"
	"reflect"
	"strings"

	"github.com/jinzhu/copier"
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
	Type     reflect.Type
}

// Asset can be any type.
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

// Any type that implements DefaultInitializer will
// have DefaultInitialize called when assets are created.
// Types do not need to be assets for this to work.
type DefaultInitializer interface {
	DefaultInitialize()
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

func RegisterAsset(zeroAsset any) {
	assetManager.RegisterAssetFactory(zeroAsset, func() (Asset, error) {
		zeroType := reflect.TypeOf(zeroAsset)
		zero := reflect.New(zeroType)
		return zero.Interface().(Asset), nil
	})
}

func ReadFile(assetPath Path) ([]byte, error) {
	return assetManager.ReadFile(assetPath)
}

type LoadOptions struct {
	// ForceReload will reload the asset from disk.  If the asset already
	// exists in memory that same object will be reused.
	ForceReload bool

	// createInstance loads the instance in such a way that a new asset is created
	// Pointers to assets in the object will point to existing in-memory assets.  Anything saved
	// inline will be loaded fresh
	createInstance bool
}

func LoadWithOptions(assetPath Path, options LoadOptions) (Asset, error) {
	return assetManager.LoadWithOptions(assetPath, options)
}

func Load(assetPath Path) (Asset, error) {
	return assetManager.Load(assetPath)
}

func NewInstance(a Asset) (Asset, error) {
	return assetManager.NewInstance(a)
}

func Save(path Path, toSave Asset) error {
	return assetManager.Save(path, toSave)
}

func SetParent(child Asset, parent Asset) error {
	return assetManager.SetParent(child, parent)
}

func LoadPathForAsset(a Asset) (Path, error) {
	path, ok := assetManager.AssetToLoadPath[a]
	if !ok {
		return path, fmt.Errorf("Asset not loaded")
	}
	return path, nil
}

// WalkFiles is like fs.WalkDir, but it will walk all the readable file systems
// registered with asset.RegisterFileSystem
func WalkFiles(fn fs.WalkDirFunc) error {
	return assetManager.WalkFiles(fn)
}

func FilterFilesByType[T any]() ([]string, error) {
	typ := reflect.TypeOf((*T)(nil))
	return FilterFilesByReflectType(typ)
}

func FilterFilesByReflectType(typ reflect.Type) ([]string, error) {
	return assetManager.FilterFilesByType(typ)
}

func FilterAssetDescriptorsByType[T any]() []*AssetDescriptor {
	typ := reflect.TypeOf((*T)(nil))
	return FilterAssetDescriptorsByReflectType(typ)
}

func FilterAssetDescriptorsByReflectType(typ reflect.Type) []*AssetDescriptor {
	return assetManager.FilterAssetDescriptorsByType(typ)
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

// loadedAssetContainer must be the same as savedAssetContainer, except for the type of Inner
type loadedAssetContainer struct {
	Type   string
	Parent Path
	Inner  json.RawMessage
}

// savedAssetContainer must be the same as loadedAssetContainer, except for the type of Inner
type savedAssetContainer struct {
	Type   string
	Parent Path
	Inner  interface{}
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

	// ChildToParent maps a child asset to its parent
	ChildToParent map[Asset]Path
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
		ChildToParent:    map[Asset]Path{},
	}
}

func (a *assetManagerImpl) AddFS(wrapper *fsWrapper) error {
	a.FileSystems = append(a.FileSystems, wrapper)
	slices.SortFunc(a.FileSystems, func(a, b *fsWrapper) int {
		return a.Priority - b.Priority
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

// FilterFilesByType will return all the assets that have the exact type as typ.  If typ
// is an interface, return all the files that implement the interface
func (a *assetManagerImpl) FilterFilesByType(typ reflect.Type) ([]string, error) {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	var ret []string
	err := a.WalkFiles(func(path string, d fs.DirEntry, _ error) error {
		if d != nil && d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".json") {
			return nil
		}
		// TODO - unify this with the load snippet in assetload.go
		data, err := assetManager.ReadFile(Path(path))
		if err != nil {
			return err
		}
		container := loadedAssetContainer{}
		err = json.Unmarshal(data, &container)
		if err != nil {
			return nil
		}
		if desc, ok := a.AssetDescriptors[container.Type]; ok {
			if matchesOrImplements(typ, desc.Type) {
				ret = append(ret, path)
			}
		}

		return nil
	})
	return ret, err
}

// returns true if a and b are the same type or b implements a
func matchesOrImplements(a, b reflect.Type) bool {
	if a == b {
		return true
	}
	if a.Kind() == reflect.Interface {
		// if we are matching against an interface we need to use PtrTo
		// because the Type in the descriptor is the real type, not a *T
		ptrTyp := reflect.PtrTo(b)
		return ptrTyp.Implements(a)
	}
	return false
}

// FilterAssetDecriptorsByType will return all the asset descriptors that have the exact type as typ.  If typ
// is an interface, return all the descriptors that implement the interface
func (a *assetManagerImpl) FilterAssetDescriptorsByType(typ reflect.Type) []*AssetDescriptor {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	var ret []*AssetDescriptor
	for _, desc := range a.AssetDescriptorList {
		if matchesOrImplements(typ, desc.Type) {
			ret = append(ret, desc)
		}
	}
	return ret
}

func (a *assetManagerImpl) RegisterAssetFactory(zeroAsset any, factoryFunction FactoryFunc) {
	zeroType := reflect.TypeOf(zeroAsset)
	if zeroType.Kind() != reflect.Struct {
		log.Panicf("RegisterAssetFactory must be called with a concrete type that is a struct.  This is a programming error - %v", zeroAsset)
	}

	// wrap the client provided factoryFunc with one that also initializes structs
	createFunc := func() (Asset, error) {
		a, err := factoryFunction()
		if a != nil {
			callAllDefaultInitializers(a)
		}
		return a, err
	}

	name, typeName := ObjectTypeName(zeroAsset)
	println("Registered asset ", typeName)
	descriptor := &AssetDescriptor{
		Name:     name,
		FullName: typeName,
		Create:   createFunc,
		Type:     zeroType,
	}
	a.AssetDescriptors[typeName] = descriptor
	a.AssetDescriptorList = append(a.AssetDescriptorList, descriptor)
	slices.SortFunc(a.AssetDescriptorList, func(a, b *AssetDescriptor) int {
		return strings.Compare(a.Name, b.Name)
	})
}

// SetParent is used to set the parent of an Asset and also update
// the child with new parent defaults.
func (a *assetManagerImpl) SetParent(child Asset, parent Asset) error {
	parentPath, ok := a.AssetToLoadPath[parent]
	if !ok {
		return fmt.Errorf("parent is not a loaded asset %v", parent)
	}
	// To set a new parent we need to find the diffs between the old parent and the child
	var oldParent any
	if oldParentPath, ok := a.ChildToParent[child]; ok {
		var err error
		oldParent, err = a.Load(oldParentPath)
		if err != nil {
			return err
		}
	}

	a.ChildToParent[child] = parentPath

	// find diffs between the old parent and the child
	diffs := a.findDiffsFromParent(oldParent, child)

	// copy the new parent values into the child
	copier.CopyWithOption(child, parent, copier.Option{DeepCopy: true})
	b, err := json.Marshal(diffs)
	if err != nil {
		return err
	}
	// unmarshal the diffs into the child
	json.Unmarshal(b, child)

	return err
}
