package asset

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
)

func (a *assetManagerImpl) Load(assetPath Path) (Asset, error) {
	return a.LoadWithOptions(assetPath, LoadOptions{})
}

func (a *assetManagerImpl) NewInstance(assetToInstance Asset) (Asset, error) {
	descriptor := a.GetAssetDescriptor(assetToInstance)
	if descriptor == nil {
		return nil, fmt.Errorf("Unable to find descriptor for asset %v", assetToInstance)
	}

	concreteOrigin := reflect.ValueOf(assetToInstance).Elem().Interface()
	commonFormat := a.toCommonFormat(concreteOrigin)
	instance, err := a.loadFromCommonFormat(descriptor.FullName, "", commonFormat, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to create instance from %v", descriptor)
	}
	return instance, nil
}

func (a *assetManagerImpl) LoadWithOptions(assetPath Path, options LoadOptions) (Asset, error) {
	// If we are able, don't reload an existing asset
	alreadyLoadedAsset, loaded := a.LoadPathToAsset[assetPath]
	if loaded && !options.ForceReload && !options.createInstance {
		return alreadyLoadedAsset, nil
	}

	if options.createInstance {
		alreadyLoadedAsset = nil
	}

	data, err := assetManager.ReadFile(assetPath)
	if err != nil {
		return nil, err
	}

	// load the on disk format and validate things
	container := onDiskLoadFormat{}
	err = json.Unmarshal(data, &container)
	if err != nil {
		return nil, err
	}

	loadedAsset, err := a.loadFromOnDiskLoadFormat(&container, alreadyLoadedAsset)
	if err != nil {
		return nil, err
	}

	if err == nil && !options.createInstance {
		// save the references to these assets to prevent future loading
		a.AssetToLoadPath[loadedAsset] = assetPath
		a.LoadPathToAsset[assetPath] = loadedAsset
	}
	return loadedAsset, nil
}

func (a *assetManagerImpl) loadFromOnDiskLoadFormat(container *onDiskLoadFormat, alreadyLoadedAsset Asset) (Asset, error) {
	var commonFormat any
	err := json.Unmarshal(container.Inner, &commonFormat)
	if err != nil {
		return nil, err
	}
	parentPath := container.Parent
	if parentPath == "" {
		parentPath = a.ChildToParent[alreadyLoadedAsset]
	}
	return a.loadFromCommonFormat(container.Type, parentPath, commonFormat, alreadyLoadedAsset)
}

func (a *assetManagerImpl) loadFromCommonFormat(
	assetType string,
	parentPath Path,
	commonFormat any,
	alreadyLoadedAsset Asset) (Asset, error) {

	assetDescriptor, ok := assetManager.AssetDescriptors[assetType]
	if !ok {
		return nil, fmt.Errorf("Unknown asset '%s' - is type registered?", assetType)
	}

	assetToLoadInto := alreadyLoadedAsset
	if assetToLoadInto == nil {
		var err error
		assetToLoadInto, err = assetDescriptor.Create()
		if err != nil {
			return nil, err
		}
	}

	_, TType := ObjectTypeName(assetToLoadInto)
	//println("TType ", TType)
	if TType != assetType {
		return nil, fmt.Errorf("Load type mismatch.  Wanted %s, loaded %s", TType, assetType)
	}

	{ // copy the parent into the child
		// load the parent (if it has one) and copy into the child
		if parentPath != "" {
			parent, err := a.Load(parentPath)
			if err != nil {
				return nil, err
			}
			// we don't want the common format for a pointer to the parent, but
			// to the real struct
			parentConcrete := reflect.ValueOf(parent).Elem().Interface()
			parentInCommonFormat := a.toCommonFormat(parentConcrete)
			a.loadFromCommonFormat(assetType, "", parentInCommonFormat, assetToLoadInto)

			// set the parent for this asset
			a.ChildToParent[assetToLoadInto] = parentPath
		}
	}

	var err error
	// commonFormat can be nil - it means the whole object is default/inherited
	if commonFormat != nil {
		err = a.unmarshalCommonFormat(commonFormat, assetToLoadInto)
		if err != nil {
			return nil, err
		}
	}
	//fmt.Printf("%v %#v\n", reflect.TypeOf(obj).Name(), obj)
	if postLoad, ok := assetToLoadInto.(PostLoadingAsset); ok {
		postLoad.PostLoad()
	}

	return assetToLoadInto, err
}

func (a *assetManagerImpl) unmarshalCommonFormat(data any, v any) error {
	return a.unmarshalCommonFormatFromValues(reflect.ValueOf(data), reflect.ValueOf(v).Elem())
}

func safeLen(value reflect.Value) int {
	if value.Kind() == reflect.String {
		return value.Len()
	}
	if !value.IsValid() || value.IsZero() || value.IsNil() {
		return 0
	}
	return value.Len()
}

