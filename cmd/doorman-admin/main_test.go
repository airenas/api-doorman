package main

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func Test_initProjectReset(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(strings.NewReader("asrMonthlyReset: 600\nttsMonthlyReset: 2000"))
	assert.Nil(t, err)
	res, err := initProjectReset([]string{"asr", "tts"}, v)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(res))
	assert.Equal(t, 600.0, res["asr"])
	assert.Equal(t, 2000.0, res["tts"])
}
