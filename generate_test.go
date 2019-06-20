package backTrace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAllStrage(t *testing.T) {
	reg := GenerateAllBuyStrage()
	assert.Equal(t, 86, len(reg.Names))
	reg = GenerateAllSellStrage()
	assert.Equal(t, 86, len(reg.Names))
}
