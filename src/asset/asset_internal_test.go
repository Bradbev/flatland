package asset

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/psanford/memfs"
	"github.com/stretchr/testify/assert"
)

type testLeaf struct{ Leaf string }
type testNode struct {
	AssetType    Asset
	Name         string
	Num          int32
	Flt          float32
	Small        uint8
	MissingValue uint8
	Slice        []int
	Array        [2]int
	Inline       testLeaf
	Ref          *testLeaf
	SliceOfIface []any
}

// The primary test here is that the Ref pointer
// is saved out as an object that has the path to load from
// and not saved as an inline object
func TestBuildJsonToSave(t *testing.T) {
	leaf := testLeaf{Leaf: "Leaf"}

	node := testNode{
		Name:         "Node",
		Inline:       testLeaf{Leaf: "Inline"},
		Ref:          &leaf,
		AssetType:    &leaf,
		SliceOfIface: []any{&leaf},
	}

	a := assetManagerImpl{
		AssetToLoadPath: map[Asset]Path{},
	}
	a.AssetToLoadPath[&leaf] = "fullPath.json"
	m := a.toCommonFormat(node)
	diffsFromParent := findDiffsFromParentCommonFormat(nil, m)

	j, err := json.MarshalIndent(diffsFromParent, "", "")
	assert.NoError(t, err)

	expected :=
		`{
"AssetType": {
"Type": "github.com/bradbev/flatland/src/asset.testLeaf",
"Path": "fullPath.json"
},
"Inline": {
"Leaf": "Inline"
},
"Name": "Node",
"Ref": {
"Type": "github.com/bradbev/flatland/src/asset.testLeaf",
"Path": "fullPath.json"
},
"SliceOfIface": [
{
"Type": "github.com/bradbev/flatland/src/asset.testLeaf",
"Path": "fullPath.json"
}
]
}`

	assert.Equal(t, expected, string(j))
}

func js(a any) string {
	b, _ := json.MarshalIndent(a, "", "  ")
	return string(b)
}

func TestUnmashallFromAny(t *testing.T) {
	jsToUse := `{
		"Name": "Node",
		"Num":34,
		"Small":45,
		"Flt":1.234,
 	    "Slice":[1,2,3],
		"Array":[4,5],
		"Inline": {"Leaf":"InlineLeaf"},
		"Ref": {
			"Type": "github.com/bradbev/flatland/src/asset.testLeaf",
			"Path": "fullPath.json"
		},
		"AssetType": {
			"Type": "github.com/bradbev/flatland/src/asset.testLeaf",
			"Path": "fullPath.json"
		}
	}`

	var toUnmarshal any
	err := json.Unmarshal([]byte(jsToUse), &toUnmarshal)
	assert.NoError(t, err)

	a := assetManagerImpl{
		AssetToLoadPath: map[Asset]Path{},
		LoadPathToAsset: map[Path]Asset{},
	}
	leaf := testLeaf{Leaf: "RefLeaf"}
	a.AssetToLoadPath[&leaf] = "fullPath.json"
	a.LoadPathToAsset["fullPath.json"] = &leaf

	var node testNode
	func() {
		// the recover is here because this is a PITA to debug
		defer func() {
			a := recover()
			if a != nil {
				fmt.Printf("Recovered %#v", a)
				fmt.Println()
			}
		}()

		a.unmarshalCommonFormat(toUnmarshal, &node)

	}()

	assert.Equal(t, &leaf, node.Ref, "Expected node.Ref to equal leaf")

	expected := testNode{
		Name:      "Node",
		Num:       34,
		Small:     45,
		Flt:       1.234,
		Slice:     []int{1, 2, 3},
		Array:     [2]int{4, 5},
		Inline:    testLeaf{Leaf: "InlineLeaf"},
		Ref:       &testLeaf{Leaf: "RefLeaf"}, // NOTE, Equal tests the values, not the pointer addresses
		AssetType: &testLeaf{Leaf: "RefLeaf"}, // NOTE, Equal tests the values, not the pointer addresses
	}
	assert.Equal(t, expected, node)
}

type parentInner struct {
	PInnerA string
	PInnerB string
}

type parent struct {
	StrA  string
	StrB  string
	Inner parentInner
	Slice []int
}

type jsmap = map[string]any

func TestFindDiffsFromParent(t *testing.T) {
	assetman := &assetManagerImpl{}
	fd := func(a, b any) any {
		r := findDiffsFromParentCommonFormat(
			assetman.toCommonFormat(a),
			assetman.toCommonFormat(b))
		//jsp("Return::::::", r)
		return r
	}

	{ // basics
		assert.Equal(t, 1, fd("parent", 1))
		assert.Equal(t, nil, fd("parent", "parent"))
		assert.Equal(t, "child", fd("parent", "child"))
	}

	{ // structs with nesting
		p := parent{StrA: "Foo", StrB: "Bar",
			Inner: parentInner{"PInnerA", "PInnerB"}}
		c := parent{StrA: "Bar", StrB: "Bar",
			Inner: parentInner{"child", "PInnerB"}}
		expected := jsmap{"StrA": "Bar",
			"Inner": jsmap{"PInnerA": "child"}}
		assert.Equal(t, expected, fd(p, c), "Nested children need to replace parent values")
	}

	{ // structs with defaults
		p := parent{StrA: "Foo", StrB: "Bar"}
		c := parent{StrA: "", StrB: "Bar"}
		expected := jsmap{"StrA": ""}
		assert.Equal(t, expected, fd(p, c), "Child zero values can replace parent values")
	}

	{ // slices
		p := parent{Slice: []int{}}
		c := parent{Slice: []int{}}
		assert.Nil(t, fd(p, c), "Empty slices shouldn't be saved")

		p = parent{Slice: []int{1, 2}}
		c = parent{Slice: []int{1, 2}}
		assert.Nil(t, fd(p, c), "Same slices should be nil")

		p = parent{Slice: []int{1, 2}}
		c = parent{Slice: []int{1, 3}}
		expected := jsmap{"Slice": []any{1, 3}}
		assert.Equal(t, expected, fd(p, c), "Differing slices need to be saved")
	}
}

