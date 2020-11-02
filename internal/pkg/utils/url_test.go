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
func TestHidePass_Mongo(t *testing.T) {
	assert.Equal(t, "mongodb://mongo:27017", HidePass("mongodb://mongo:27017"))
	assert.Equal(t, "mongodb://l:----@mongo:27017", HidePass("mongodb://l:olia@mongo:27017"))
}
