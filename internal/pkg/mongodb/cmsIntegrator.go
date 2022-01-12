package mongodb

import (
	"context"
	"strings"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
	"github.com/airenas/api-doorman/internal/pkg/randkey"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	saveRequestTag = "x-tts-collect-data:always"
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
	return res, inserted, nil
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
	return res, nil
}

func (ss *CmsIntegrator) Change(keyID string) (*api.Key, error) {
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
		key, err := ss.changeKey(sessCtx, keyMapR)
		if err != nil {
			return nil, err
		}
		return key, nil
	})

	if err != nil {
		return nil, err
	}
	keyR := resInt.(*keyRecord)
	return &api.Key{Key: keyR.Key}, nil
}

func (ss *CmsIntegrator) Update(keyID string, input map[string]interface{}) (*api.Key, error) {
	sessCtx, cancel, err := newSessionWithContext(ss.sessionProvider)
	if err != nil {
		return nil, err
	}
	defer cancel()

	keyMapR, err := loadKeyMapRecord(sessCtx, keyID)
	if err != nil {
		return nil, err
	}

	update, err := prepareKeyUpdates(input, time.Now())
	if err != nil {
		return nil, err
	}
	keyR := &keyRecord{}
	c := sessCtx.Client().Database(keyMapR.Project).Collection(keyTable)
	err = c.FindOneAndUpdate(sessCtx,
		keyFilter(keyMapR.Key),
		bson.M{"$set": update},
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&keyR)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, api.ErrNoRecord
		}
		return nil, errors.Wrapf(err, "can't update %s.key", keyTable)
	}
	res := mapToKey(keyMapR, keyR)
	res.Key = ""
	return res, nil
}

func prepareKeyUpdates(input map[string]interface{}, now time.Time) (bson.M, error) {
	res := bson.M{}
	for k, v := range input {
		var err error
		ok := true
		if k == "validTo" {
			res["validTo"], ok = v.(time.Time)
			if !ok {
				res["validTo"], err = time.Parse(time.RFC3339, v.(string))
				ok = err == nil
			}
			if ok {
				ok = res["validTo"].(time.Time).After(now)
				if !ok {
					return nil, &api.ErrField{Field: k, Msg: "past date"}
				}
			}
		} else if k == "disabled" {
			res["disabled"], ok = v.(bool)
		} else if k == "IPWhiteList" {
			var s string
			s, ok = v.(string)
			if ok {
				err := utils.ValidateIPsCIDR(s)
				if err != nil {
					return nil, &api.ErrField{Field: k, Msg: "wrong IP CIDR format"}
				}
				res["IPWhiteList"] = v
			}
		} else {
			err = &api.ErrField{Field: k, Msg: "unknown or unsuported update"}
		}
		if !ok || err != nil {
			if err != nil {
				goapp.Log.Error(err)
			}
			return nil, &api.ErrField{Field: k, Msg: "can't parse"}
		}
	}
	if len(res) == 0 {
		return nil, &api.ErrField{Field: "", Msg: "no updates"}
	}
	res["updated"] = now
	return res, nil
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

	filter := makeDateFilter(keyMapR.Key, keyMapR.Old, from, to)

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

func makeDateFilter(key string, old []oldKey, from, to *time.Time) bson.M {
	res := bson.M{"key": sanitize(key)}
	if len(old) > 0 {
		keys := []string{sanitize(key)}
		for _, k := range old {
			keys = append(keys, k.Key)
		}
		res["key"] = bson.M{"$in": keys}
	}
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
		keyFilter(keyMapR.Key),
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
	err := c.FindOne(sessCtx, keyFilter(key)).Decode(&keyR)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, api.ErrNoRecord
		}
		return nil, errors.Wrapf(err, "can't load from %s.key", project)
	}
	return keyR, nil
}

func keyFilter(key string) bson.M {
	return bson.M{"key": sanitize(key), "manual": true}
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
		Updated:      toTime(&keyR.Updated),
		IPWhiteList:  keyR.IPWhiteList,
		SaveRequests: mapToSaveRequests(keyR.Tags),
	}
}

func mapToSaveRequests(tags []string) bool {
	for _, s := range tags {
		if s == saveRequestTag {
			return true
		}
	}
	return false
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

func (ss *CmsIntegrator) changeKey(sessCtx mongo.SessionContext, keyMapR *keyMapRecord) (*keyRecord, error) {
	oldKey := keyMapR.Key
	newKey := randkey.Generate(ss.newKeySize)

	// update map
	c := sessCtx.Client().Database(keyMapDB).Collection(keyMapTable)
	err := c.FindOneAndUpdate(sessCtx,
		bson.M{"externalID": keyMapR.ExternalID},
		bson.M{"$set": bson.M{"key": newKey, "updated": time.Now()},
			"$push": bson.M{"old": bson.M{"changedOn": time.Now(), "key": oldKey}}}).Err()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, api.ErrNoRecord
		}
		return nil, errors.Wrap(err, "can't update keymap")
	}

	//update key
	c = sessCtx.Client().Database(keyMapR.Project).Collection(keyTable)
	res := &keyRecord{}
	err = c.FindOneAndUpdate(sessCtx,
		keyFilter(oldKey),
		bson.M{"$set": bson.M{"key": newKey, "updated": time.Now()}},
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&res)

	if err != nil {
		return nil, errors.Wrap(err, "can't update key")
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
		res.Tags = []string{saveRequestTag}
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
