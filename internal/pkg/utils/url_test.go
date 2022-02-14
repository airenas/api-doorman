package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURL(t *testing.T) {
	u, err := ParseURL("http://192.1.1.1:8000")
	assert.Nil(t, err)
	assert.NotNil(t, u)
}

func TestURL_Fail(t *testing.T) {
	_, err := ParseURL(":8000")
	assert.NotNil(t, err)
	_, err = ParseURL("http://")
	assert.NotNil(t, err)
}
