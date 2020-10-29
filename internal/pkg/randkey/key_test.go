package randkey

import (
	"math/rand"
	"testing"

	"gotest.tools/assert"
)

func TestKey(t *testing.T) {
	rand.Seed(10)
	assert.Equal(t, "wSv9wq3TdG", Generate(10))
	assert.Equal(t, "TjgXccmV5G", Generate(10))
}

func TestLen(t *testing.T) {
	rand.Seed(10)
	assert.Equal(t, 2, len(Generate(2)))
	assert.Equal(t, 22, len(Generate(22)))
}
