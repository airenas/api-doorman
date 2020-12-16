package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorInit(t *testing.T) {
	pv, err := NewProjectConfigValidator("")
	assert.Nil(t, pv)
	assert.NotNil(t, err)
}

func TestValidatorInit_Fail(t *testing.T) {
	pv, err := NewProjectConfigValidator("a")
	assert.NotNil(t, pv)
	assert.Nil(t, err)
}

func TestValidator_Check(t *testing.T) {
	pv, _ := NewProjectConfigValidator("a")
	assert.True(t, pv.Check("a"))
	assert.False(t, pv.Check("aa"))
}

func TestValidator_CheckSeveral(t *testing.T) {
	pv, _ := NewProjectConfigValidator("a,bbb, aaaa  ")
	assert.True(t, pv.Check("a"))
	assert.True(t, pv.Check("bbb"))
	assert.True(t, pv.Check("aaaa"))
	assert.False(t, pv.Check("B"))
	assert.False(t, pv.Check("aa"))
}
