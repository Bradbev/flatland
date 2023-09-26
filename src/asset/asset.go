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
	"fmt"
	"io/fs"
	systemLog "log"
	"os"
	"reflect"
	"strings"
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

// SetEditorMode informs the asset package that it is being used in an
// editor context.  For example when enabled extra meta data about
// which fields a child asset overrides must be saved.
func SetEditorMode() {
	assetManager.EditorMode = true
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

// SetParent is used to set the parent of an Asset.
// When an Asset is reparented, all values that are not overridden by the child
// are copied in from the parent.  If there is no previous parent then the parent
// and the child are diffed and in places where they differ the child will override
// the parent.
func SetParent(child Asset, parent Asset) error {
	return assetManager.SetParent(child, parent)
}

func GetParent(child Asset) Path {
	return assetManager.GetParent(child)
}

func GetLoadPathForAsset(a Asset) (Path, error) {
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

func ResetForTest() {
	assetManager = newAssetManagerImpl()
	assetManager.EditorMode = true
}

func GetAssetDescriptors() []*AssetDescriptor {
	return assetManager.AssetDescriptorList
}

func GetDescriptorForAsset(asset Asset) *AssetDescriptor {
	_, typeName := ObjectTypeName(asset)
	return assetManager.AssetDescriptors[typeName]
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

func GetFlatTag(sf *reflect.StructField, key string) (value string, exists bool) {
	tags, ok := sf.Tag.Lookup("flat")
	if !ok {
		return "", false
	}

	parts := strings.Split(tags, ";")
	for _, part := range parts {
		keyValue := strings.Split(part, ":")
		keyPart := strings.Trim(keyValue[0], " ")
		if keyPart == key {
			if len(keyValue) > 1 {
				return strings.Trim(keyValue[1], " "), true
			}
			return "", true
		}
	}
	return "", false
}
