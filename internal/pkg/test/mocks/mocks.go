package mocks

import (
	"encoding/json"
	io "io"
	"strings"
	"testing"

	"github.com/petergtz/pegomock/v4"
)

//go:generate pegomock generate --package=mocks --output=oneKeyRetriever.go github.com/airenas/api-doorman/internal/pkg/admin OneKeyRetriever
//go:generate pegomock generate --package=mocks --output=logProvider.go github.com/airenas/api-doorman/internal/pkg/admin LogProvider
//go:generate pegomock generate --package=mocks --output=prValidator.go github.com/airenas/api-doorman/internal/pkg/admin PrValidator
//go:generate pegomock generate --package=mocks --output=usageRestorer.go github.com/airenas/api-doorman/internal/pkg/admin UsageRestorer

//go:generate pegomock generate --package=mocks --output=keyValidator.go github.com/airenas/api-doorman/internal/pkg/handler KeyValidator
//go:generate pegomock generate --package=mocks --output=quotaValidator.go github.com/airenas/api-doorman/internal/pkg/handler QuotaValidator
//go:generate pegomock generate --package=mocks --output=audioLenGetter.go github.com/airenas/api-doorman/internal/pkg/handler AudioLenGetter
//go:generate pegomock generate --package=mocks --output=textGetter.go github.com/airenas/api-doorman/internal/pkg/handler TextGetter
//go:generate pegomock generate --package=mocks --output=dbSaver.go github.com/airenas/api-doorman/internal/pkg/handler DBSaver
//go:generate pegomock generate --package=mocks --output=ipSaver.go github.com/airenas/api-doorman/internal/pkg/handler IPSaver
//go:generate pegomock generate --package=mocks --output=countGetter.go github.com/airenas/api-doorman/internal/pkg/handler CountGetter

//go:generate pegomock generate --package=mocks --output=ipManager.go github.com/airenas/api-doorman/internal/pkg/service IPManager

// AttachMockToTest register pegomock verification to be passed to testing engine
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

// ToReader convert object to string reader of JSON
func ToReader(data interface{}) io.Reader {
	bytes, _ := json.Marshal(data)
	return strings.NewReader(string(bytes))
}
