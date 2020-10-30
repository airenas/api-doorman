package mongodb

import (
	"context"
	"time"

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
func (ss *KeyValidator) IsValid(key string, manual bool) (bool, error) {
	logrus.Infof("Validating key")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return false, err
	}

	defer session.EndSession(context.Background())
	c := session.Client().Database(store).Collection(keyTable)
	cursor, err := c.Find(ctx, bson.M{"key": key, "manual": manual}, options.Find().SetLimit(1))
	if err != nil {
		return false, errors.Wrap(err, "Can't get keys")
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var res keyRecord
		if err = cursor.Decode(&res); err != nil {
			return false, errors.Wrap(err, "Can't get key")
		}
		ok := res.ValidTo.After(time.Now())
		if !ok {
			logrus.Infof("Key expired")
			return ok, nil
		}
		ok = !res.Disabled
		if !ok {
			logrus.Infof("Key disabled")
		}
		return ok, nil
	}
	return false, nil
}

//SaveValidate add qv to quota and validates with quota limit
func (ss *KeyValidator) SaveValidate(key string, ip string, qv float64) (bool, float64, float64, error) {
	logrus.Infof("Validating key")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return false, 0, 0, err
	}

	defer session.EndSession(context.Background())
	c := session.Client().Database(store).Collection(keyTable)

	session.StartTransaction()
	var res keyRecord
	err = c.FindOne(ctx, bson.M{"key": key}).Decode(&res)
	if err != nil {
		return false, 0, 0, err
	}
	res.QuotaValue += qv
	ok := quotaUpdateValidate(&res, qv)
	res.LastUsed = time.Now()
	res.LastIP = ip

	update := bson.M{"$set": bson.M{"quotaValue": res.QuotaValue, "quotaValueFailed": res.QuotaValueFailed,
		"lastUsed": res.LastUsed, "lastIP": res.LastIP}}
	err = c.FindOneAndUpdate(ctx, bson.M{"key": key}, update).Err()
	if err != nil {
		return false, 0, 0, err
	}
	session.CommitTransaction(ctx)

	return ok, res.Limit - (res.QuotaValue - res.QuotaValueFailed), res.Limit, nil
}

func quotaUpdateValidate(res *keyRecord, qv float64) bool {
	res.QuotaValue += qv
	if res.Limit < (res.QuotaValue - res.QuotaValueFailed) {
		res.QuotaValueFailed += qv
		return false
	}
	return true
}
