package mongodb

import (
	"context"
	"strings"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// KeyValidator validates key in mongo db
type KeyValidator struct {
	SessionProvider *DBProvider
}

//NewKeyValidator creates KeyValidator instance
func NewKeyValidator(sessionProvider *DBProvider) (*KeyValidator, error) {
	f := KeyValidator{SessionProvider: sessionProvider}
	return &f, nil
}

// IsValid validates key
func (ss *KeyValidator) IsValid(key string, IP string, manual bool) (bool, []string, error) {
	goapp.Log.Debugf("Validating key")
	ctx, cancel := mongoContext()
	defer cancel()

	session, db, err := ss.SessionProvider.NewSesionDatabase()
	if err != nil {
		return false, nil, err
	}

	defer session.EndSession(context.Background())
	c := db.Collection(keyTable)
	var res keyRecord
	err = c.FindOne(ctx, bson.M{"key": sanitize(key), "manual": manual}).Decode(&res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			goapp.Log.Infof("No key")
			return false, nil, nil
		}
		return false, nil, errors.Wrap(err, "Can't get key")
	}
	ok, err := validateKey(&res, IP)
	if err != nil {
		return ok, nil, err
	}
	return ok, res.Tags, nil
}

func validateKey(key *keyRecord, IP string) (bool, error) {
	if key.Disabled {
		goapp.Log.Infof("Key disabled")
		return false, nil
	}
	if !key.ValidTo.After(time.Now()) {
		goapp.Log.Infof("Key expired")
		return false, nil
	}
	res, err := utils.ValidateIPInWhiteList(key.IPWhiteList, IP)
	if !res {
		goapp.Log.Infof("IP white list '%s' does not allow IP '%s'", key.IPWhiteList, IP)
		if err != nil {
			goapp.Log.Error("Error: ", err)
		}
	}
	return res, err
}

//SaveValidate add qv to quota and validates with quota limit
func (ss *KeyValidator) SaveValidate(key string, ip string, manual bool, qv float64) (bool, float64, float64, error) {
	goapp.Log.Debugf("Validating key")
	ctx, cancel := mongoContext()
	defer cancel()

	session, db, err := ss.SessionProvider.NewSesionDatabase()
	if err != nil {
		return false, 0, 0, err
	}

	defer session.EndSession(context.Background())
	c := db.Collection(keyTable)

	var res keyRecord
	err = c.FindOne(ctx, bson.M{"key": sanitize(key), "manual": manual}).Decode(&res)
	if err != nil {
		return false, 0, 0, err
	}

	remRequired := res.Limit - qv
	if remRequired <= 0 {
		return ss.updateFailed(c, key, ip, manual, qv)
	}
	now := time.Now()
	update := bson.M{"$set": bson.M{"lastUsed": now, "updated": now, "lastIP": ip},
		"$inc": bson.M{"quotaValue": qv}}
	var resNew keyRecord
	err = c.FindOneAndUpdate(ctx, bson.M{"key": sanitize(key), "manual": manual,
		"quotaValue": bson.M{"$not": bson.M{"$gt": remRequired}}},
		update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&resNew)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ss.updateFailed(c, key, ip, manual, qv)
		}
		return false, 0, 0, err
	}

	return true, resNew.Limit - resNew.QuotaValue, resNew.Limit, nil
}

//Restore restores quota value after failed service call
func (ss *KeyValidator) Restore(key string, manual bool, qv float64) (float64, float64, error) {
	goapp.Log.Debugf("Restoring quota for key")
	ctx, cancel := mongoContext()
	defer cancel()

	session, db, err := ss.SessionProvider.NewSesionDatabase()
	if err != nil {
		return 0, 0, err
	}

	defer session.EndSession(context.Background())
	c := db.Collection(keyTable)

	now := time.Now()
	update := bson.M{"$set": bson.M{"lastUsed": now, "updated": now},
		"$inc": bson.M{"quotaValueFailed": qv, "quotaValue": -qv}}
	var resNew keyRecord
	err = c.FindOneAndUpdate(ctx, bson.M{"key": sanitize(key), "manual": manual},
		update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&resNew)
	if err != nil {
		return 0, 0, err
	}

	return resNew.Limit - resNew.QuotaValue, resNew.Limit, nil
}

func (ss *KeyValidator) updateFailed(c *mongo.Collection, key string, ip string, manual bool, qv float64) (bool, float64, float64, error) {
	ctx, cancel := mongoContext()
	defer cancel()
	update := bson.M{"$set": bson.M{"lastUsed": time.Now(), "lastIP": ip},
		"$inc": bson.M{"quotaValueFailed": qv}}
	var res keyRecord
	err := c.FindOneAndUpdate(ctx, bson.M{"key": sanitize(key), "manual": manual},
		update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&res)
	if err != nil {
		return false, 0, 0, err
	}
	return false, res.Limit - res.QuotaValue, res.Limit, nil
}

func sanitize(s string) string {
	return strings.Trim(s, " $/^\\")
}
