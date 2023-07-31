package editor_test

import (
	"flatland/src/editor"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testReflect struct {
	F64 float64
	//String  string
	//private bool
}

type testEditorImpl struct{}

func (t *testEditorImpl) BeginStruct(string) {}
func (t *testEditorImpl) EndStruct()         {}
func (t *testEditorImpl) FieldName(string)   {}

func TestGenericEdit(t *testing.T) {
	e := editor.NewCommonEditor(&testEditorImpl{})
	f := float64(45)

	e.AddType(new(float64), func(types *editor.CommonEditor, v reflect.Value) error {
		assert.Equal(t, 45.0, v.Float())
		v.SetFloat(34)
		return nil
	})

	e.Edit(&f)
	assert.Equal(t, 34.0, f)

	tf := &testReflect{F64: 45.0}
	e.Edit(tf)
	assert.Equal(t, 34.0, tf.F64)
}