type testDefInit struct {
	DidInit bool
}

func (t *testDefInit) DefaultInitialize() {
	t.DidInit = true
}

func TestDefaultInitializer(t *testing.T) {
	a := struct {
		A testDefInit
		B *testDefInit
		C any
	}{
		B: &testDefInit{},
		C: &testDefInit{},
	}

	callAllDefaultInitializers(&a)

	assert.True(t, a.A.DidInit)
	assert.True(t, a.B.DidInit)
	assert.True(t, a.C.(*testDefInit).DidInit)
}

type TestTags struct {
	A int `flat:"inline   ;  key:value   ;   desc:A long description ;  last:last value" other:"other value"`
}

func TestGetFlatTag(t *testing.T) {
	sf, _ := reflect.TypeOf(TestTags{}).FieldByName("A")
	table := []struct {
		key           string
		expectedValue string
	}{
		{"inline", ""},
		{"key", "value"},
		{"desc", "A long description"},
		{"last", "last value"},
	}

	for _, entry := range table {
		value, exists := GetFlatTag(&sf, entry.key)
		assert.True(t, exists, "Key %s must exist", entry.key)
		assert.Equal(t, entry.expectedValue, value)
	}
}

type writeFS struct {
	fs *memfs.FS
}

func newWriteFS() *writeFS {
	return &writeFS{
		fs: memfs.New(),
	}
}

func (f *writeFS) WriteFile(path Path, data []byte) error {
	return f.fs.WriteFile(string(path), data, 0777)
}

type testAssetParent struct {
	StrA string
	StrB string
}

// childContainer simulates the World object - that is, inline
// children who have parents
type childContainer struct {
	Children []*testAssetParent `flat:"inline"`
}

type testMemberPathCreations struct {
	First struct {
		Second int
		Next   struct{ Last int }
	}
	Third int
}

func TestCreatePathsFromCommonFormat(t *testing.T) {
	c := assetManager.toCommonFormat(testMemberPathCreations{})

	o := newChildOverrides()
	o.BuildFromCommonFormat(c)

	expectedPaths := []string{
		"First.Second",
		"First.Next.Last",
		"Third",
	}
	for _, p := range expectedPaths {
		assert.True(t, o.PathHasOverride(p))
	}
	assert.Len(t, o.overrides, len(expectedPaths))
}

func TestParentLoadingSavingSetting(t *testing.T) {
	rootFS := newWriteFS()
	reset := func() {
		ResetForTest()
		RegisterFileSystem(rootFS.fs, 0)
		RegisterWritableFileSystem(rootFS)
		RegisterAsset(testAssetParent{})
		RegisterAsset(childContainer{})
	}

	{ // save to the FS
		reset()
		parent := &testAssetParent{
			StrA: "ParentA",
			StrB: "ParentB",
		}
		err := Save("parent.json", parent)
		assert.NoError(t, err)

		child := &childContainer{}
		child.Children = append(child.Children, &testAssetParent{
			StrA: "ParentA",
			StrB: "ChildB",
		})
		SetParent(child.Children[0], parent)
		overrides := assetManager.ChildAssetOverrides[child.Children[0]]
		assert.NotNil(t, overrides)
		assert.False(t, overrides.PathHasOverride("StrA"), "When a parent is set, any values that match the parent will be inherited")
		assert.True(t, overrides.PathHasOverride("StrB"), "When a parent is set, any values that match the parent will be inherited")

		err = Save("child.json", child)
		assert.NoError(t, err)
	}

	// load them
	{
		reset()

		loadedChild, err := Load("child.json")
		assert.NoError(t, err)
		child := loadedChild.(*childContainer)

		overrides := assetManager.ChildAssetOverrides[child.Children[0]]
		assert.NotNil(t, overrides)
		assert.False(t, overrides.PathHasOverride("StrA"), "When a parent is set, any values that match the parent will be inherited")
		assert.True(t, overrides.PathHasOverride("StrB"), "When a parent is set, any values that match the parent will be inherited")

		expected := &testAssetParent{
			StrA: "ParentA",
			StrB: "ChildB",
		}
		assert.Equal(t, expected, child.Children[0])

		loadedParent, err := Load("parent.json")
		assert.NoError(t, err)
		parent := loadedParent.(*testAssetParent)
		parent.StrA = "ChangedParent"
		Save("parent.json", parent)

		// WORKMARK - reloading child assets needs to work
		// Overrides need to be respected when loading
		expected.StrA = "ChangedParent"
		assert.Equal(t, expected, child.Children[0])

	}

}
