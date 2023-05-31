package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/google/uuid"

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

// NewKeySaver creates KeySaver instance
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

	err := utils.ValidateIPsCIDR(key.IPWhiteList)
	if err != nil {
		return nil, errors.Wrapf(adminapi.ErrWrongField, "Wrong IP CIDR format: "+key.IPWhiteList)
	}

	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(project).Collection(keyTable)
	res := &keyRecord{}
	res.Key = key.Key
	if res.Key == "" {
		res.Key, err = randkey.Generate(ss.NewKeySize)
		if err != nil {
			return nil, errors.Wrap(err, "can't generate key")
		}
	}
	res.Limit = key.Limit
	if key.ValidTo != nil {
		res.ValidTo = *key.ValidTo
	}
	res.KeyID = uuid.NewString()
	res.Created = time.Now()
	res.Updated = res.Created
	res.Manual = true
	res.Description = key.Description
	res.IPWhiteList = key.IPWhiteList
	res.Tags = key.Tags
	_, err = c.InsertOne(ctx, res)
	return mapTo(res), err
}

// List return all keys
func (ss *KeySaver) List(project string) ([]*adminapi.Key, error) {
	goapp.Log.Infof("getting list")
	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
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
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("can't get logs: %w", err)
	}
	return res, nil
}

// Get return one key record
func (ss *KeySaver) Get(project string, key string) (*adminapi.Key, error) {
	goapp.Log.Debug("Getting key")
	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(project).Collection(keyTable)
	var res keyRecord
	err = c.FindOne(ctx, bson.M{"key": Sanitize(key)}).Decode(&res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, adminapi.ErrNoRecord
		}
		return nil, errors.Wrap(err, "Can't get keys")
	}
	return mapTo(&res), nil
}

// Update update key record
func (ss *KeySaver) Update(project string, key string, data map[string]interface{}) (*adminapi.Key, error) {
	goapp.Log.Debug("Updating key")
	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return nil, err
	}

	defer session.EndSession(context.Background())
	c := session.Client().Database(project).Collection(keyTable)

	err = session.StartTransaction()
	if err != nil {
		return nil, err
	}
	var res keyRecord
	err = c.FindOne(ctx, bson.M{"key": Sanitize(key)}).Decode(&res)
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
	err = c.FindOneAndUpdate(ctx, bson.M{"key": Sanitize(key)}, update,
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&res)
	if err != nil {
		return nil, err
	}
	err = session.CommitTransaction(ctx)
	if err != nil {
		return nil, err
	}

	return mapTo(&res), nil
}

// RestoreUsage erstores usage by requestID
func (ss *KeySaver) RestoreUsage(project string, manual bool, requestID string, errMsg string) error {
	goapp.Log.Debugf("Restoring quota for requestID %s", requestID)
	sessCtx, cancel, err := newSessionWithContext(ss.SessionProvider)
	if err != nil {
		return err
	}
	defer cancel()

	_, err = sessCtx.WithTransaction(sessCtx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, restoreQuota(sessCtx, project, manual, requestID, errMsg)
	})
	return err
}

func restoreQuota(sessCtx mongo.SessionContext, project string, manual bool, requestID string, errMsg string) error {
	c := sessCtx.Client().Database(project).Collection(logTable)
	var logR logRecord
	err := c.FindOne(sessCtx, bson.M{"requestID": Sanitize(requestID)}).Decode(&logR)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return adminapi.ErrNoRecord
		}
		return err
	}

	if logR.Fail {
		return adminapi.ErrLogRestored
	}

	err = c.FindOneAndUpdate(sessCtx,
		bson.M{"requestID": Sanitize(requestID)},
		bson.M{"$set": bson.M{"fail": true, "errorMsg": errMsg}},
		options.FindOneAndUpdate()).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return api.ErrNoRecord
		}
		return errors.Wrapf(err, "can't update %s.log", project)
	}

	now := time.Now()
	update := bson.M{"$set": bson.M{"updated": now},
		"$inc": bson.M{"quotaValueFailed": logR.QuotaValue, "quotaValue": -logR.QuotaValue}}
	c = sessCtx.Client().Database(project).Collection(keyTable)
	keyF, key := "keyID", logR.KeyID
	if key == "" {
		keyF, key = "key", logR.Key
	}
	return c.FindOneAndUpdate(sessCtx, bson.M{keyF: key, "manual": manual},
		update, options.FindOneAndUpdate()).Err()
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

		} else if k == "description" {
			res["description"] = v
		} else if k == "disabled" {
			res["disabled"], ok = v.(bool)
		} else if k == "IPWhiteList" {
			var s string
			s, ok = v.(string)
			if ok {
				err := utils.ValidateIPsCIDR(s)
				if err != nil {
					return nil, errors.Wrapf(adminapi.ErrWrongField, "Wrong IP CIDR format: "+s)
				}
				res["IPWhiteList"] = v
			}
		} else if k == "tags" {
			var s []string
			s, ok = asStringSlice(v)
			if ok {
				res["tags"] = s
			}
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
	res.KeyID = v.KeyID
	res.Manual = v.Manual
	res.ValidTo = toTime(&v.ValidTo)
	res.Limit = v.Limit
	res.QuotaValue = v.QuotaValue
	res.QuotaFailed = v.QuotaValueFailed
	res.Created = toTime(&v.Created)
	res.LastUsed = toTime(&v.LastUsed)
	res.LastIP = v.LastIP
	res.Updated = toTime(&v.Updated)
	res.Disabled = v.Disabled
	res.IPWhiteList = v.IPWhiteList
	res.Description = v.Description
	res.Tags = v.Tags
	return res
}

func asStringSlice(d interface{}) ([]string, bool) {
	ds, ok := d.([]interface{})
	if !ok {
		return nil, ok
	}
	res := make([]string, len(ds))
	for i, v := range ds {
		res[i], ok = v.(string)
		if !ok {
			return nil, ok
		}
	}
	return res, ok
}
