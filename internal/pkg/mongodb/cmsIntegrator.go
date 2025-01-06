package mongodb

// import (
// 	"context"
// 	"fmt"
// 	"strings"
// 	"time"

// 	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
// 	"github.com/airenas/api-doorman/internal/pkg/randkey"
// 	"github.com/airenas/api-doorman/internal/pkg/utils"
// 	"github.com/google/uuid"
// 	"github.com/pkg/errors"
// 	"github.com/rs/zerolog/log"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// const (
// 	saveRequestTag = "x-tts-collect-data:always"
// )

// // CmsIntegrator integrator function implementation with mongoDB persistence
// type CmsIntegrator struct {
// 	sessionProvider        *SessionProvider
// 	newKeySize             int
// 	defaultValidToDuration time.Duration
// }

// // NewCmsIntegrator creates CmsIntegrator instance
// func NewCmsIntegrator(sessionProvider *SessionProvider, keySize int) (*CmsIntegrator, error) {
// 	f := CmsIntegrator{sessionProvider: sessionProvider}
// 	if keySize < 10 || keySize > 100 {
// 		return nil, errors.New("wrong keySize")
// 	}
// 	f.newKeySize = keySize
// 	f.defaultValidToDuration = time.Hour * 24 * 365 * 10 //aprox 10 years
// 	return &f, nil
// }

// // Create creates new key
// func (ss *CmsIntegrator) Create(input *api.CreateInput) (*api.Key, bool, error) {
// 	if err := validateInput(input); err != nil {
// 		return nil, false, err
// 	}

// 	sessCtx, cancel, err := newSessionWithContext(ss.sessionProvider)
// 	if err != nil {
// 		return nil, false, err
// 	}
// 	defer cancel()

// 	inserted := false

// 	resInt, err := sessCtx.WithTransaction(sessCtx, func(sessCtx mongo.SessionContext) (interface{}, error) {
// 		c := sessCtx.Client().Database(keyMapDB).Collection(keyMapTable)
// 		var keyMap keyMapRecord
// 		err = c.FindOne(sessCtx, bson.M{"externalID": Sanitize(input.ID)}).Decode(&keyMap)
// 		if err != nil {
// 			if err == mongo.ErrNoDocuments {
// 				key, err := ss.createKeyWithQuota(sessCtx, input)
// 				if err == nil {
// 					inserted = true
// 					return &api.Key{Key: key.Key}, nil
// 				}
// 				return nil, err
// 			}
// 			return nil, err
// 		}
// 		if keyMap.Project != input.Service {
// 			return nil, &api.ErrField{Field: "service", Msg: "exists for other service"}
// 		}
// 		return &api.Key{}, nil
// 	})

// 	if err != nil {
// 		return nil, false, err
// 	}
// 	res := resInt.(*api.Key)
// 	return res, inserted, nil
// }

// // GetKey by ID
// func (ss *CmsIntegrator) GetKey(keyID string) (*api.Key, error) {
// 	if keyID == "" {
// 		return nil, api.ErrNoRecord
// 	}
// 	sessCtx, cancel, err := newSessionWithContext(ss.sessionProvider)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cancel()
// 	keyMapR, err := loadKeyMapRecord(sessCtx, keyID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	keyR, err := loadKeyRecord(sessCtx, keyMapR.Project, keyMapR.KeyID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return mapToKey(keyMapR.Project, keyR, true), nil
// }

// // GetKeyID returns keyID by key value
// func (ss *CmsIntegrator) GetKeyID(key string) (*api.KeyID, error) {
// 	if key == "" {
// 		return nil, api.ErrNoRecord
// 	}
// 	sessCtx, cancel, err := newSessionWithContext(ss.sessionProvider)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cancel()

// 	c := sessCtx.Client().Database(keyMapDB).Collection(keyMapTable)
// 	keyMapR := &keyMapRecord{}
// 	err = c.FindOne(sessCtx, bson.M{"keyHash": utils.HashKey(key)}).Decode(&keyMapR)
// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			return nil, api.ErrNoRecord
// 		}
// 		return nil, errors.Wrapf(err, "can't load from %s.%s", keyMapDB, keyMapTable)
// 	}
// 	return &api.KeyID{ID: keyMapR.ExternalID, Service: keyMapR.Project}, nil
// }

// // AddCredits to the key
// func (ss *CmsIntegrator) AddCredits(keyID string, input *api.CreditsInput) (*api.Key, error) {
// 	if err := validateCreditsInput(input); err != nil {
// 		return nil, err
// 	}

// 	sessCtx, cancel, err := newSessionWithContext(ss.sessionProvider)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cancel()

