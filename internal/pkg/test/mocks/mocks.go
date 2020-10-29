package mocks

import (
	"testing"

	"github.com/petergtz/pegomock"
)

//go:generate pegomock generate --package=mocks --output=keyCreator.go -m github.com/airenas/api-doorman/internal/pkg/admin KeyCreator

//go:generate pegomock generate --package=mocks --output=keyRetriever.go -m github.com/airenas/api-doorman/internal/pkg/admin KeyRetriever

//go:generate pegomock generate --package=mocks --output=oneKeyRetriever.go -m github.com/airenas/api-doorman/internal/pkg/admin OneKeyRetriever

//go:generate pegomock generate --package=mocks --output=logRetriever.go -m github.com/airenas/api-doorman/internal/pkg/admin LogRetriever

//go:generate pegomock generate --package=mocks --output=keyUpdater.go -m github.com/airenas/api-doorman/internal/pkg/admin KeyUpdater

//AttachMockToTest register pegomock verification to be passed to testing engine
func AttachMockToTest(t *testing.T) {
	pegomock.RegisterMockFailHandler(handleByTest(t))
}

func handleByTest(t *testing.T) pegomock.FailHandler {
	return func(message string, callerSkip ...int) {
		if message != "" {
			t.Error(message)
		}
	}
}
