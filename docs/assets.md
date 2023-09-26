# Assets
Assets are the data building blocks of `flat` games.  The editor packages
create and edit assets.  Assets are just plain Go structs are are registered
with the `asset` package:
```go
type BasicAsset struct {
	Name string
}
asset.RegisterAsset(BasicAsset{})
```
*NOTE* It might be better to make `RegisterAsset` a generic function.

# Serialization and editing
The editor relies heavily on `reflect` and other packages that use reflect,
like `json`.  The Go convention is that unexported fields are ignored for
saving and loading.  `flat` follows this convention, bringing us to:

## Rule 1 - Unexported fields are ignored
The editor will not edit, save or load fields that are unexported.  There are
many good reasons you might want to edit or save assets that have unexported
fields, but that requires extra effort by the asset author (ie, custom editor,
custom marshaling functions)

## Rule 2 - `json` is the main serializer
The `json` package is used to save and load assets, so all the rules of that
package apply to `flat` assets.
A struct with nested structs will be saved inline, ie
```go
type InnerStruct struct {
	InnerName string
}
type Outer struct {
	Name string
	Inner InnerStruct
}
```
will save as
```json
{
	"Name": "name",
	"Inner": { 
		"InnerName": "innername" 
	}
}
```

## Rule 3 - Pointers and Interfaces to assets are saved as paths
Consider the asset
```go
type HasReference struct {
	Other *OtherAsset
}
```
In this case, `OtherAsset` is a registered asset that is saved on disk as
"otherasset.json".  Normally the `json` serialization would save the values of
`Other` directly inline, but `asset` serialization is different.  Instead, a
private structure is saved into the json file that contains a reference to the
path "otherasset.json".  At load time `asset.Load("otherasset.json")` will be
called to provide the pointer for `Other`.  
> ### TODO
> * What is true for pointers is true for interfaces.
> * A pointer to a known asset type that is not saved on disk will be serialized inline

## Rule 4 - Assets are singletons
Assets should be loaded with the `asset.Load(path)` function.  This function 
will load an asset from disk, or return a pointer to the already loaded asset.  
This is a very useful property to allow live-tuning of assets.  Code should 
take care when changing the values inside an asset at runtime.

## Rule 5 - `inline` values are new instances


# Parent Values for Assets
Assets may take their default values from another Asset of the *same type*.  This allows for some basic reuse of data values, with the following rules.
1. Assets in memory have the full set of data, there is no connection to their parent at runtime.  In other words, you can change parent values in memory and no already loaded child assets will change.
2. Parent values are applied at Load time, using the values of the in-memory parent asset.  No extra reloading from disk is done.
3. Assets are saved sparsely, only the fields that differ from their parent's field are saved.
4. When an asset is loaded, the zero object for that asset is created, the parent object is copied into the child field by field and then the sparse save data is loaded into the child object.