// 	keyMapR, err := loadKeyMapRecord(sessCtx, keyID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	resInt, err := sessCtx.WithTransaction(sessCtx, func(sessCtx mongo.SessionContext) (interface{}, error) {
// 		return addQuota(sessCtx, keyMapR, input)
// 	})

// 	if err != nil && !errors.Is(err, api.ErrOperationExists) {
// 		return nil, err
// 	}
// 	keyR := resInt.(*keyRecord)
// 	return mapToKey(keyMapR.Project, keyR, false), err
// }

// // Change generates new key for keyID, disables the old one, returns new key
// func (ss *CmsIntegrator) Change(keyID string) (*api.Key, error) {
// 	sessCtx, cancel, err := newSessionWithContext(ss.sessionProvider)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cancel()

// 	keyMapR, err := loadKeyMapRecord(sessCtx, keyID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	resInt, err := sessCtx.WithTransaction(sessCtx, func(sessCtx mongo.SessionContext) (interface{}, error) {
// 		key, err := ss.changeKey(sessCtx, keyMapR)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return key, nil
// 	})

// 	if err != nil {
// 		return nil, err
// 	}
// 	keyR := resInt.(*keyRecord)
// 	return &api.Key{Key: keyR.Key}, nil
// }

// // Update updates key table fields
// func (ss *CmsIntegrator) Update(keyID string, input map[string]interface{}) (*api.Key, error) {
// 	sessCtx, cancel, err := newSessionWithContext(ss.sessionProvider)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cancel()

// 	keyMapR, err := loadKeyMapRecord(sessCtx, keyID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	update, err := prepareKeyUpdates(input, time.Now())
// 	if err != nil {
// 		return nil, err
// 	}
// 	keyR := &keyRecord{}
// 	c := sessCtx.Client().Database(keyMapR.Project).Collection(keyTable)
// 	err = c.FindOneAndUpdate(sessCtx,
// 		keyFilter(keyMapR.KeyID),
// 		bson.M{"$set": update},
// 		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&keyR)
// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			return nil, api.ErrNoRecord
// 		}
// 		return nil, errors.Wrapf(err, "can't update %s.key", keyTable)
// 	}
// 	return mapToKey(keyMapR.Project, keyR, false), nil
// }

// func prepareKeyUpdates(input map[string]interface{}, now time.Time) (bson.M, error) {
// 	res := bson.M{}
// 	for k, v := range input {
// 		var err error
// 		ok := true
// 		if k == "validTo" {
// 			res["validTo"], ok = v.(time.Time)
// 			if !ok {
// 				res["validTo"], err = time.Parse(time.RFC3339, v.(string))
// 				ok = err == nil
// 			}
// 			if ok {
// 				ok = res["validTo"].(time.Time).After(now)
// 				if !ok {
// 					return nil, &api.ErrField{Field: k, Msg: "past date"}
// 				}
// 			}
// 		} else if k == "disabled" {
// 			res["disabled"], ok = v.(bool)
// 		} else if k == "IPWhiteList" {
// 			var s string
// 			s, ok = v.(string)
// 			if ok {
// 				err := utils.ValidateIPsCIDR(s)
// 				if err != nil {
// 					return nil, &api.ErrField{Field: k, Msg: "wrong IP CIDR format"}
// 				}
// 				res["IPWhiteList"] = v
// 			}
// 		} else {
// 			err = &api.ErrField{Field: k, Msg: "unknown or unsuported update"}
// 		}
// 		if !ok || err != nil {
// 			if err != nil {
// 				log.Error().Err(err).Send()
// 			}
// 			return nil, &api.ErrField{Field: k, Msg: "can't parse"}
// 		}
// 	}
// 	if len(res) == 0 {
// 		return nil, &api.ErrField{Field: "", Msg: "no updates"}
// 	}
// 	res["updated"] = now
// 	return res, nil
// }

// // Usage returns usage information for the key
// func (ss *CmsIntegrator) Usage(keyID string, from, to *time.Time, full bool) (*api.Usage, error) {
// 	sessCtx, cancel, err := newSessionWithContext(ss.sessionProvider)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cancel()

// 	keyMapR, err := loadKeyMapRecord(sessCtx, keyID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	filter := makeDateFilter(keyMapR.KeyID, from, to)

