package asset

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"reflect"
	"strings"

	"golang.org/x/exp/slices"
)

// onDiskLoadFormat must be the same as onDiskSaveFormat, except for the type of Inner
type onDiskLoadFormat struct {
	Type   string
	Parent Path
	Inner  json.RawMessage
}

// onDiskSaveFormat must be the same as onDiskLoadFormat, except for the type of Inner
type onDiskSaveFormat struct {
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

	EditorMode bool

	// ChildAssetOverrides stores metadata about a child asset and
	// the feilds that it overrides.  If an asset is not a child
	// it will not be in this map.
	ChildAssetOverrides map[Asset]*childOverrides
}

type fsWrapper struct {
	FileSystem fs.FS
	Priority   int
}

var assetManager = newAssetManagerImpl()

func newAssetManagerImpl() *assetManagerImpl {
	return &assetManagerImpl{
		FileSystems:         []*fsWrapper{},
		AssetDescriptors:    map[string]*AssetDescriptor{},
		AssetToLoadPath:     map[Asset]Path{},
		LoadPathToAsset:     map[Path]Asset{},
		ChildToParent:       map[Asset]Path{},
		ChildAssetOverrides: map[Asset]*childOverrides{},
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
		container := onDiskLoadFormat{}
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

func (a *assetManagerImpl) GetAssetDescriptor(target Asset) *AssetDescriptor {
	_, typeName := ObjectTypeName(target)
	return a.AssetDescriptors[typeName]
}

// SetParent is used to set the parent of an Asset.
// When an Asset is reparented, all values that are not overridden by the child
// are copied in from the parent.  If there is no previous parent then the parent
// and the child are diffed and in places where they differ the child will override
// the parent.
func (a *assetManagerImpl) SetParent(child Asset, parent Asset) error {
	if !a.EditorMode {
		panic("Must be in editor mode to call SetParent")
	}
	parentPath, ok := a.AssetToLoadPath[parent]
	if !ok {
		return fmt.Errorf("parent is not a loaded asset %v", parent)
	}

	// no parent case
	if _, hasParent := a.ChildToParent[child]; !hasParent {
		diffs := a.findDiffsFromParent(parent, child)
		if diffs != nil {
			overrides := newChildOverrides()
			overrides.BuildFromCommonFormat(diffs)
			a.ChildAssetOverrides[child] = overrides
		}
	}

	a.ChildToParent[child] = parentPath
	return nil
}

type OverrideEnableType uint8

const (
	OverrideEnable OverrideEnableType = iota
	OverrideDisable
)

func SetChildOverrideForField(child Asset, pathToField string, enable OverrideEnableType) error {
	return assetManager.SetChildOverrideForField(child, pathToField, enable)
}

func (a *assetManagerImpl) SetChildOverrideForField(child Asset, pathToField string, enable OverrideEnableType) error {
	overrides := a.ChildAssetOverrides[child]
	if enable == OverrideEnable {
		if overrides == nil {
			overrides = newChildOverrides()
			a.ChildAssetOverrides[child] = overrides
		}
		overrides.AddPath(pathToField)
	}
	if enable == OverrideDisable && overrides != nil {
		overrides.RemovePath(pathToField)
		if overrides.Empty() {
			delete(a.ChildAssetOverrides, child)
		}
		if parentPath, ok := a.ChildToParent[child]; ok {
			a.refreshParentValuesForChild(child, parentPath)
		}
	}
	return nil
}
