package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuotaUpdate(t *testing.T) {
	var res keyRecord
	res.Limit = 100
	res.QuotaValue = 50
	res.QuotaValueFailed = 0
	ok := quotaUpdateValidate(&res, 10)
	assert.True(t, ok)
	assert.Equal(t, 60.0, res.QuotaValue)
	assert.Equal(t, 0.0, res.QuotaValueFailed)

	res.QuotaValue = 90
	res.QuotaValueFailed = 0
	ok = quotaUpdateValidate(&res, 10)
	assert.True(t, ok)
	assert.Equal(t, 100.0, res.QuotaValue)
}

func TestQuotaUpdate_Fail(t *testing.T) {
	var res keyRecord
	res.Limit = 100
	res.QuotaValue = 91
	res.QuotaValueFailed = 0
	ok := quotaUpdateValidate(&res, 10)
	assert.False(t, ok)
	assert.Equal(t, 101.0, res.QuotaValue)
	assert.Equal(t, 10.0, res.QuotaValueFailed)
}

func TestQuotaUpdate_WithFail(t *testing.T) {
	var res keyRecord
	res.Limit = 100
	res.QuotaValue = 100
	res.QuotaValueFailed = 10
	ok := quotaUpdateValidate(&res, 10)
	assert.True(t, ok)
	assert.Equal(t, 110.0, res.QuotaValue)
	assert.Equal(t, 10.0, res.QuotaValueFailed)

	res.QuotaValue = 100
	res.QuotaValueFailed = 9
	ok = quotaUpdateValidate(&res, 10)
	assert.False(t, ok)
	assert.Equal(t, 110.0, res.QuotaValue)
	assert.Equal(t, 19.0, res.QuotaValueFailed)
}

func TestSanitize(t *testing.T) {
	assert.Equal(t, "olia", sanitize("olia"))
	assert.Equal(t, "olia", sanitize("$^olia$"))
	assert.Equal(t, "olia", sanitize("\\olia$ "))
	assert.Equal(t, "olia", sanitize("/$olia"))
}
