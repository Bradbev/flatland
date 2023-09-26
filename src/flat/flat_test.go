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

type testC1 struct{ flat.Component }
type testC2 struct{ flat.Component }

func TestFindComponent(t *testing.T) {

}
