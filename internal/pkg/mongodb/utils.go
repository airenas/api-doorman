package mongodb

import (
	"strings"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

// IsDuplicate returns true if error indicates index duplicate key error
func IsDuplicate(err error) bool {
	var e mongo.WriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}
	return false
}

// Sanitize sanitizes for mongo input
func Sanitize(s string) string {
	return strings.Trim(s, " $/^\\")
}
