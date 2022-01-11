package mongodb

import (
	"context"
	"strings"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
	"github.com/airenas/api-doorman/internal/pkg/randkey"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CmsIntegrator struct {
	sessionProvider        *SessionProvider
	newKeySize             int
	defaultValidToDuration time.Duration
}

//CmsIntegrator creates CmsIntegrator instance
func NewCmsIntegrator(sessionProvider *SessionProvider, keySize int) (*CmsIntegrator, error) {
	f := CmsIntegrator{sessionProvider: sessionProvider}
	if keySize < 10 || keySize > 100 {
		return nil, errors.New("wrong keySize")
	}
	f.newKeySize = keySize
	f.defaultValidToDuration = time.Hour * 24 * 365 * 10 //aprox 10 years
	return &f, nil
}

func (ss *CmsIntegrator) Create(input *api.CreateInput) (*api.Key, bool, error) {
	if err := validateInput(input); err != nil {
		return nil, false, err
	}

	sessCtx, cancel, err := newSessionWithContext(ss.sessionProvider)
	if err != nil {
		return nil, false, err
	}
	defer cancel()

	inserted := false

	resInt, err := sessCtx.WithTransaction(sessCtx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		c := sessCtx.Client().Database(keyMapDB).Collection(keyMapTable)
		var keyMap keyMapRecord
		err = c.FindOne(sessCtx, bson.M{"externalID": sanitize(input.ID)}).Decode(&keyMap)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				key, err := ss.createKeyWithQuota(sessCtx, input)
				if err == nil {
					inserted = true
					return &api.Key{Key: key.Key}, nil
				}
				return nil, err
			}
			return nil, err
		}
		if keyMap.Project != input.Service {
			return nil, &api.ErrField{Field: "service", Msg: "exists for other service"}
		}
		return &api.Key{Key: keyMap.Key}, nil
	})

	if err != nil {
		return nil, false, err
	}
	res := resInt.(*api.Key)
	return res, inserted, err
}

func (ss *CmsIntegrator) GetKey(keyID string) (*api.Key, error) {
	if keyID == "" {
		return nil, api.ErrNoRecord
	}
	sessCtx, cancel, err := newSessionWithContext(ss.sessionProvider)
	if err != nil {
		return nil, err
	}
	defer cancel()
	keyMapR, err := loadKeyMapRecord(sessCtx, keyID)
	if err != nil {
		return nil, err
	}
	keyR, err := loadKeyRecord(sessCtx, keyMapR.Project, keyMapR.Key)
	if err != nil {
		return nil, err
	}
	return mapToKey(keyMapR, keyR), nil
}

func (ss *CmsIntegrator) GetKeyID(key string) (*api.KeyID, error) {
	if key == "" {
		return nil, api.ErrNoRecord
	}
	sessCtx, cancel, err := newSessionWithContext(ss.sessionProvider)
	if err != nil {
		return nil, err
	}
	defer cancel()

	c := sessCtx.Client().Database(keyMapDB).Collection(keyMapTable)
	keyMapR := &keyMapRecord{}
	err = c.FindOne(sessCtx, bson.M{"key": sanitize(key)}).Decode(&keyMapR)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, api.ErrNoRecord
		}
		return nil, errors.Wrapf(err, "can't load from %s.%s", keyMapDB, keyMapTable)
	}
	return &api.KeyID{ID: keyMapR.ExternalID, Service: keyMapR.Project}, nil
}

func (ss *CmsIntegrator) AddCredits(keyID string, input *api.CreditsInput) (*api.Key, error) {
	if err := validateCreditsInput(input); err != nil {
		return nil, err
	}

	sessCtx, cancel, err := newSessionWithContext(ss.sessionProvider)
	if err != nil {
		return nil, err
	}
	defer cancel()

	keyMapR, err := loadKeyMapRecord(sessCtx, keyID)
	if err != nil {
		return nil, err
	}

	resInt, err := sessCtx.WithTransaction(sessCtx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		key, err := addQuota(sessCtx, keyMapR, input)
		if err != nil {
			return nil, err
		}
		return key, nil
	})

	if err != nil {
		return nil, err
	}
	keyR := resInt.(*keyRecord)
	res := mapToKey(keyMapR, keyR)
	res.Key = ""
	return res, err
}

