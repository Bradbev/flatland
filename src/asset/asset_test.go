package asset_test

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"reflect"
	"testing"

	"github.com/bradbev/flatland/src/asset"

	"github.com/psanford/memfs"
	"github.com/stretchr/testify/assert"
)

type testAsset struct {
	Anykey         string
	postLoadCalled bool
}

func (t *testAsset) PostLoad() {
	t.postLoadCalled = true
}

func TestAssetLoad(t *testing.T) {
	data := `{
	"Type": "github.com/bradbev/flatland/src/asset_test.testAsset",
	"Inner": {
		"Anykey": "hi"
	}
}`
	testAssetLoad := func(callback func()) {
		asset.Reset()
		name := asset.Path("asset.json")
		callback()
		newTestAsset(name, data)

		a, err := asset.Load(name)
		assert.NoError(t, err)
		toTest, ok := a.(*testAsset)
		assert.True(t, ok, "Unable to convert %v to testType", a)
		assert.Equal(t, toTest.Anykey, "hi")
		assert.True(t, toTest.postLoadCalled, "PostLoad was not called")
	}
	testAssetLoad(func() {
		t.Log("Testing RegisterAsset")
		asset.RegisterAsset(testAsset{})
	})
	testAssetLoad(func() {
		t.Log("Testing RegisterAssetFactory")
		asset.RegisterAssetFactory(testAsset{}, func() (asset.Asset, error) { return &testAsset{}, nil })
	})
}

func newTestAsset(name asset.Path, data string) {
	rootFS := memfs.New()
	rootFS.WriteFile(string(name), []byte(data), 0777)
	asset.RegisterFileSystem(rootFS, 0)
}

type writeFS struct {
	fs *memfs.FS
}

func newWriteFS() *writeFS {
	return &writeFS{
		fs: memfs.New(),
	}
}

func (f *writeFS) WriteFile(path asset.Path, data []byte) error {
	return f.fs.WriteFile(string(path), data, 0777)
}

type js = map[string]any

func TestAssetSave(t *testing.T) {
	asset.Reset()
	wfs := newWriteFS()
	asset.RegisterWritableFileSystem(wfs)
	asset.RegisterAsset(testAsset{})

	name := asset.Path("asset.json")
	a := &testAsset{Anykey: "saved"}
	err := asset.Save(name, a)
	assert.NoError(t, err)

	back, err := fs.ReadFile(wfs.fs, string(name))
	assert.NoError(t, err)

	jsonBack := js{}
	err = json.Unmarshal(back, &jsonBack)
	assert.NoError(t, err)

	expected := js{
		"Type":   "github.com/bradbev/flatland/src/asset_test.testAsset",
		"Parent": "",
		"Inner": js{
			"Anykey": "saved",
		},
	}
	assert.Equal(t, expected, jsonBack)

	testLoad := func(postLoadState bool, msg string) {
		asset.RegisterFileSystem(wfs.fs, 0)
		loaded, err := asset.Load(name)
		assert.NoError(t, err)

		expectedObj := &testAsset{Anykey: "saved", postLoadCalled: postLoadState}
		assert.Equal(t, expectedObj, loaded)
	}
	testLoad(false, "Expected post load to be false because we have only saved the asset and kept hold of the orignal object")
	asset.Reset()
	asset.RegisterAsset(testAsset{})
	testLoad(true, "Expected post load to be true because we reset the asset package")
}

func TestAssetList(t *testing.T) {
	asset.Reset()
	asset.RegisterAsset(testAsset{})

	expected := "github.com/bradbev/flatland/src/asset_test.testAsset"
	assets := asset.GetAssetDescriptors()
	assert.Equal(t, expected, assets[0].FullName, "Names don't match")
}
func TestAssetCreate(t *testing.T) {
	asset.Reset()
	asset.RegisterAsset(testAsset{})

	allAssets := asset.GetAssetDescriptors()

	obj, err := allAssets[0].Create()
	assert.NoError(t, err)

	_, ok := obj.(*testAsset)

	assert.True(t, ok)

}

type testAssetNode struct {
	Name      string
	Inline    testAssetLeaf
	Reference *testAssetLeaf
}

type testAssetLeaf struct {
	SecondName string
}

/*
Linked Assets
  - Value types are saved inline
  - Pointers and interfaces are checked to see if the asset has already been
    loaded.  If it is an existing asset, the path will be saved.
  - Pointers that are not asset-loaded are discarded, ie - transient
*/

