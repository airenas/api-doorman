package mongodb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSanitize(t *testing.T) {
	assert.Equal(t, "olia", sanitize("olia"))
	assert.Equal(t, "olia", sanitize("$^olia$"))
	assert.Equal(t, "olia", sanitize("\\olia$ "))
	assert.Equal(t, "olia", sanitize("/$olia"))
}

func TestValidateKey(t *testing.T) {
	key := &keyRecord{Disabled: false, IPWhiteList: "", ValidTo: time.Now().Add(time.Minute)}
	testValidate(t, key, "", true, true)
	key = &keyRecord{Disabled: true, IPWhiteList: "", ValidTo: time.Now().Add(time.Minute)}
	testValidate(t, key, "", false, true)
	key = &keyRecord{Disabled: false, IPWhiteList: "", ValidTo: time.Now()}
	testValidate(t, key, "", false, true)
	key = &keyRecord{Disabled: false, IPWhiteList: "1.1.1.1/32", ValidTo: time.Now().Add(time.Minute)}
	testValidate(t, key, "1.1.1.1", true, true)
}

func TestValidateKey_Fail(t *testing.T) {
	key := &keyRecord{Disabled: false, IPWhiteList: "1.1.1.1/32", ValidTo: time.Now().Add(time.Minute)}
	testValidate(t, key, "", false, false)
	key = &keyRecord{Disabled: false, IPWhiteList: "1.1.1.1", ValidTo: time.Now().Add(time.Minute)}
	testValidate(t, key, "1.1.1.1", false, false)
}

func testValidate(t *testing.T, key *keyRecord, ip string, okExp, errExp bool) {
	ok, err := validateKey(key, ip)
	assert.True(t, okExp == ok)
	assert.True(t, (err == nil) == errExp)
}
