package mongodb

import (
	"context"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/cmdapp"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	cmdapp.Log.Infof("Validating ip")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(store).Collection(keyTable)

	session.StartTransaction()
	cursor, err := c.Find(ctx, bson.M{"key": sanitize(ip), "manual": false}, options.Find().SetLimit(1))
	if err != nil {
		return errors.Wrap(err, "Can't get keys")
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var res keyRecord
		if err = cursor.Decode(&res); err != nil {
			return errors.Wrap(err, "Can't get key")
		}
		return nil
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
	session.CommitTransaction(ctx)
	return nil
}
