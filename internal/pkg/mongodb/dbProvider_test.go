package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDBProvider(t *testing.T) {
	pr, _ := NewSessionProvider("mongo://url")
	dpr, err := NewDBProvider(pr, "db")
	assert.NotNil(t, dpr)
	assert.Nil(t, err)
}

func TestNewDBProvider_Fail(t *testing.T) {
	pr, _ := NewSessionProvider("mongo://url")
	dpr, err := NewDBProvider(pr, "")
	assert.Nil(t, dpr)
	assert.NotNil(t, err)
	dpr, err = NewDBProvider(nil, "")
	assert.Nil(t, dpr)
	assert.NotNil(t, err)
	dpr, err = NewDBProvider(nil, "db")
	assert.Nil(t, dpr)
	assert.NotNil(t, err)
}
