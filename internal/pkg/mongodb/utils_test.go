package mongodb

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestIsDuplicate(t *testing.T) {
	assert.False(t, IsDuplicate(nil))
	assert.False(t, IsDuplicate(errors.New("Olia")))
	assert.False(t, IsDuplicate(mongo.WriteException{WriteErrors: []mongo.WriteError{{Code: 1100}}}))
	assert.True(t, IsDuplicate(mongo.WriteException{WriteErrors: []mongo.WriteError{{Code: 11000}}}))
}


func TestSanitize(t *testing.T) {
	assert.Equal(t, "olia", Sanitize("olia"))
	assert.Equal(t, "olia", Sanitize("$^olia$"))
	assert.Equal(t, "olia", Sanitize("\\olia$ "))
	assert.Equal(t, "olia", Sanitize("/$olia"))
}
