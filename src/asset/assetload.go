package asset

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/jinzhu/copier"
)

func (a *assetManagerImpl) Load(assetPath Path) (Asset, error) {
	// Never load an asset twice
	if asset, loaded := a.LoadPathToAsset[assetPath]; loaded {
		return asset, nil
	}

	data, err := assetManager.ReadFile(assetPath)
	if err != nil {
		return nil, err
	}

	// load the generic format and validate things
	container := loadedAssetContainer{}
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

	// load the parent (if it has one) and copy it into the child
	if container.Parent != "" {
		parent, err := a.Load(container.Parent)
		if err != nil {
			return nil, err
		}
		copier.Copy(obj, parent)
	}

	var anyInner any
	err = json.Unmarshal(container.Inner, &anyInner)
	if err != nil {
		return nil, err
	}

	// anyInner can be nil - it means the whole object is default/inherited
	if anyInner != nil {
		err = a.unmarshalFromAny(anyInner, obj)
		if err != nil {
			return nil, err
		}
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

func safeLen(value reflect.Value) int {
	if value.Kind() == reflect.String {
		return value.Len()
	}
	if !value.IsValid() || value.IsZero() || value.IsNil() {
		return 0
	}
	return value.Len()
}

//var assetType = reflect.ValueOf(new(Asset)).Elem().Type()

func (a *assetManagerImpl) unmarshalFromValues(source reflect.Value, dest reflect.Value) error {
	//fmt.Printf("source:%#v \nsettable? %v \nkind %s\n", source, source.CanSet(), source.Kind())
	//fmt.Printf("dest:%#v \nkind %s\n-----\n", dest, dest.Kind())
	t := dest.Type()
	switch dest.Kind() {
	case reflect.Interface:
		//fmt.Println("Handle interface (fallthrough to pointer)")
		fallthrough
	case reflect.Pointer:
		//fmt.Println("Handle ptr")
		// There will be a serialized assetLoadPath in data
		if !source.IsValid() {
			return fmt.Errorf("attempt to load pointer, but not valid")
		}
		if source.IsNil() {
			return fmt.Errorf("attempt to load pointer, but is nil")
		}
		loadPathInfo, ok := source.Interface().(map[string]any)
		if !ok {
			return fmt.Errorf("unable to cast %v to map[string]any", source)
		}
		path := loadPathInfo["Path"].(string)
		asset, err := a.Load(Path(path))
		//fmt.Printf("Disk link %#v\n", data)
		//fmt.Printf("asset %#v\n", asset)
		if err != nil {
			return fmt.Errorf("unable to load asset at path %s", path)
		}
		dest.Set(reflect.ValueOf(asset))

	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if !t.Field(i).IsExported() {
				continue
			}
			fieldToSet := dest.Field(i)
			key := reflect.ValueOf(t.Field(i).Name)
			dataToRead := source.MapIndex(key)
			//fmt.Printf("Key %v Data %v\n", key, dataToRead)

			if dataToRead.Kind() == reflect.Invalid {
				log.Printf("dataToRead for key (%s) is missing, skipping", key)
				continue
			}
			a.unmarshalFromValues(dataToRead.Elem(), fieldToSet)
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
				a.unmarshalFromValues(dataToRead.Elem(), indexToSet)
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

// makePristineManager creates a brand new assetManager and sets it up
// for read-only access to the same data as the existing manager.
// Used to force-load an asset from disk (eg, to compare an unsaved asset)
// BEWARE - this call shares some maps, do not mutate the maps!
func (a *assetManagerImpl) makePristineManager() *assetManagerImpl {
	clean := newAssetManagerImpl()
	clean.FileSystems = a.FileSystems
	clean.AssetDescriptors = a.AssetDescriptors
	clean.AssetDescriptorList = a.AssetDescriptorList
	clean.ChildToParent = a.ChildToParent
	return clean
}
