package mongodb

import (
	"context"
	"strings"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/cmdapp"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
func (ss *KeyValidator) IsValid(key string, manual bool) (bool, error) {
	cmdapp.Log.Debugf("Validating key")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return false, err
	}

	defer session.EndSession(context.Background())
	c := session.Client().Database(store).Collection(keyTable)
	var res keyRecord
	err = c.FindOne(ctx, bson.M{"key": sanitize(key), "manual": manual}).Decode(&res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			cmdapp.Log.Infof("No key")
			return false, nil
		}
		return false, errors.Wrap(err, "Can't get key")
	}
	ok := res.ValidTo.After(time.Now())
	if !ok {
		cmdapp.Log.Infof("Key expired")
		return ok, nil
	}
	ok = !res.Disabled
	if !ok {
		cmdapp.Log.Infof("Key disabled")
	}
	return ok, nil
}

//SaveValidate add qv to quota and validates with quota limit
func (ss *KeyValidator) SaveValidate(key string, ip string, qv float64) (bool, float64, float64, error) {
	cmdapp.Log.Debugf("Validating key")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return false, 0, 0, err
	}

	defer session.EndSession(context.Background())
	c := session.Client().Database(store).Collection(keyTable)

	var res keyRecord
	err = c.FindOne(ctx, bson.M{"key": sanitize(key)}).Decode(&res)
	if err != nil {
		return false, 0, 0, err
	}

	remRequired := res.Limit - qv
	if remRequired <= 0 {
		return ss.updateFailed(c, key, ip, qv)
	}
	update := bson.M{"$set": bson.M{"lastUsed": time.Now(), "lastIP": ip},
		"$inc": bson.M{"quotaValue": qv}}
	var resNew keyRecord
	err = c.FindOneAndUpdate(ctx, bson.M{"key": sanitize(key),
		"quotaValue": bson.M{"$not": bson.M{"$gt": remRequired}}},
		update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&resNew)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ss.updateFailed(c, key, ip, qv)
		}
		return false, 0, 0, err
	}

	return true, resNew.Limit - resNew.QuotaValue, resNew.Limit, nil
}

func (ss *KeyValidator) updateFailed(c *mongo.Collection, key string, ip string, qv float64) (bool, float64, float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	update := bson.M{"$set": bson.M{"lastUsed": time.Now(), "lastIP": ip},
		"$inc": bson.M{"quotaValueFailed": qv}}
	var res keyRecord
	err := c.FindOneAndUpdate(ctx, bson.M{"key": sanitize(key)},
		update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&res)
	if err != nil {
		return false, 0, 0, err
	}
	return false, res.Limit - res.QuotaValue, res.Limit, nil
}

func sanitize(s string) string {
	return strings.Trim(s, " $/^\\")
}
