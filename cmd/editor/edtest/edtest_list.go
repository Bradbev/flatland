package edtest

import (
	"github.com/bradbev/flatland/src/editor/edgui"
)

var listTestFilter string

func ListTest() {
	testFilter.Draw("list2", nil)
}

var testFilter = edgui.FilteredList[*testListNode]{List: testList}

var testList = []*testListNode{
	{"foo"},
	{"bar"},
	{"baz"},
	{"foo1"},
	{"bar1"},
	{"baz1"},
}

type testListNode struct {
	name string
}

func (t *testListNode) Name() string { return t.name }
