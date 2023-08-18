package flat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRotation(t *testing.T) {
	tr := Transform{}

	tr.AddRotation(370)
	assert.Equal(t, 10.0, tr.Rotation, "Rotation needs to remain normalized to 0..360")
	tr.AddRotation(-20)
	assert.Equal(t, 350.0, tr.Rotation, "Rotation needs to remain normalized to 0..360")
}
