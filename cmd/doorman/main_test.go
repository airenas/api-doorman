package main

import (
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
)

func TestInitFromConfig(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(strings.NewReader("handlers: default ,default\ndefault:\n  backend: http://olia.lt"))
	assert.Nil(t, err)
	res, err := initFromConfig(v, nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(res))
}

func TestInitFromConfig_Fail(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(strings.NewReader("handlers: default\n"))
	assert.Nil(t, err)
	res, err := initFromConfig(v, nil)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(res))
}
