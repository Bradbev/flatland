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

func TestEmbedding(t *testing.T) {
	assert.True(t, isActor(&flat.ActorBase{}))

	type Foo struct {
		flat.ActorBase
		MoreStuff bool
	}
	assert.True(t, isActor(&Foo{}))
}