// 	c := sessCtx.Client().Database(keyMapR.Project).Collection(logTable)
// 	cursor, err := c.Find(sessCtx, filter)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "can't get logs")
// 	}
// 	defer cursor.Close(sessCtx)
// 	res := &api.Usage{}
// 	for cursor.Next(sessCtx) {
// 		var logR logRecord
// 		if err := cursor.Decode(&logR); err != nil {
// 			return nil, errors.Wrap(err, "can't get log record")
// 		}
// 		if logR.Fail {
// 			res.FailedCredits += logR.QuotaValue
// 		} else {
// 			res.UsedCredits += logR.QuotaValue
// 		}
// 		res.RequestCount++
// 		if full {
// 			res.Logs = append(res.Logs, mapLogRecord(&logR))
// 		}
// 	}
// 	if err := cursor.Err(); err != nil {
// 		return nil, fmt.Errorf("can't get logs: %w", err)
// 	}
// 	return res, err
// }

// // Changes returns changed keys information
// func (ss *CmsIntegrator) Changes(from *time.Time, services []string) (*api.Changes, error) {
// 	sessCtx, cancel, err := newSessionWithContext(ss.sessionProvider)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cancel()

// 	res := &api.Changes{}
// 	res.From = from
// 	to := time.Now().Add(-time.Millisecond) // make sure we will not loose some updates, so add -1 ms
// 	for _, s := range services {
// 		keys, err := loadKeys(sessCtx, s, from)
// 		if err != nil {
// 			return nil, err
// 		}
// 		for _, k := range keys {
// 			if k.ExternalID != "" { // skip without IDs
// 				res.Data = append(res.Data, mapToKey(s, k, false))
// 			}
// 		}
// 	}
// 	res.Till = &to
// 	return res, nil
// }

// func loadKeys(sessCtx mongo.SessionContext, service string, from *time.Time) ([]*keyRecord, error) {
// 	c := sessCtx.Client().Database(service).Collection(keyTable)
// 	filter := makeDateFilterForKey(from, nil)
// 	cursor, err := c.Find(sessCtx, filter)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "can't get keys")
// 	}
// 	defer cursor.Close(sessCtx)
// 	res := []*keyRecord{}
// 	for cursor.Next(sessCtx) {
// 		var keyR keyRecord
// 		if err := cursor.Decode(&keyR); err != nil {
// 			return nil, errors.Wrap(err, "can't get key record")
// 		}
// 		res = append(res, &keyR)
// 	}
// 	if err := cursor.Err(); err != nil {
// 		return nil, fmt.Errorf("can't get keys: %w", err)
// 	}
// 	return res, nil
// }

// func makeDateFilterForKey(from, to *time.Time) bson.M {
// 	res := bson.M{"manual": true}
// 	df := getDateFilter(from, to)
// 	if len(df) > 0 {
// 		res["updated"] = df
// 	}
// 	return res
// }

// func getDateFilter(from, to *time.Time) bson.M {
// 	var res bson.M
// 	if from != nil || to != nil {
// 		res = bson.M{}
// 		if from != nil {
// 			res["$gte"] = *from
// 		}
// 		if to != nil {
// 			res["$lt"] = *to
// 		}
// 	}
// 	return res
// }

// func makeDateFilter(keyID string, from, to *time.Time) bson.M {
// 	res := bson.M{"keyID": Sanitize(keyID)}
// 	df := getDateFilter(from, to)
// 	if len(df) > 0 {
// 		res["date"] = df
// 	}
// 	return res
// }

// func mapLogRecord(log *logRecord) *api.Log {
// 	res := &api.Log{}
// 	res.Date = toTime(&log.Date)
// 	res.Fail = log.Fail
// 	res.Response = log.ResponseCode
// 	res.IP = log.IP
// 	res.UsedCredits = log.QuotaValue
// 	return res
// }

// func newSessionWithContext(sessionProvider *SessionProvider) (mongo.SessionContext, func(), error) {
// 	session, err := sessionProvider.NewSession()
// 	if err != nil {
// 		return nil, func() {}, err
// 	}
// 	ctx, cancel := mongoContext()
// 	cf := func() {
// 		defer cancel()
// 		defer session.EndSession(context.Background())
// 	}
// 	return mongo.NewSessionContext(ctx, session), cf, nil
// }

// func addQuota(sessCtx mongo.SessionContext, keyMapR *keyMapRecord, input *api.CreditsInput) (*keyRecord, error) {
// 	c := sessCtx.Client().Database(keyMapR.Project).Collection(operationTable)
// 	var operation operationRecord
// 	err := c.FindOne(sessCtx, bson.M{"operationID": Sanitize(input.OperationID)}).Decode(&operation)
// 	if err == nil {
// 		if operation.KeyID != keyMapR.KeyID {
// 			return nil, &api.ErrField{Field: "operationID", Msg: "exists for other key"}
// 		}
// 		res, err := loadKeyRecord(sessCtx, keyMapR.Project, keyMapR.KeyID)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return res, api.ErrOperationExists
// 	}
// 	if err != mongo.ErrNoDocuments {
// 		return nil, fmt.Errorf("find operation: %v", err)
// 	}

