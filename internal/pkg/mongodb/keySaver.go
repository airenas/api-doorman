package mongodb

import (
	"context"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/randkey"
	"github.com/pkg/errors"

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

// KeySaver saves keys to mongo db
type KeySaver struct {
	SessionProvider *SessionProvider
}

//NewKeySaver creates KeySaver instance
func NewKeySaver(sessionProvider *SessionProvider) (*KeySaver, error) {
	f := KeySaver{SessionProvider: sessionProvider}
	return &f, nil
}

// Create key to DB
func (ss *KeySaver) Create(key *adminapi.Key) (*adminapi.Key, error) {
	logrus.Infof("Saving key - valid to: %v, limit: %f", key.ValidTo, key.Limit)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(store).Collection(keyTable)
	res := adminapi.Key{}
	res.Key = randkey.Generate()
	res.Limit = key.Limit
	res.ValidTo = key.ValidTo
	// _, err = c.InsertOne(ctx, bson.M{"key": res.Key, "validTo": key.ValidTo.UnixNano(), "limit": key.Limit})
	_, err = c.InsertOne(ctx, res)
	return &res, err
}

// List return all keys
func (ss *KeySaver) List() ([]*adminapi.Key, error) {
	logrus.Infof("getting list")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(store).Collection(keyTable)
	cursor, err := c.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "Can't get keys")
	}
	defer cursor.Close(ctx)
	res := make([]*adminapi.Key, 0)
	for cursor.Next(ctx) {
		var key adminapi.Key
		if err = cursor.Decode(&key); err != nil {
			return nil, errors.Wrap(err, "Can't get key")
		}
		res = append(res, &key)
	}
	return res, nil
}
