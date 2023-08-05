package asset

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testLeaf struct{ Leaf string }
type testNode struct {
	Name         string
	Num          int32
	Flt          float32
	Small        uint8
	MissingValue uint8
	Slice        []int
	Array        [2]int
	Inline       testLeaf
	Ref          *testLeaf
}

// The primary test here is that the Ref pointer
// is saved out as an object that has the path to load from
// and not saved as an inline object
func TestBuildJsonToSave(t *testing.T) {
	leaf := testLeaf{Leaf: "Leaf"}

	node := testNode{
		Name:   "Node",
		Inline: testLeaf{Leaf: "Inline"},
		Ref:    &leaf,
	}

	a := assetManagerImpl{
		AssetToLoadPath: map[Asset]Path{},
	}
	a.AssetToLoadPath[&leaf] = "fullPath.json"
	m := a.buildJsonToSave(node)

	j, err := json.MarshalIndent(m, "", "")
	assert.NoError(t, err)

	expected :=
		`{
"Array": [
0,
0
],
"Flt": 0,
"Inline": {
"Leaf": "Inline"
},
"MissingValue": 0,
"Name": "Node",
"Num": 0,
"Ref": {
"Type": "flatland/src/asset.testLeaf",
"Path": "fullPath.json"
},
"Slice": null,
"Small": 0
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
			"Type": "flatland/src/asset.testLeaf",
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
		defer func() {
			a := recover()
			if a != nil {
				fmt.Printf("Recovered %#v", a)
				fmt.Println()
			}
		}()

		a.unmarshalFromAny(toUnmarshal, &node)

	}()
	fmt.Printf("---- NODE\n  %v\n", js(node))

	assert.Equal(t, &leaf, node.Ref, "Expected node.Ref to equal leaf")

	expected := testNode{
		Name:   "Node",
		Num:    34,
		Small:  45,
		Flt:    1.234,
		Slice:  []int{1, 2, 3},
		Array:  [2]int{4, 5},
		Inline: testLeaf{Leaf: "InlineLeaf"},
		Ref:    &testLeaf{Leaf: "RefLeaf"}, // NOTE, Equal tests the values, not the pointer addresses
	}
	assert.Equal(t, expected, node)

	//t.Fail()

}