// 	operation.Date = time.Now()
// 	operation.KeyID = keyMapR.KeyID
// 	operation.OperationID = input.OperationID
// 	operation.QuotaValue = input.Credits
// 	_, err = c.InsertOne(sessCtx, operation)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "can't insert operation")
// 	}
// 	update := bson.M{"$inc": bson.M{"limit": input.Credits}, "$set": bson.M{"updated": time.Now()}}
// 	keyR := &keyRecord{}
// 	c = sessCtx.Client().Database(keyMapR.Project).Collection(keyTable)
// 	err = c.FindOneAndUpdate(sessCtx,
// 		keyFilter(keyMapR.KeyID),
// 		update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&keyR)
// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			return nil, api.ErrNoRecord
// 		}
// 		return nil, errors.Wrapf(err, "can't update %s.key", keyTable)
// 	}
// 	return keyR, nil
// }

// func loadKeyRecord(sessCtx mongo.SessionContext, project, keyID string) (*keyRecord, error) {
// 	c := sessCtx.Client().Database(project).Collection(keyTable)
// 	keyR := &keyRecord{}
// 	err := c.FindOne(sessCtx, keyFilter(keyID)).Decode(&keyR)
// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			return nil, api.ErrNoRecord
// 		}
// 		return nil, errors.Wrapf(err, "can't load from %s.key", project)
// 	}
// 	return keyR, nil
// }

// func keyFilter(keyID string) bson.M {
// 	return bson.M{"keyID": Sanitize(keyID), "manual": true}
// }

// func loadKeyMapRecord(sessCtx mongo.SessionContext, keyID string) (*keyMapRecord, error) {
// 	c := sessCtx.Client().Database(keyMapDB).Collection(keyMapTable)
// 	keyMapR := &keyMapRecord{}
// 	err := c.FindOne(sessCtx, bson.M{"externalID": Sanitize(keyID)}).Decode(&keyMapR)
// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			return nil, api.ErrNoRecord
// 		}
// 		return nil, errors.Wrapf(err, "can't load from %s.%s", keyMapDB, keyMapTable)
// 	}
// 	return keyMapR, nil
// }

// func mapToKey(service string, keyR *keyRecord, returnKey bool) *api.Key {
// 	res := &api.Key{Service: service,
// 		ID:       keyR.ExternalID,
// 		ValidTo:  toTime(&keyR.ValidTo),
// 		LastUsed: toTime(&keyR.LastUsed), LastIP: keyR.LastIP,
// 		TotalCredits: keyR.Limit, UsedCredits: keyR.QuotaValue, FailedCredits: keyR.QuotaValueFailed,
// 		Disabled: keyR.Disabled, Created: toTime(&keyR.Created),
// 		Updated:      toTime(&keyR.Updated),
// 		IPWhiteList:  keyR.IPWhiteList,
// 		SaveRequests: mapToSaveRequests(keyR.Tags),
// 	}
// 	if returnKey {
// 		res.Key = keyR.Key
// 	}
// 	return res
// }

// func mapToSaveRequests(tags []string) bool {
// 	for _, s := range tags {
// 		if s == saveRequestTag {
// 			return true
// 		}
// 	}
// 	return false
// }

// func toTime(time *time.Time) *time.Time {
// 	if time == nil || time.IsZero() {
// 		return nil
// 	}
// 	return time
// }

// func (ss *CmsIntegrator) createKeyWithQuota(sessCtx mongo.SessionContext, input *api.CreateInput) (*keyRecord, error) {
// 	// create map
// 	c := sessCtx.Client().Database(keyMapDB).Collection(keyMapTable)
// 	var keyMap keyMapRecord
// 	keyMap.Created = time.Now()
// 	keyMap.ExternalID = input.ID
// 	key, err := randkey.Generate(ss.newKeySize)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "can't generate key")
// 	}
// 	keyMap.Project = input.Service
// 	keyMap.KeyHash = utils.HashKey(key)
// 	keyMap.KeyID = uuid.NewString()
// 	_, err = c.InsertOne(sessCtx, keyMap)
// 	if err != nil {
// 		if IsDuplicate(err) {
// 			return nil, errors.New("can't insert keymap - duplicate")
// 		}
// 		return nil, errors.Wrap(err, "can't insert keymap")
// 	}

