package mongodb

import (
	"context"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/cmdapp"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// IPSaver validates saves ip into DB
type IPSaver struct {
	SessionProvider *SessionProvider
}

//NewIPSaver creates IPSaver instance
func NewIPSaver(sessionProvider *SessionProvider) (*IPSaver, error) {
	f := IPSaver{SessionProvider: sessionProvider}
	return &f, nil
}

// CheckCreate new key record if no exist
func (ss *IPSaver) CheckCreate(ip string, limit float64) error {
	cmdapp.Log.Debug("Validating IP")
	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(store).Collection(keyTable)

	err = c.FindOne(ctx, bson.M{"key": sanitize(ip), "manual": false}).Err()
	if err == nil {
		return nil
	}
	if err != mongo.ErrNoDocuments {
		return errors.Wrap(err, "Can't get keys")
	}

	res := &keyRecord{}
	res.Key = ip
	res.Manual = false
	res.Limit = limit
	res.ValidTo = time.Date(2100, time.Month(1), 1, 01, 0, 0, 0, time.UTC)
	res.Created = time.Now()
	_, err = c.InsertOne(ctx, res)
	if err != nil {
		return err
	}
	return nil
}
