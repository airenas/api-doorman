package mongodb

import (
	"context"
	"fmt"
	"strings"
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
	goapp.Log.Infof("reset project default quotas for %s(%f), at %s", project, limit, since.Format(time.RFC3339))
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
		return fmt.Errorf("get reset config: %w", err)
	}
	settings, ok := cfg.(*settingsRecord)
	if !ok {
		return fmt.Errorf("wrong reset cfg record")
	}
	if since.Before(settings.NextReset) {
		goapp.Log.Infof("skip reset %s before %s", since.Format(time.RFC3339), settings.NextReset.Format(time.RFC3339))
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
		err = reset(sessCtx, &keyMapRecord{KeyID: getKeyNoHash(it.KeyID, it.Key), Project: project}, settings.NextReset, limit-(it.Limit-it.QuotaValue))
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

func getKeyNoHash(id, key string) string {
	if id != "" {
		return id
	}
	return key
}

func getUpdateResetConfig(sessCtx mongo.SessionContext, project string, since time.Time) (*settingsRecord, error) {
	c := sessCtx.Client().Database(project).Collection(settingTable)
	var settings settingsRecord
	err := c.FindOne(sessCtx, bson.M{}).Decode(&settings)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return nil, err
		}
		goapp.Log.Warnf("no %s.setting", project)
		settings.NextReset = utils.StartOfMonth(since, 0)
	}
	err = c.FindOneAndUpdate(sessCtx,
		bson.M{},
		bson.M{"$set": bson.M{"updated": since, "resetStarted": since}},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)).Err()
	if err != nil {
		return nil, errors.Wrapf(err, "can't update %s.setting", project)
	}
	return &settings, nil
}

func updateResetConfig(sessCtx mongo.SessionContext, project string, next time.Time) error {
	c := sessCtx.Client().Database(project).Collection(settingTable)
	err := c.FindOneAndUpdate(sessCtx,
		bson.M{},
		bson.M{"$set": bson.M{"updated": time.Now(), "nextReset": next}},
		options.FindOneAndUpdate().SetUpsert(true)).Err()
	if err != nil {
		return errors.Wrapf(err, "can't update %s.setting", project)
	}
	return nil
}

func reset(sessCtx mongo.SessionContext, keyMapRecord *keyMapRecord, at time.Time, quota float64) error {
	goapp.Log.Debugf("updating %s(%s), quota: +%f (%s)", start(keyMapRecord.KeyID), keyMapRecord.Project, quota, at.Format(time.RFC3339))
	_, err := addQuotaForIP(sessCtx, keyMapRecord, &api.CreditsInput{OperationID: uuid.NewString(), Credits: quota, Msg: "monthly reset"}, at)
	if err != nil {
		return fmt.Errorf("addQuota: %v", err)
	}
	return nil
}

func start(s string) string {
	if len(s) > 3 {
		return s[:3] + "..."
	}
	return s
}

func addQuotaForIP(sessCtx mongo.SessionContext, keyMapR *keyMapRecord, input *api.CreditsInput, at time.Time) (*keyRecord, error) {
	c := sessCtx.Client().Database(keyMapR.Project).Collection(operationTable)
	var operation operationRecord
	err := c.FindOne(sessCtx, bson.M{"operationID": Sanitize(input.OperationID)}).Decode(&operation)
	if err == nil {
		if operation.KeyID != keyMapR.KeyID {
			return nil, &api.ErrField{Field: "operationID", Msg: "exists for other key"}
		}
		return nil, api.ErrOperationExists
	}
	if err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("find operations: %v", err)
	}

	operation.Date = time.Now()
	operation.KeyID = keyMapR.KeyID
	operation.OperationID = input.OperationID
	operation.QuotaValue = input.Credits
	operation.Msg = input.Msg
	_, err = c.InsertOne(sessCtx, operation)
	if err != nil {
		return nil, errors.Wrap(err, "can't insert operation")
	}
	update := bson.M{"$inc": bson.M{"limit": input.Credits}, "$set": bson.M{"updated": time.Now(), "resetAt": at}}
	keyR := &keyRecord{}
	c = sessCtx.Client().Database(keyMapR.Project).Collection(keyTable)
	err = c.FindOneAndUpdate(sessCtx,
		keyFilterIP(keyMapR.KeyID),
		update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&keyR)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, api.ErrNoRecord
		}
		return nil, errors.Wrapf(err, "can't update %s.key", keyMapR.Project)
	}
	return keyR, nil
}

func keyFilterIP(keyID string) bson.M {
	return bson.M{getKeyField(keyID): Sanitize(keyID), "manual": false}
}

func getKeyField(key string) string {
	if strings.Contains(key, "-") { // UUID?
		return "keyID"
	}
	return "key"
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
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("can't get keys: %w", err)
	}
	return res, nil
}

func makeResetFilterForDate(at time.Time) bson.M {
	res := bson.M{"manual": false}
	res["created"] = bson.M{"$lt": at}
	res["$or"] = bson.A{bson.M{"resetAt": nil}, bson.M{"resetAt": bson.M{"$lt": at}}}
	return res
}
