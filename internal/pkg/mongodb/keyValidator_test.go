package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitize(t *testing.T) {
	assert.Equal(t, "olia", sanitize("olia"))
	assert.Equal(t, "olia", sanitize("$^olia$"))
	assert.Equal(t, "olia", sanitize("\\olia$ "))
	assert.Equal(t, "olia", sanitize("/$olia"))
}
