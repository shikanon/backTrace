package backTrace

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateRedisClient(t *testing.T) {
	c, err := CreateRedisClient()
	assert.Equal(t, nil, err)

	_, _ = c.RPush("test", "A_B").Result()
	str, err := c.LPop("test").Result()
	assert.Equal(t, nil, err)
	assert.Equal(t, "A_B", str)
}
