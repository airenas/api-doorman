package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomContext(t *testing.T) {
	r, data := customContext(httptest.NewRequest("POST", "/pref", nil))
	assert.NotNil(t, data)
	data.Key = "1111"
	r1, data1 := customContext(r)
	assert.Equal(t, "1111", data1.Key)
	assert.Equal(t, r, r1)
}
