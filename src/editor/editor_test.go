package editor

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/bradbev/flatland/src/flat"

	"github.com/stretchr/testify/assert"
)

type testActor struct {
	flat.ActorBase
	More string
}

// Actors that embed flat.ActorBase are required to implement BeginPlay
// and call ActorBase.BeginPlay
func (ta *testActor) BeginPlay() {
	ta.ActorBase.BeginPlay(ta)
}

func TestImplements(t *testing.T) {
	//a := flat.Actor(&testActor{})
	a := new(flat.Actor)
	p := fmt.Printf

	ty := reflect.TypeOf(a)
	p("%v\n", ty)
	p("%v\n", ty.Kind())
	p("%v\n", ty.Elem().Kind())

	ta := reflect.TypeOf(&testActor{})
	tye := ty.Elem()
	p("%v\n", ta.Implements(tye))
}

func TestTypeEditorBasicFuncCalled(t *testing.T) {
	te := &typeEditor{
		typeEditFuncs: map[string]TypeEditorFn{},
	}
	intEdited := false
	te.AddType(new(int), func(tec *TypeEditContext, v reflect.Value) error {
		intEdited = true
		return nil
	})
	var i int = 32
	te.Edit(&TypeEditContext{}, &i)
	assert.True(t, intEdited)
}

type ifaceA interface{ A() }
type ifaceB interface{ B() }

type implesA struct{}

func (i *implesA) A() {}

type implesB struct{}

func (i *implesB) B() {}

type implesAB struct{}

func (i *implesAB) A() {}
func (i *implesAB) B() {}

func TestTypeEditorInterfaceSupport(t *testing.T) {
	te := &typeEditor{
		typeEditFuncs:      map[string]TypeEditorFn{},
		interfaceEditFuncs: map[reflect.Type]TypeEditorFn{},
	}
	mkErr := func(ret string) func(tec *TypeEditContext, v reflect.Value) error {
		return func(tec *TypeEditContext, v reflect.Value) error {
			return fmt.Errorf(ret)
		}
	}
	errsToStr := func(value reflect.Value) []string {
		fns := te.gatherInterfaceEditorFuncs(value)
		ret := []string{}
		for _, f := range fns {
			ret = append(ret, f.fn(nil, reflect.ValueOf(0)).Error())
		}
		return ret
	}

	te.AddType(new(ifaceA), mkErr("A"))
	assert.Len(t, te.interfaceEditFuncs, 1, "There must be one entry in the interface editors")
	te.AddType(new(ifaceB), mkErr("B"))
	assert.Len(t, te.interfaceEditFuncs, 2, "There must be two entries in the interface editors")

	a := implesA{}
	assert.Equal(t, []string{"A"}, errsToStr(reflect.ValueOf(&a)))

	b := implesB{}
	assert.Equal(t, []string{"B"}, errsToStr(reflect.ValueOf(&b)))

	ab := implesAB{}
	assert.Equal(t, []string{"A", "B"}, errsToStr(reflect.ValueOf(&ab)), "The order matters too, alphabetical")
}

func TestTypeEditorInterfacesMustHaveFuncs(t *testing.T) {
	te := &typeEditor{
		typeEditFuncs:      map[string]TypeEditorFn{},
		interfaceEditFuncs: map[reflect.Type]TypeEditorFn{},
	}
	assert.Panics(t, func() {
		te.AddType(new(any), func(tec *TypeEditContext, v reflect.Value) error {
			return nil
		})
	}, "Interfaces that can be edited must have at least one function on them, or they will catch every type")
}
