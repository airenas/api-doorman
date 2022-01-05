package mongodb

import (
	"errors"
	"testing"

	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
	"github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
)

func TestNewDBProvider(t *testing.T) {
	mocks.AttachMockToTest(t)
	pr := mocks.NewMockSProvider()
	dpr, err := NewDBProvider(pr, "db")
	assert.NotNil(t, dpr)
	assert.Nil(t, err)
	pr.VerifyWasCalled(pegomock.Once()).CheckIndexes(pegomock.AnyStringSlice())
}

func TestNewDBProvider_Fail(t *testing.T) {
	mocks.AttachMockToTest(t)
	pr := mocks.NewMockSProvider()
	dpr, err := NewDBProvider(pr, "")
	assert.Nil(t, dpr)
	assert.NotNil(t, err)
	dpr, err = NewDBProvider(nil, "")
	assert.Nil(t, dpr)
	assert.NotNil(t, err)
	dpr, err = NewDBProvider(nil, "db")
	assert.Nil(t, dpr)
	assert.NotNil(t, err)
	pegomock.When(pr.CheckIndexes(pegomock.AnyStringSlice())).ThenReturn(errors.New("olia"))
	dpr, err = NewDBProvider(pr, "db")
	assert.Nil(t, dpr)
	assert.NotNil(t, err)
}
