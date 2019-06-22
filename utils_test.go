package backTrace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMeanFunction(t *testing.T) {
	var value = []float32{1.0, 1.5, 2.0, 2.5, 3.0}
	mean := Mean(value)
	assert.Equal(t, mean, float32(2))
}