func (ss *CmsIntegrator) Usage(keyID string, from, to *time.Time, full bool) (*api.Usage, error) {
	sessCtx, cancel, err := newSessionWithContext(ss.sessionProvider)
	if err != nil {
		return nil, err
	}
	defer cancel()

	keyMapR, err := loadKeyMapRecord(sessCtx, keyID)
	if err != nil {
		return nil, err
	}

	filter := makeDateFilter(keyMapR.Key, from, to)

	c := sessCtx.Client().Database(keyMapR.Project).Collection(logTable)
	cursor, err := c.Find(sessCtx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "can't get logs")
	}
	defer cursor.Close(sessCtx)
	res := &api.Usage{}
	for cursor.Next(sessCtx) {
		var logR logRecord
		if err := cursor.Decode(&logR); err != nil {
			return nil, errors.Wrap(err, "can't get log record")
		}
		if logR.Fail {
			res.FailedCredits += logR.QuotaValue
		} else {
			res.UsedCredits += logR.QuotaValue
		}
		res.RequestCount++
		if full {
			res.Logs = append(res.Logs, mapLogRecord(&logR))
		}
	}
	return res, err
}

func makeDateFilter(key string, from, to *time.Time) bson.M {
	res := bson.M{"key": sanitize(key)}
	if from != nil || to != nil {
		df := bson.M{}
		if from != nil {
			df["$gte"] = *from
		}
		if to != nil {
			df["$lt"] = *to
		}
		res["date"] = df
	}
	return res
}

func mapLogRecord(log *logRecord) *api.Log {
	res := &api.Log{}
	res.Date = toTime(&log.Date)
	res.Fail = log.Fail
	res.Response = log.ResponseCode
	res.IP = log.IP
	res.UsedCredits = log.QuotaValue
	return res
}

func newSessionWithContext(sessionProvider *SessionProvider) (mongo.SessionContext, func(), error) {
	session, err := sessionProvider.NewSession()
	if err != nil {
		return nil, func() {}, err
	}
	ctx, cancel := mongoContext()
	cf := func() {
		defer cancel()
		defer session.EndSession(context.Background())
	}
	return mongo.NewSessionContext(ctx, session), cf, nil
}

func addQuota(sessCtx mongo.SessionContext, keyMapR *keyMapRecord, input *api.CreditsInput) (*keyRecord, error) {
	c := sessCtx.Client().Database(keyMapR.Project).Collection(operationTable)
	var operation operationRecord
	err := c.FindOne(sessCtx, bson.M{"operationID": sanitize(input.OperationID)}).Decode(&operation)
	if err == nil {
		if operation.Key != keyMapR.Key {
			return nil, &api.ErrField{Field: "operationID", Msg: "exists for other key"}
		}
		return loadKeyRecord(sessCtx, keyMapR.Project, keyMapR.Key)
	}
	if err != mongo.ErrNoDocuments {
		return nil, err
	}

	operation.Date = time.Now()
	operation.Key = keyMapR.Key
	operation.OperationID = input.OperationID
	operation.QuotaValue = input.Credits
	_, err = c.InsertOne(sessCtx, operation)
	if err != nil {
		return nil, errors.Wrap(err, "can't insert operation")
	}
	update := bson.M{"$inc": bson.M{"limit": input.Credits}}
	keyR := &keyRecord{}
	c = sessCtx.Client().Database(keyMapR.Project).Collection(keyTable)
	err = c.FindOneAndUpdate(sessCtx,
		bson.M{"key": sanitize(keyMapR.Key), "manual": true},
		update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&keyR)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, api.ErrNoRecord
		}
		return nil, errors.Wrapf(err, "can't update %s.key", keyTable)
	}
	return keyR, nil
}

func loadKeyRecord(sessCtx mongo.SessionContext, project, key string) (*keyRecord, error) {
	c := sessCtx.Client().Database(project).Collection(keyTable)
	keyR := &keyRecord{}
	err := c.FindOne(sessCtx, bson.M{"key": sanitize(key), "manual": true}).Decode(&keyR)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, api.ErrNoRecord
		}
		return nil, errors.Wrapf(err, "can't load from %s.key", project)
	}
	return keyR, nil
}

func loadKeyMapRecord(sessCtx mongo.SessionContext, keyID string) (*keyMapRecord, error) {
	c := sessCtx.Client().Database(keyMapDB).Collection(keyMapTable)
	keyMapR := &keyMapRecord{}
	err := c.FindOne(sessCtx, bson.M{"externalID": sanitize(keyID)}).Decode(&keyMapR)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, api.ErrNoRecord
		}
		return nil, errors.Wrapf(err, "can't load from %s.%s", keyMapDB, keyMapTable)
	}
	return keyMapR, nil
}

