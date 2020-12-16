package mongodb

import (
	"context"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/randkey"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
func (ss *KeySaver) Create(project string, key *adminapi.Key) (*adminapi.Key, error) {
	goapp.Log.Infof("Saving key - valid to: %v, limit: %f", key.ValidTo, key.Limit)
	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession(project)
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(project).Collection(keyTable)
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
func (ss *KeySaver) List(project string) ([]*adminapi.Key, error) {
	goapp.Log.Infof("getting list")
	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession(project)
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(project).Collection(keyTable)
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
func (ss *KeySaver) Get(project string, key string) (*adminapi.Key, error) {
	goapp.Log.Debug("Getting key")
	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession(project)
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(project).Collection(keyTable)
	var res keyRecord
	err = c.FindOne(ctx, bson.M{"key": sanitize(key)}).Decode(&res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, adminapi.ErrNoRecord
		}
		return nil, errors.Wrap(err, "Can't get keys")
	}
	return mapTo(&res), nil
}

//Update update key record
func (ss *KeySaver) Update(project string, key string, data map[string]interface{}) (*adminapi.Key, error) {
	goapp.Log.Debug("Updating key")
	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession(project)
	if err != nil {
		return nil, err
	}

	defer session.EndSession(context.Background())
	c := session.Client().Database(project).Collection(keyTable)

	session.StartTransaction()
	var res keyRecord
	err = c.FindOne(ctx, bson.M{"key": sanitize(key)}).Decode(&res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, adminapi.ErrNoRecord
		}
		return nil, err
	}

	updates, err := prepareUpdates(data)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, adminapi.ErrNoRecord
		}
		return nil, err
	}

	update := bson.M{"$set": updates}
	err = c.FindOneAndUpdate(ctx, bson.M{"key": sanitize(key)}, update,
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&res)
	if err != nil {
		return nil, err
	}
	session.CommitTransaction(ctx)

	return mapTo(&res), nil
}

func prepareUpdates(data map[string]interface{}) (bson.M, error) {
	res := bson.M{}
	for k, v := range data {
		var err error
		ok := true
		if k == "limit" {
			res["limit"], ok = v.(float64)
		} else if k == "validTo" {
			res["validTo"], ok = v.(time.Time)
			if !ok {
				res["validTo"], err = time.Parse(time.RFC3339, v.(string))
				ok = err == nil
			}

		} else if k == "disabled" {
			res["disabled"], ok = v.(bool)
		} else {
			err = errors.Wrapf(adminapi.ErrWrongField, "Unknown field '%s'", k)
		}
		if !ok {
			return nil, errors.Wrapf(adminapi.ErrWrongField, "Can't parse %s: '%s'", k, v)
		}
		if err != nil {
			return nil, errors.Wrap(err, "Can't parse input")
		}
	}
	if len(res) == 0 {
		return nil, errors.Wrapf(adminapi.ErrWrongField, "No updates")
	}
	res["updated"] = time.Now()
	return res, nil
}

func mapTo(v *keyRecord) *adminapi.Key {
	res := &adminapi.Key{}
	res.Key = v.Key
	res.Manual = v.Manual
	res.ValidTo = v.ValidTo
	res.Limit = v.Limit
	res.QuotaValue = v.QuotaValue
	res.QuotaFailed = v.QuotaValueFailed
	res.Created = v.Created
	res.LastUsed = v.LastUsed
	res.LastIP = v.LastIP
	res.Updated = v.Updated
	res.Disabled = v.Disabled
	return res
}
