package asset_test

import (
	"encoding/json"
	"flatland/src/asset"
	"fmt"
	"io/fs"
	"testing"

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
	"Type": "flatland/src/asset_test.testAsset",
	"Inner": {
		"anykey": "hi"
	}
}`
	testAssetLoad := func(callback func()) {
		defer asset.Reset()
		name := "asset.json"
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

func newTestAsset(name string, data string) {
	rootFS := memfs.New()
	rootFS.WriteFile(name, []byte(data), 0777)
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

func (f *writeFS) WriteFile(path string, data []byte) error {
	return f.fs.WriteFile(path, data, 0777)
}

type js = map[string]any

func TestAssetSave(t *testing.T) {
	defer asset.Reset()
	wfs := newWriteFS()
	asset.RegisterWritableFileSystem(wfs)
	asset.RegisterAsset(testAsset{})

	name := "asset.json"
	a := &testAsset{Anykey: "saved"}
	err := asset.Save(name, a)
	assert.NoError(t, err)

	back, err := fs.ReadFile(wfs.fs, name)
	assert.NoError(t, err)

	jsonBack := js{}
	err = json.Unmarshal(back, &jsonBack)
	assert.NoError(t, err)

	expected := js{
		"Type": "flatland/src/asset_test.testAsset",
		"Inner": js{
			"Anykey": "saved",
		},
	}
	assert.Equal(t, expected, jsonBack)

	asset.RegisterFileSystem(wfs.fs, 0)
	loaded, err := asset.Load(name)
	assert.NoError(t, err)

	expectedObj := &testAsset{Anykey: "saved", postLoadCalled: true}
	assert.Equal(t, expectedObj, loaded)
}

func TestAssetList(t *testing.T) {
	defer asset.Reset()
	asset.RegisterAsset(testAsset{})

	expected := "flatland/src/asset_test.testAsset"
	assets := asset.ListAssets()
	assert.Equal(t, expected, assets[0].FullName, "Names don't match")
}
func TestAssetCreate(t *testing.T) {
	defer asset.Reset()
	asset.RegisterAsset(testAsset{})

	allAssets := asset.ListAssets()

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
	defer asset.Reset()
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
  "Type": "flatland/src/asset_test.testAssetNode",
  "Inner": {
    "Inline": {
      "SecondName": "inline"
    },
    "Name": "node",
    "Reference": {
      "Type": "flatland/src/asset_test.testAssetLeaf",
      "Path": "leaf.json"
    }
  }
}`
	assert.Equal(t, expected, string(bytesBack))

	bytesBack, err = fs.ReadFile(rootFS.fs, "leaf.json")
	leafExpected :=
		`{
  "Type": "flatland/src/asset_test.testAssetLeaf",
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
	defer asset.Reset()
	rootFS := newWriteFS()
	asset.RegisterFileSystem(rootFS.fs, 0)
	asset.RegisterWritableFileSystem(rootFS)

	asset.RegisterAsset(testAssetNode{})
	asset.RegisterAsset(testAssetLeaf{})

	leafRaw := `{
		"Type": "flatland/src/asset_test.testAssetLeaf",
		"Inner": {
		  "SecondName": "Leaf"
		}
	  }`
	err := rootFS.WriteFile("leaf.json", []byte(leafRaw))
	assert.NoError(t, err)

	nodeRaw := `{
		"Type": "flatland/src/asset_test.testAssetNode",
		"Inner": {
		  "Inline": {
			"SecondName": "inline"
		  },
		  "Name": "node",
		  "Reference": {
			"Type": "flatland/src/asset_test.testAssetLeaf",
			"Path": "leaf.json"
		  }
		}
	  }`
	err = rootFS.WriteFile("node.json", []byte(nodeRaw))
	assert.NoError(t, err)

	nodeAsset, err := asset.Load("node.json")
	assert.NoError(t, err)

	node := nodeAsset.(*testAssetNode)

	fmt.Printf("%#v\n", node)

	t.Fail()
}