func mapToKey(keyMapR *keyMapRecord, keyR *keyRecord) *api.Key {
	return &api.Key{Key: keyMapR.Key, Service: keyMapR.Project, ValidTo: toTime(&keyR.ValidTo),
		LastUsed: toTime(&keyR.LastUsed), LastIP: keyR.LastIP,
		TotalCredits: keyR.Limit, UsedCredits: keyR.QuotaValue, FailedCredits: keyR.QuotaValueFailed,
		Disabled: keyR.Disabled, Created: toTime(&keyR.Created),
		Updated:     toTime(&keyR.Updated),
		IPWhiteList: keyR.IPWhiteList,
	}
}

func toTime(time *time.Time) *time.Time {
	if time == nil || time.IsZero() {
		return nil
	}
	return time
}

func (ss *CmsIntegrator) createKeyWithQuota(sessCtx mongo.SessionContext, input *api.CreateInput) (*keyRecord, error) {
	// create map
	c := sessCtx.Client().Database(keyMapDB).Collection(keyMapTable)
	var keyMap keyMapRecord
	keyMap.Created = time.Now()
	keyMap.ExternalID = input.ID
	keyMap.Key = randkey.Generate(ss.newKeySize)
	keyMap.Project = input.Service
	_, err := c.InsertOne(sessCtx, keyMap)
	if err != nil {
		if IsDuplicate(err) {
			return nil, errors.New("can't insert keymap - duplicate")
		}
		return nil, errors.Wrap(err, "can't insert keymap")
	}

	c = sessCtx.Client().Database(input.Service).Collection(operationTable)
	var operation operationRecord
	operation.Date = time.Now()
	operation.Key = keyMap.Key
	operation.OperationID = input.OperationID
	operation.QuotaValue = input.Credits
	_, err = c.InsertOne(sessCtx, operation)
	if err != nil {
		if IsDuplicate(err) {
			return nil, &api.ErrField{Field: "operationID", Msg: "duplicate"}
		}
		return nil, errors.Wrap(err, "can't insert operation")
	}
	c = sessCtx.Client().Database(input.Service).Collection(keyTable)
	res := initNewKey(input, ss.defaultValidToDuration, time.Now())
	res.Key = keyMap.Key
	_, err = c.InsertOne(sessCtx, res)
	if err != nil {
		if IsDuplicate(err) {
			return nil, errors.New("can't insert key - duplicate")
		}
		return nil, errors.Wrap(err, "can't insert key")
	}
	return res, err
}

func initNewKey(input *api.CreateInput, defDuration time.Duration, now time.Time) *keyRecord {
	res := &keyRecord{}
	res.Limit = input.Credits
	if input.ValidTo != nil {
		res.ValidTo = *input.ValidTo
	} else {
		res.ValidTo = now.Add(defDuration)
	}
	res.Created = now
	res.Manual = true
	if input.SaveRequests {
		res.Tags = []string{"x-tts-collect-data:always"}
	}
	return res
}

func validateInput(input *api.CreateInput) error {
	if input == nil {
		return &api.ErrField{Field: "id", Msg: "missing"}
	}
	if strings.TrimSpace(input.ID) == "" {
		return &api.ErrField{Field: "id", Msg: "missing"}
	}
	if strings.TrimSpace(input.OperationID) == "" {
		return &api.ErrField{Field: "operationID", Msg: "missing"}
	}
	if strings.TrimSpace(input.Service) == "" {
		return &api.ErrField{Field: "service", Msg: "missing"}
	}
	if input.ValidTo != nil && input.ValidTo.Before(time.Now()) {
		return &api.ErrField{Field: "validTo", Msg: "past date"}
	}
	if input.Credits <= 0.1 {
		return &api.ErrField{Field: "credits", Msg: "less than 0.1"}
	}
	return nil
}

func validateCreditsInput(input *api.CreditsInput) error {
	if input == nil {
		return &api.ErrField{Field: "operationID", Msg: "missing"}
	}
	if strings.TrimSpace(input.OperationID) == "" {
		return &api.ErrField{Field: "operationID", Msg: "missing"}
	}
	if input.Credits <= 0.1 {
		return &api.ErrField{Field: "credits", Msg: "less than 0.1"}
	}
	return nil
}
