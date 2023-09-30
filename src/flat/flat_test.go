package flat_test

import (
	"testing"

	"github.com/bradbev/flatland/src/flat"

	"github.com/stretchr/testify/assert"
)

func isActor(obj any) bool {
	_, ok := obj.(flat.Actor)
	return ok
}

type Foo struct {
	flat.ActorBase
	MoreStuff bool
}

func (f *Foo) BeginPlay() {}

func TestEmbedding(t *testing.T) {
	assert.False(t, isActor(&flat.ActorBase{}), "ActorBase isn't an Actor, because it doesn't implement BeginPlay")
	assert.True(t, isActor(&Foo{}))
}

type testActor1 struct{ flat.ActorBase }

func (t *testActor1) BeginPlay() {}

type testActor2 struct{ flat.ActorBase }

func (t *testActor2) BeginPlay() {}

func TestFindActor(t *testing.T) {
	w := flat.NewWorld()
	a1a := &testActor1{}
	a1b := &testActor1{}
	a2a := &testActor2{}
	a2b := &testActor2{}
	w.AddToWorld(a1a)
	w.AddToWorld(a1b)
	w.AddToWorld(a2a)
	w.AddToWorld(a2b)

	a1List := flat.FindActorsByType[*testActor1](w)
	assert.Equal(t, []*testActor1{a1a, a1b}, a1List)

	a2List := flat.FindActorsByType[*testActor2](w)
	assert.Equal(t, []*testActor2{a2a, a2b}, a2List)
}
