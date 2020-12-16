package mongodb

import (
	"context"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// IPSaver validates saves ip into DB
type IPSaver struct {
	SessionProvider *DBProvider
}

//NewIPSaver creates IPSaver instance
func NewIPSaver(sessionProvider *DBProvider) (*IPSaver, error) {
	f := IPSaver{SessionProvider: sessionProvider}
	return &f, nil
}

// CheckCreate new key record if no exist
func (ss *IPSaver) CheckCreate(ip string, limit float64) error {
	goapp.Log.Debug("Validating IP")
	ctx, cancel := mongoContext()
	defer cancel()

	session, db, err := ss.SessionProvider.NewSesionDatabase()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())
	c := db.Collection(keyTable)

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
