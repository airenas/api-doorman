package mongodb

import (
	"context"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/randkey"
	"github.com/pkg/errors"

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// KeySaver saves keys to mongo db
type KeySaver struct {
	SessionProvider *SessionProvider
	NewKeySize      int
}

//NewKeySaver creates KeySaver instance
func NewKeySaver(sessionProvider *SessionProvider, keySize int) (*KeySaver, error) {
	f := KeySaver{SessionProvider: sessionProvider}
	if keySize < 10 || keySize > 100 {
		return nil, errors.New("Wrong keySize")
	}
	f.NewKeySize = keySize
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
	res.Key = randkey.Generate(ss.NewKeySize)
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
	cursor, err := c.Find(ctx, bson.M{"key": sanitize(key)})
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

//Update update key record
func (ss *KeySaver) Update(key string, data map[string]interface{}) (*adminapi.Key, error) {
	logrus.Infof("Updating key")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return nil, err
	}

	defer session.EndSession(context.Background())
	c := session.Client().Database(store).Collection(keyTable)

	session.StartTransaction()
	var res keyRecord
	err = c.FindOne(ctx, bson.M{"key": sanitize(key)}).Decode(&res)
	if err != nil {
		return nil, err
	}

	updates := bson.M{}
	for k, v := range data {
		var err error
		ok := true
		if k == "limit" {
			updates["limit"], ok = v.(float64)
		} else if k == "validTo" {
			updates["validTo"], ok = v.(time.Time)
		} else if k == "disabled" {
			updates["disabled"], ok = v.(bool)
		} else {
			err = errors.New("Unknown field " + k)
		}
		if !ok {
			return nil, errors.Errorf("Can't parse %s: '%s'", k, v)
		}
		if err != nil {
			return nil, errors.Wrap(err, "Can't parse input")
		}
	}
	updates["updated"] = time.Now()

	update := bson.M{"$set": updates}
	err = c.FindOneAndUpdate(ctx, bson.M{"key": sanitize(key)}, update,
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&res)
	if err != nil {
		return nil, err
	}
	session.CommitTransaction(ctx)

	return mapTo(&res), nil
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
	res.Updated = v.Updated
	res.Disabled = v.Disabled
	return res
}
