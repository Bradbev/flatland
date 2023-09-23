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
		asset.ResetForTest()
		name := asset.Path("asset.json")
		callback()
		newTestAsset(name, data)

		a, err := asset.Load(name)
		assert.NoError(t, err)
		toTest, ok := a.(*testAsset)
		assert.True(t, ok, "Unable to convert %v to testType", a)
		assert.Equal(t, toTest.Anykey, "hi")
		assert.True(t, toTest.postLoadCalled, "PostLoad was not called")

		inst, err := asset.NewInstance(toTest)
		assert.NoError(t, err)
		prevTest := toTest
		toTest, ok = inst.(*testAsset)
		assert.True(t, toTest != prevTest, "Pointers must be different, but they are the same.")
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
	asset.ResetForTest()
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
	asset.ResetForTest()
	asset.RegisterAsset(testAsset{})
	testLoad(true, "Expected post load to be true because we reset the asset package")
}

func TestAssetList(t *testing.T) {
	asset.ResetForTest()
	asset.RegisterAsset(testAsset{})

	expected := "github.com/bradbev/flatland/src/asset_test.testAsset"
	assets := asset.GetAssetDescriptors()
	assert.Equal(t, expected, assets[0].FullName, "Names don't match")
}
func TestAssetCreate(t *testing.T) {
	asset.ResetForTest()
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
	asset.ResetForTest()
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
	asset.ResetForTest()
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
		asset.ResetForTest()
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

type inlineInnerSaveStruct struct {
	ItemA string
	ItemB string
}

type inlineSaving struct {
	WillNotSave         *inlineInnerSaveStruct
	SaveInline          *inlineInnerSaveStruct `flat:"inline"`
	SaveInlineNoChanges *inlineInnerSaveStruct `flat:"inline"`
	SaveInterfaceInline any                    `flat:"inline"`
}

func TestInlineAssetSaving(t *testing.T) {
	rootFS := newWriteFS()
	reset := func() {
		asset.ResetForTest()
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
	toSave.SaveInlineNoChanges = parent
	asset.SetParent(toSave.SaveInlineNoChanges, parent)

	err := asset.Save("inlined.json", toSave)
	assert.NoError(t, err)

	d, _ := asset.ReadFile("inlined.json")
	fmt.Println(string(d))

	reset()

	loaded, err := asset.Load("inlined.json")
	assert.NoError(t, err)
	expected := &inlineSaving{
		SaveInline:          &inlineInnerSaveStruct{"InlineA", "ParentB"},
		SaveInlineNoChanges: &inlineInnerSaveStruct{"ParentA", "ParentB"},
		SaveInterfaceInline: &inlineInnerSaveStruct{"IfaceA", "IfaceB"},
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
    },
    "SaveInlineNoChanges": {
      "Type": "github.com/bradbev/flatland/src/asset_test.inlineInnerSaveStruct",
      "Parent": "parent.json",
      "Inner": null
    },
    "SaveInterfaceInline": {
      "Type": "github.com/bradbev/flatland/src/asset_test.inlineInnerSaveStruct",
      "Parent": "",
      "Inner": {
        "ItemA": "IfaceA",
        "ItemB": "IfaceB"
      }
    }
  }
}`
	data, err := fs.ReadFile(rootFS.fs, "inlined.json")
	assert.Equal(t, exactFileContent, string(data), "The inlined field has a parent etc")
}

type postLoadChecker struct {
	hasPostLoaded bool
}

func (p *postLoadChecker) PostLoad() { p.hasPostLoaded = true }

type postLoadParent struct {
	InlineSaved *postLoadChecker `flat:"inline"`
}

type testFilter struct{}

func (t *testFilter) IFaceFunc() {}

type testFilter2 struct{}

type testIFace interface {
	IFaceFunc()
}

func TestAssetTypeFiltering(t *testing.T) {
	rootFS := newWriteFS()
	reset := func() {
		asset.ResetForTest()
		asset.RegisterFileSystem(rootFS.fs, 0)
		asset.RegisterWritableFileSystem(rootFS)
		asset.RegisterAsset(testFilter{})
		asset.RegisterAsset(testFilter2{})
	}
	reset()

	asset.Save("testfilter1.json", &testFilter{})
	// have a second file to validate we aren't just returning everything
	asset.Save("testfilter2.json", &testFilter2{})

	files, err := asset.FilterFilesByType[testFilter]()
	assert.NoError(t, err)
	assert.Equal(t, []string{"testfilter1.json"}, files, "Exact type of the object matches")

	files, err = asset.FilterFilesByType[testIFace]()
	assert.NoError(t, err)
	assert.Equal(t, []string{"testfilter1.json"}, files, "Interface types matches")

	files, err = asset.FilterFilesByType[any]()
	assert.NoError(t, err)
	assert.Equal(t, []string{"testfilter1.json", "testfilter2.json"}, files, "Any matches every file")
}

func TestAssetDescriptorTypeFiltering(t *testing.T) {
	rootFS := newWriteFS()
	reset := func() {
		asset.ResetForTest()
		asset.RegisterFileSystem(rootFS.fs, 0)
		asset.RegisterWritableFileSystem(rootFS)
		asset.RegisterAsset(testFilter{})
		asset.RegisterAsset(testFilter2{})
	}
	reset()

	allDescriptors := asset.GetAssetDescriptors()
	var expected *asset.AssetDescriptor
	for _, d := range allDescriptors {
		if d.Name == "testFilter" {
			expected = d
		}
	}

	desc := asset.FilterAssetDescriptorsByType[testFilter]()
	assert.Equal(t, expected, desc[0], "Exact type of the object matches")

	desc = asset.FilterAssetDescriptorsByType[testIFace]()
	assert.Equal(t, expected, desc[0], "Exact type of the object matches")

	desc = asset.FilterAssetDescriptorsByType[any]()
	assert.Equal(t, allDescriptors, desc, "Exact type of the object matches")
}

type newInstanceTest struct {
	SavedValue    int
	postloaded    bool
	internalValue int
}

func (n *newInstanceTest) PostLoad() {
	n.postloaded = true
}

func (n *newInstanceTest) DefaultInitialize() {
	n.internalValue = 111
}

func TestAssetNewInstance(t *testing.T) {
	rootFS := newWriteFS()
	reset := func() {
		asset.ResetForTest()
		asset.RegisterFileSystem(rootFS.fs, 0)
		asset.RegisterWritableFileSystem(rootFS)
		asset.RegisterAsset(newInstanceTest{})
	}
	reset()

	desc := asset.FilterAssetDescriptorsByType[newInstanceTest]()
	instA, err := desc[0].Create()
	inst := instA.(*newInstanceTest)
	assert.NoError(t, err)
	assert.Equal(t, 111, inst.internalValue, "Default initialize must have been called")
	assert.False(t, inst.postloaded, "Creating a new instance does not call PostLoad")
	inst.SavedValue = 123
	asset.Save("inst.json", inst)

	reset()
	instA, err = asset.Load("inst.json")
	inst = instA.(*newInstanceTest)
	assert.Equal(t, 111, inst.internalValue, "Default initialize must have been called")
	assert.True(t, inst.postloaded, "PostLoad is called when Load is used")

	newInst, err := asset.NewInstance(inst)
	instB := newInst.(*newInstanceTest)
	assert.Equal(t, 123, instB.SavedValue, "Exported values are copied")
	assert.Equal(t, 111, instB.internalValue, "Default initialize must have been called")
	assert.True(t, instB.postloaded, "PostLoad is called when Load is used")

}
