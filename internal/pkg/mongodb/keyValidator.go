package mongodb

import (
	"context"
	"time"

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// KeyValidator validates key in mongo db
type KeyValidator struct {
	SessionProvider *SessionProvider
}

//NewKeyValidator creates KeyValidator instance
func NewKeyValidator(sessionProvider *SessionProvider) (*KeyValidator, error) {
	f := KeyValidator{SessionProvider: sessionProvider}
	return &f, nil
}

// IsValid validates key
func (ss *KeyValidator) IsValid(key string) (bool, error) {
	logrus.Infof("Validating key")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return false, err
	}

	defer session.EndSession(context.Background())
	c := session.Client().Database(store).Collection(keyTable)
	cursor, err := c.Find(ctx, bson.M{"key": key}, options.Find().SetLimit(1))
	if err != nil {
		return false, errors.Wrap(err, "Can't get keys")
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var res adminapi.Key
		if err = cursor.Decode(&res); err != nil {
			return false, errors.Wrap(err, "Can't get key")
		}
		ok := res.ValidTo.After(time.Now())
		if !ok {
			logrus.Infof("Key expired")
		}
		return ok, nil
	}
	return false, nil
}