// TestAssetSaveLinked is primarily about ensuring that testAssetNode
// is saved such that node.Reference is an object that contains a path
// to load, ie - not serialized as if it were inline.
func TestAssetSaveLinked(t *testing.T) {
	asset.Reset()
	rootFS := newWriteFS()
	asset.RegisterFileSystem(rootFS.fs, 0)
	asset.RegisterWritableFileSystem(rootFS)

	asset.RegisterAsset(testAssetNode{})
	asset.RegisterAsset(testAssetLeaf{})

	leaf := &testAssetLeaf{SecondName: "Leaf"}
	err := asset.Save("leaf.json", leaf)
	assert.NoError(t, err)
	node := &testAssetNode{
		Name:      "node",
		Inline:    testAssetLeaf{SecondName: "inline"},
		Reference: leaf,
	}
	err = asset.Save("node.json", node)
	assert.NoError(t, err)

	bytesBack, err := fs.ReadFile(rootFS.fs, "node.json")
	assert.NoError(t, err)
	expected :=
		`{
  "Type": "github.com/bradbev/flatland/src/asset_test.testAssetNode",
  "Parent": "",
  "Inner": {
    "Inline": {
      "SecondName": "inline"
    },
    "Name": "node",
    "Reference": {
      "Type": "github.com/bradbev/flatland/src/asset_test.testAssetLeaf",
      "Path": "leaf.json"
    }
  }
}`
	assert.Equal(t, expected, string(bytesBack))

	bytesBack, err = fs.ReadFile(rootFS.fs, "leaf.json")
	leafExpected :=
		`{
  "Type": "github.com/bradbev/flatland/src/asset_test.testAssetLeaf",
  "Parent": "",
  "Inner": {
    "SecondName": "Leaf"
  }
}`
	assert.Equal(t, leafExpected, string(bytesBack))

}

// Loading needs to unmarshal Json into a generic map
// and iterate the members of a struct.  Pointers and
// interface members cannot be directly loaded, instead
// they must be inspected in the json and loaded from
// disk, then replaced in the map with the real value
// Finally the general map needs to be loaded into the
// target structure.
func TestAssetLoadLinked(t *testing.T) {
	asset.Reset()
	rootFS := newWriteFS()
	asset.RegisterFileSystem(rootFS.fs, 0)
	asset.RegisterWritableFileSystem(rootFS)

	asset.RegisterAsset(testAssetNode{})
	asset.RegisterAsset(testAssetLeaf{})

	leafRaw := `{
		"Type": "github.com/bradbev/flatland/src/asset_test.testAssetLeaf",
		"Inner": {
		  "SecondName": "Leaf"
		}
	  }`
	err := rootFS.WriteFile("leaf.json", []byte(leafRaw))
	assert.NoError(t, err)

	nodeRaw := `{
		"Type": "github.com/bradbev/flatland/src/asset_test.testAssetNode",
		"Inner": {
		  "Inline": {
			"SecondName": "inline"
		  },
		  "Name": "node",
		  "Reference": {
			"Type": "github.com/bradbev/flatland/src/asset_test.testAssetLeaf",
			"Path": "leaf.json"
		  }
		}
	  }`
	err = rootFS.WriteFile("node.json", []byte(nodeRaw))
	assert.NoError(t, err)

	nodeAsset, err := asset.Load("node.json")
	assert.NoError(t, err)

	node := nodeAsset.(*testAssetNode)

	expected := testAssetLeaf{SecondName: "Leaf"}
	assert.Equal(t, expected, *node.Reference, "Reference needs to be loaded from disk, not default")
}

type Float float32
type MyBool bool
type Ints []int
type Bytes []byte
type testAssetPath struct {
	Path asset.Path
	Flt  Float
	Bool MyBool
	B    Bytes
	I    Ints
}

func TestAliasedTypes(t *testing.T) {
	rootFS := newWriteFS()
	reset := func() {
		asset.Reset()
		asset.RegisterFileSystem(rootFS.fs, 0)
		asset.RegisterWritableFileSystem(rootFS)
		asset.RegisterAsset(testAssetPath{})
	}

	reset()
	toSave := &testAssetPath{
		Path: "pathType",
		Flt:  5,
		Bool: true,
		I:    Ints{4, 5, 6},
		B:    Bytes("Bytes I Saved"),
	}
	err := asset.Save("pathtest.json", toSave)
	assert.NoError(t, err)

	//b, _ := fs.ReadFile(rootFS.fs, "pathtest.json")
	//fmt.Printf("%v", string(b))
	//t.Fail()

	reset()
	loaded, err := asset.Load("pathtest.json")
	assert.NoError(t, err)

	assert.Equal(t, toSave, loaded)
}

type reflectStr struct {
	Foo string
}

func TestReflectCopy(t *testing.T) {
	p := fmt.Printf
	a := &reflectStr{"This is A"}
	wrappedA := any(a) // interface of (*reflectStr, a)

	typeToMake := reflect.TypeOf(wrappedA).Elem() // type is reflectStr
	p("typeToMake %v\n", typeToMake)

	b := reflect.New(typeToMake) // returns a *typeToMake
	b.Elem().Set(reflect.ValueOf(wrappedA).Elem())

	assert.Equal(t, a, b.Interface()) // they're equal!

	a.Foo = "bar"
	assert.NotEqual(t, a, b.Interface())
}