// 	c = sessCtx.Client().Database(input.Service).Collection(operationTable)
// 	var operation operationRecord
// 	operation.Date = time.Now()
// 	operation.KeyID = keyMap.KeyID
// 	operation.OperationID = input.OperationID
// 	operation.QuotaValue = input.Credits
// 	_, err = c.InsertOne(sessCtx, operation)
// 	if err != nil {
// 		if IsDuplicate(err) {
// 			return nil, &api.ErrField{Field: "operationID", Msg: "duplicate"}
// 		}
// 		return nil, errors.Wrap(err, "can't insert operation")
// 	}
// 	c = sessCtx.Client().Database(input.Service).Collection(keyTable)
// 	res := initNewKey(input, ss.defaultValidToDuration, time.Now())
// 	res.Key = key
// 	res.KeyID = keyMap.KeyID
// 	_, err = c.InsertOne(sessCtx, res)
// 	if err != nil {
// 		if IsDuplicate(err) {
// 			return nil, errors.New("can't insert key - duplicate")
// 		}
// 		return nil, errors.Wrap(err, "can't insert key")
// 	}
// 	return res, err
// }

// func (ss *CmsIntegrator) changeKey(sessCtx mongo.SessionContext, keyMapR *keyMapRecord) (*keyRecord, error) {
// 	oldKey := keyMapR.KeyHash
// 	newKey, err := randkey.Generate(ss.newKeySize)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "can't generate key")
// 	}

// 	// update map
// 	c := sessCtx.Client().Database(keyMapDB).Collection(keyMapTable)
// 	err = c.FindOneAndUpdate(sessCtx,
// 		bson.M{"externalID": keyMapR.ExternalID},
// 		bson.M{"$set": bson.M{"keyHash": utils.HashKey(newKey), "updated": time.Now()},
// 			"$push": bson.M{"old": bson.M{"changedOn": time.Now(), "keyHash": oldKey}}}).Err()

// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			return nil, api.ErrNoRecord
// 		}
// 		return nil, errors.Wrap(err, "can't update keymap")
// 	}

// 	//update key
// 	c = sessCtx.Client().Database(keyMapR.Project).Collection(keyTable)
// 	res := &keyRecord{}
// 	err = c.FindOneAndUpdate(sessCtx,
// 		keyFilter(keyMapR.KeyID),
// 		bson.M{"$set": bson.M{"key": newKey, "updated": time.Now()}},
// 		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&res)

// 	if err != nil {
// 		return nil, errors.Wrap(err, "can't update key")
// 	}
// 	return res, err
// }

// func initNewKey(input *api.CreateInput, defDuration time.Duration, now time.Time) *keyRecord {
// 	res := &keyRecord{}
// 	res.Limit = input.Credits
// 	if input.ValidTo != nil {
// 		res.ValidTo = *input.ValidTo
// 	} else {
// 		res.ValidTo = now.Add(defDuration)
// 	}
// 	res.Created = now
// 	res.Updated = now
// 	res.Manual = true
// 	res.ExternalID = input.ID
// 	if input.SaveRequests {
// 		res.Tags = []string{saveRequestTag}
// 	}
// 	return res
// }

// func validateInput(input *api.CreateInput) error {
// 	if input == nil {
// 		return &api.ErrField{Field: "id", Msg: "missing"}
// 	}
// 	if strings.TrimSpace(input.ID) == "" {
// 		return &api.ErrField{Field: "id", Msg: "missing"}
// 	}
// 	if strings.TrimSpace(input.OperationID) == "" {
// 		return &api.ErrField{Field: "operationID", Msg: "missing"}
// 	}
// 	if strings.TrimSpace(input.Service) == "" {
// 		return &api.ErrField{Field: "service", Msg: "missing"}
// 	}
// 	if input.ValidTo != nil && input.ValidTo.Before(time.Now()) {
// 		return &api.ErrField{Field: "validTo", Msg: "past date"}
// 	}
// 	if input.Credits <= 0.1 {
// 		return &api.ErrField{Field: "credits", Msg: "less than 0.1"}
// 	}
// 	return nil
// }

// func validateCreditsInput(input *api.CreditsInput) error {
// 	if input == nil {
// 		return &api.ErrField{Field: "operationID", Msg: "missing"}
// 	}
// 	if strings.TrimSpace(input.OperationID) == "" {
// 		return &api.ErrField{Field: "operationID", Msg: "missing"}
// 	}
// 	if input.Credits <= 0.1 {
// 		return &api.ErrField{Field: "credits", Msg: "less than 0.1"}
// 	}
// 	return nil
// }
