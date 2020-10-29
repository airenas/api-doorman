package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNewKeySaver(t *testing.T) {
	pr, err := NewKeySaver(nil, 10)
	assert.NotNil(t, pr)
	_, err = NewKeySaver(nil, 50)
	assert.Nil(t, err)
}

func TestNewNewKeySaver_Fail(t *testing.T) {
	_, err := NewKeySaver(nil, 5)
	assert.NotNil(t, err)
	_, err = NewKeySaver(nil, 101)
	assert.NotNil(t, err)
}