type testAssetParent struct {
	StrA string
	StrB string
}

func TestParentLoadingSavingSetting(t *testing.T) {
	rootFS := newWriteFS()
	reset := func() {
		asset.Reset()
		asset.RegisterFileSystem(rootFS.fs, 0)
		asset.RegisterWritableFileSystem(rootFS)
		asset.RegisterAsset(testAssetParent{})
	}
	reset()

	{ // save to the FS
		toSave := &testAssetParent{
			StrA: "ParentA",
			StrB: "ParentB",
		}
		err := asset.Save("parent.json", toSave)
		assert.NoError(t, err)

		err = asset.Save("child.json", &testAssetParent{})
		assert.NoError(t, err)
	}

	{ // load them both
		loadedParent, err := asset.Load("parent.json")
		assert.NoError(t, err)
		parent := loadedParent.(*testAssetParent)
		assert.Equal(t, &testAssetParent{"ParentA", "ParentB"}, parent)

		loadedChild, err := asset.Load("child.json")
		assert.NoError(t, err)
		child := loadedChild.(*testAssetParent)
		assert.Equal(t, &testAssetParent{}, child)

		parent.StrB = "ParentNewB"
		child.StrA = "ChildA"
		err = asset.SetParent(child, parent)
		assert.NoError(t, err)
		assert.Equal(t, &testAssetParent{"ChildA", "ParentNewB"}, child, "Setting the parent will immediately fix the child")
		asset.Save("child.json", child)

		parent.StrB = "ThirdB"
		loadedAgain, err := asset.LoadWithOptions("child.json", asset.LoadOptions{ForceReload: true})
		assert.NoError(t, err)
		assert.Equal(t, loadedChild, loadedAgain, "The already loaded asset is returned")
		assert.Equal(t, &testAssetParent{"ChildA", "ThirdB"}, child, "The asset is modified in place")

		parent.StrB = "Fourth"
		asset.Save("parent.json", parent)
		assert.Equal(t, &testAssetParent{"ChildA", "Fourth"}, child, "Child assets are updated in memory when parents are saved")

	}

	reset()
	{
		loadedChild, err := asset.Load("child.json")
		assert.NoError(t, err)
		child := loadedChild.(*testAssetParent)
		assert.Equal(t, &testAssetParent{"ChildA", "Fourth"}, child)
	}

}

type inlineInnerSaveStruct struct {
	ItemA string
	ItemB string
}

type inlineSaving struct {
	WillNotSave         *inlineInnerSaveStruct
	SaveInline          *inlineInnerSaveStruct `flat:"inline"`
	SaveInterfaceInline any                    `flat:"inline"`
}

func TestInlineAssetSaving(t *testing.T) {
	rootFS := newWriteFS()
	reset := func() {
		asset.Reset()
		asset.RegisterFileSystem(rootFS.fs, 0)
		asset.RegisterWritableFileSystem(rootFS)
		asset.RegisterAsset(inlineInnerSaveStruct{})
		asset.RegisterAsset(inlineSaving{})
	}
	reset()

	parent := &inlineInnerSaveStruct{"ParentA", "ParentB"}
	asset.Save("parent.json", parent)

	anyS := any(&inlineInnerSaveStruct{"IfaceA", "IfaceB"})
	toSave := &inlineSaving{
		WillNotSave:         &inlineInnerSaveStruct{"ignoreA", "ignoreB"},
		SaveInline:          &inlineInnerSaveStruct{"InlineA", "ParentB"},
		SaveInterfaceInline: anyS,
	}

	asset.SetParent(toSave.SaveInline, parent)

	err := asset.Save("inlined.json", toSave)
	assert.NoError(t, err)
	{
		data, _ := fs.ReadFile(rootFS.fs, "inlined.json")
		fmt.Printf("%s\n", string(data))
	}

	reset()

	loaded, err := asset.Load("inlined.json")
	assert.NoError(t, err)
	expected := &inlineSaving{
		SaveInline: &inlineInnerSaveStruct{"InlineA", "ParentB"},
	}
	assert.Equal(t, expected, loaded)

	exactFileContent := `{
  "Type": "github.com/bradbev/flatland/src/asset_test.inlineSaving",
  "Parent": "",
  "Inner": {
    "SaveInline": {
      "Type": "github.com/bradbev/flatland/src/asset_test.inlineInnerSaveStruct",
      "Parent": "parent.json",
      "Inner": {
        "ItemA": "InlineA"
      }
    }
  }
}`
	data, err := fs.ReadFile(rootFS.fs, "inlined.json")
	assert.Equal(t, exactFileContent, string(data), "The inlined field has a parent etc")
}