// unmarshalCommonFormatFromValues accepts a source Value in Common format and a concrete
// Go type in dest.
// Common Format erases some type information about structs, so dest is reflected to recover types.
func (a *assetManagerImpl) unmarshalCommonFormatFromValues(source reflect.Value, dest reflect.Value) error {
	//fmt.Printf("source:%#v \nsettable? %v \nkind %s\n", source, source.CanSet(), source.Kind())
	//fmt.Printf("dest:%#v \nkind %s\n-----\n", dest, dest.Kind())
	t := dest.Type()
	switch dest.Kind() {
	case reflect.Interface:
		//fmt.Println("Handle interface (fallthrough to pointer)")
		fallthrough
	case reflect.Pointer:
		// There will be a serialized assetLoadPath in data, or a savedAssetContainer
		if !source.IsValid() {
			return fmt.Errorf("attempt to load pointer, but not valid")
		}
		if source.IsNil() {
			return fmt.Errorf("attempt to load pointer, but is nil")
		}

		// this is a quirk - normal load would be returning a map[string]any because it's
		// come via the json loader.  However if we got to the common format internally (toCommonFormat)
		// then it'll be *mostly* common format, but the onDiskLoadFormat structs will still be native.
		var diskLoadFormat onDiskLoadFormat
		diskSaveFormat, isDiskFormat := source.Interface().(*onDiskSaveFormat)
		if isDiskFormat {
			b, _ := json.Marshal(diskSaveFormat)
			json.Unmarshal(b, &diskLoadFormat)
		}

		loadPathInfo, isLoadPath := source.Interface().(map[string]any)
		if !isLoadPath && !isDiskFormat {
			return fmt.Errorf("unable to cast %v to map[string]any, OR to onDiskLoadFormat", source)
		}

		if isLoadPath {
			// Paths must load first
			if pathAny, ok := loadPathInfo["Path"]; ok {
				// This is a saved reference to another asset
				path := pathAny.(string)
				asset, err := a.Load(Path(path))
				if err != nil {
					return fmt.Errorf("unable to load asset at path %s", path)
				}
				dest.Set(reflect.ValueOf(asset))
				return nil
			}

			if _, ok := loadPathInfo["Type"]; ok {
				d, _ := json.Marshal(source.Interface())
				json.Unmarshal(d, &diskLoadFormat)
				isDiskFormat = true
			}
		}

		if isDiskFormat {
			inlineAsset, err := a.loadFromOnDiskLoadFormat(&diskLoadFormat, nil)
			if err != nil {
				return fmt.Errorf("unable to load inline asset")
			}
			dest.Set(reflect.ValueOf(inlineAsset))
			return nil
		}

		panic("Should not get here")

	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if !t.Field(i).IsExported() {
				continue
			}
			fieldToSet := dest.Field(i)
			name := t.Field(i).Name
			key := reflect.ValueOf(name)
			dataToRead := source.MapIndex(key)

			if dataToRead.Kind() == reflect.Invalid {
				//log.Printf("dataToRead for key (%s) is missing, skipping", key)
				continue
			}
			a.unmarshalCommonFormatFromValues(dataToRead.Elem(), fieldToSet)
		}
	case reflect.Slice:
		l := safeLen(source)
		dest.Set(reflect.MakeSlice(dest.Type(), l, l))
		fallthrough
	case reflect.Array:
		l := safeLen(source)
		if l > 0 && dest.Index(0).Kind() == reflect.Uint8 {
			// byte slices are uuencoded into a string
			encoded := source.String()
			decoded, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				return err
			}
			//fmt.Printf("%v %v\n", decoded, string(decoded))
			dest.SetBytes(decoded)
		} else {
			for i := 0; i < l; i++ {
				indexToSet := dest.Index(i)
				dataToRead := source.Index(i)
				//fmt.Printf("Array Data %v\n", dataToRead)
				a.unmarshalCommonFormatFromValues(dataToRead.Elem(), indexToSet)
			}
		}
	default:
		if source.CanFloat() {
			// Json treats all numerics as floats, so we must handle float
			// to int conversion
			if dest.CanFloat() {
				dest.SetFloat(source.Float())
			} else if dest.CanInt() {
				dest.SetInt(int64(source.Float()))
			} else if dest.CanUint() {
				dest.SetUint(uint64(source.Float()))
			}
		} else if source.Kind() == reflect.String {
			// the asset.Path type is a string, but we can't just do
			// v.Set, instead we need to use SetString.  Other types
			// that are aliased like this might also break
			dest.SetString(source.String())
		} else if source.Kind() == reflect.Bool {
			dest.SetBool(source.Bool())
		} else {
			dest.Set(source)
		}
	}

	return nil
}
