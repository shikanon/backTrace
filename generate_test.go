package backTrace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAllStrage(t *testing.T) {
	reg := GenerateAllStrage()
	assert.Equal(t, 172, len(reg.Names))
}
