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
	res := &keyRecord{}
	res.Key = randkey.Generate()
	res.Limit = key.Limit
	res.ValidTo = key.ValidTo
	res.Created = time.Now()
	res.Manual = true
	_, err = c.InsertOne(ctx, res)
	return mapTo(res), err
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
		var key keyRecord
		if err = cursor.Decode(&key); err != nil {
			return nil, errors.Wrap(err, "Can't get key")
		}
		res = append(res, mapTo(&key))
	}
	return res, nil
}

// Get return one key record
func (ss *KeySaver) Get(key string) (*adminapi.Key, error) {
	logrus.Infof("getting list")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(store).Collection(keyTable)
	cursor, err := c.Find(ctx, bson.M{"key": key})
	if err != nil {
		return nil, errors.Wrap(err, "Can't get keys")
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var key keyRecord
		if err = cursor.Decode(&key); err != nil {
			return nil, errors.Wrap(err, "Can't get key")
		}
		return mapTo(&key), nil
	}
	return nil, nil
}

func mapTo(v *keyRecord) *adminapi.Key {
	res := &adminapi.Key{}
	res.Key = v.Key
	res.Manual = v.Manual
	res.ValidTo = v.ValidTo
	res.Limit = v.Limit
	res.QuotaValue = v.QuotaValue - v.QuotaValueFailed
	res.QuotaFailed = v.QuotaValueFailed
	res.Created = v.Created
	res.LastUsed = v.LastUsed
	res.LastIP = v.LastIP
	return res
}
