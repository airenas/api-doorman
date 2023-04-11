package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Reseter resets monthly usage
type Reseter struct {
	sessionProvider *SessionProvider
}

// NewReseter creates Reseter instance
func NewReseter(sessionProvider *SessionProvider) (*Reseter, error) {
	res := Reseter{sessionProvider: sessionProvider}
	return &res, nil
}

// Reset does monthly reset
func (ss *Reseter) Reset(ctx context.Context, project string, since time.Time, limit float64) error {
	goapp.Log.Infof("reset project default quotas for %s, at %s", project, since.Format(time.RFC3339))
	session, err := ss.sessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())
	ctxInt, cf := context.WithTimeout(ctx, 60*time.Second)
	defer cf()
	sessCtx := mongo.NewSessionContext(ctxInt, session)

	cfg, err := sessCtx.WithTransaction(sessCtx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return getUpdateResetConfig(sessCtx, project, since)
	})
	if err != nil {
		return err
	}
	settings := cfg.(*settingsRecord)
	if settings.NextReset.Before(since) {
		goapp.Log.Infof("skip reset")
		return nil
	}
	if settings.ResetStarted.After(since.Add(time.Hour)) {
		goapp.Log.Infof("skip reset - started at %s", settings.ResetStarted.Format(time.RFC3339))
		return nil
	}

	items, err := getResetableItems(sessCtx, project, since)
	if err != nil {
		return fmt.Errorf("get items: %w", err)
	}
	ua, ta := 0, 0.0
	goapp.Log.Infof("items to check %d", len(items))
	for _, it := range items {
		if !since.After(it.ResetAt) {
			continue
		}
		if !since.After(it.Created) {
			continue
		}
		if (it.Limit - it.QuotaValue) >= limit {
			continue
		}
		ua++
		ta += limit - (it.Limit - it.QuotaValue)
		err = reset(sessCtx, &keyMapRecord{Key: it.Key, Project: project}, since, limit-(it.Limit-it.QuotaValue))
		if err != nil {
			return fmt.Errorf("reset: %w", err)
		}
	}
	goapp.Log.Infof("updated %d, total quota added %f", ua, ta)
	_, err = sessCtx.WithTransaction(sessCtx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, updateResetConfig(sessCtx, project, utils.StartOfMonth(since, 1))
	})
	return err
}

func getUpdateResetConfig(sessCtx mongo.SessionContext, project string, since time.Time) (*settingsRecord, error) {
	c := sessCtx.Client().Database(project).Collection(settingTable)
	var settings settingsRecord
	err := c.FindOne(sessCtx, bson.M{}).Decode(&settings)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return nil, err
		}
		settings.NextReset = utils.StartOfMonth(since, 0)
	}
	err = c.FindOneAndUpdate(sessCtx,
		bson.M{},
		bson.M{"$set": bson.M{"updated": since, "resetStarted": since}},
		options.FindOneAndUpdate().SetUpsert(true)).Err()
	if err != nil {
		return nil, errors.Wrapf(err, "can't update %s.setting", project)
	}
	return &settings, nil
}

func updateResetConfig(sessCtx mongo.SessionContext, project string, next time.Time) (error) {
	c := sessCtx.Client().Database(project).Collection(settingTable)
	err := c.FindOneAndUpdate(sessCtx,
		bson.M{},
		bson.M{"$set": bson.M{"updated": time.Now(), "nextReste": next}},
		options.FindOneAndUpdate().SetUpsert(true)).Err()
	if err != nil {
		return errors.Wrapf(err, "can't update %s.setting", project)
	}
	return nil
}

func reset(sessCtx mongo.SessionContext, keyMapRecord *keyMapRecord, since time.Time, quota float64) error {
	_, err := sessCtx.WithTransaction(sessCtx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return struct{}{}, resetInt(sessCtx, keyMapRecord, since, quota)
	})
	return err
}

func resetInt(sessCtx mongo.SessionContext, keyMapRecord *keyMapRecord, since time.Time, quota float64) error {
	_, err := addQuota(sessCtx, keyMapRecord, &api.CreditsInput{OperationID: uuid.NewString(), Credits: quota})
	if err != nil {
		return err
	}
	return markResetTime(sessCtx, keyMapRecord, since)
}

func markResetTime(sessCtx mongo.SessionContext, keyMapRecord *keyMapRecord, since time.Time) error {
	c := sessCtx.Client().Database(keyMapRecord.Project).Collection(keyTable)

	now := time.Now()
	update := bson.M{"$set": bson.M{"updated": now, "resetAt": since}}
	return c.FindOneAndUpdate(sessCtx, bson.M{"key": keyMapRecord.Key, "manual": false},
		update, options.FindOneAndUpdate()).Err()
}

func getResetableItems(sessCtx mongo.SessionContext, service string, at time.Time) ([]*keyRecord, error) {
	c := sessCtx.Client().Database(service).Collection(keyTable)
	filter := makeResetFilterForDate(at)
	cursor, err := c.Find(sessCtx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "can't get keys")
	}
	defer cursor.Close(sessCtx)
	res := []*keyRecord{}
	for cursor.Next(sessCtx) {
		var keyR keyRecord
		if err := cursor.Decode(&keyR); err != nil {
			return nil, errors.Wrap(err, "can't get key record")
		}
		res = append(res, &keyR)
	}
	return res, nil
}

func makeResetFilterForDate(at time.Time) bson.M {
	res := bson.M{"manual": false}
	res["created"] = bson.M{"lt": at}
	res["$or"] = bson.A{bson.M{"resetAt": nil}, bson.M{"resetAt": bson.M{"lt": at}}}
	return res
}
