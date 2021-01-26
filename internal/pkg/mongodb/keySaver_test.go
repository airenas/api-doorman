package mongodb

import (
	"testing"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewNewKeySaver(t *testing.T) {
	pr, err := NewKeySaver(nil, 10)
	assert.NotNil(t, pr)
	assert.Nil(t, err)
	_, err = NewKeySaver(nil, 50)
	assert.Nil(t, err)
}

func TestNewNewKeySaver_Fail(t *testing.T) {
	_, err := NewKeySaver(nil, 5)
	assert.NotNil(t, err)
	_, err = NewKeySaver(nil, 101)
	assert.NotNil(t, err)
}

func TestPrepareUpdates_FailNoUpdates(t *testing.T) {
	data := make(map[string]interface{})
	pr, err := prepareUpdates(data)
	assert.Nil(t, pr)
	assert.True(t, errors.Is(err, api.ErrWrongField))
}
func TestPrepareUpdates(t *testing.T) {
	data := make(map[string]interface{})
	data["limit"] = 10.0
	data["disabled"] = true
	tn := time.Now()
	data["validTo"] = tn
	data["description"] = "olia"
	data["IPWhiteList"] = "1.1.1.1/32"
	pr, err := prepareUpdates(data)
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 10.0, pr["limit"])
	assert.Equal(t, true, pr["disabled"])
	assert.Equal(t, tn, pr["validTo"])
	assert.Equal(t, "1.1.1.1/32", pr["IPWhiteList"])
}

func TestPrepareUpdates_ParseTime(t *testing.T) {
	data := make(map[string]interface{})
	data["validTo"] = "2030-11-24T11:07:00Z"
	pr, err := prepareUpdates(data)
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	ts := pr["validTo"]
	tp := ts.(time.Time)
	assert.True(t, time.Date(2020, time.November, 10, 23, 0, 0, 0, time.UTC).Before(tp))
}

func TestPrepareUpdates_Tags(t *testing.T) {
	data := make(map[string]interface{})
	data["tags"] = append(*new([]interface{}), "olia", "aa")
	pr, err := prepareUpdates(data)
	assert.Nil(t, err)
	if assert.NotNil(t, pr) {
		ts := pr["tags"]
		assert.Equal(t, []string{"olia", "aa"}, ts)
	}
}

func TestPrepareUpdates_FailParseTime(t *testing.T) {
	data := make(map[string]interface{})
	data["validTo"] = "2030-11-24"
	_, err := prepareUpdates(data)
	assert.True(t, errors.Is(err, api.ErrWrongField))
}

func TestPrepareUpdates_Fail(t *testing.T) {
	data := make(map[string]interface{})
	data["limit1"] = 10.0
	_, err := prepareUpdates(data)
	assert.True(t, errors.Is(err, api.ErrWrongField))
}

func TestPrepareUpdates_FailConvert(t *testing.T) {
	data := make(map[string]interface{})
	data["limit"] = "aa10.0"
	_, err := prepareUpdates(data)
	assert.True(t, errors.Is(err, api.ErrWrongField))
}

func TestPrepareUpdates_FailIPWhiteList(t *testing.T) {
	data := make(map[string]interface{})
	data["IPWhiteList"] = "1.1.1"
	_, err := prepareUpdates(data)
	assert.True(t, errors.Is(err, api.ErrWrongField))
}

func TestPrepareUpdates_FailTags(t *testing.T) {
	data := make(map[string]interface{})
	data["tags"] = "aaa"
	_, err := prepareUpdates(data)
	assert.True(t, errors.Is(err, api.ErrWrongField))
}

func TestAsSlice(t *testing.T) {
	s, ok := asStringSlice(append(*new([]interface{}), "ok"))
	assert.True(t, ok)
	assert.Equal(t, []string{"ok"}, s)
	s, ok = asStringSlice(append(*new([]interface{}), "ok", "values"))
	assert.True(t, ok)
	assert.Equal(t, []string{"ok", "values"}, s)
	s, ok = asStringSlice(*new([]interface{}))
	assert.True(t, ok)
	assert.Equal(t, []string{}, s)
	s, ok = asStringSlice(append(*new([]interface{}), 1, 2, "ok"))
	assert.False(t, ok)
}

func TestMapTo(t *testing.T) {
	data := &keyRecord{}
	data.Tags = []string{"olia", "aa"}
	data.Manual = true
	res := mapTo(data)
	assert.Equal(t, []string{"olia", "aa"}, res.Tags)
	assert.Equal(t, true, res.Manual)
}
