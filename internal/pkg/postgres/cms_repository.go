package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
	"github.com/airenas/api-doorman/internal/pkg/randkey"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type CMSRepository struct {
	db                     *sqlx.DB
	newKeySize             int
	defaultValidToDuration time.Duration
}

// Change implements cms.Integrator.
func (c *CMSRepository) Change(ctx context.Context, id string) (*api.Key, error) {
	panic("unimplemented")
}

// Changes implements cms.Integrator.
func (c *CMSRepository) Changes(ctx context.Context, from *time.Time, projects []string) (*api.Changes, error) {
	panic("unimplemented")
}

// Update implements cms.Integrator.
func (c *CMSRepository) Update(ctx context.Context, id string, in map[string]interface{}) (*api.Key, error) {
	panic("unimplemented")
}

// Usage implements cms.Integrator.
func (c *CMSRepository) Usage(ctx context.Context, id string, from *time.Time, to *time.Time, full bool) (*api.Usage, error) {
	panic("unimplemented")
}

func NewCMSRepository(ctx context.Context, db *sqlx.DB, keySize int) (*CMSRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	if keySize < 10 || keySize > 100 {
		return nil, fmt.Errorf("wrong keySize")
	}
	f := CMSRepository{db: db, newKeySize: keySize, defaultValidToDuration: time.Hour * 24 * 365 * 10 /*aprox 10 years*/}
	return &f, nil
}

func (r *CMSRepository) Create(ctx context.Context, in *api.CreateInput) (*api.Key, bool /*created*/, error) {
	if err := validateInput(in); err != nil {
		return nil, false, err
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, false, fmt.Errorf("begin transaction: %w", err)
	}
	defer roolback(tx)

	if in.OperationID == "" {
		in.OperationID = uuid.NewString()
	}
	if in.ID == "" {
		in.ID = uuid.NewString()
	}
	key, err := randkey.Generate(r.newKeySize)
	if err != nil {
		return nil, false, fmt.Errorf("generate key: %w", err)
	}

	res, err := r.createKeyWithQuota(ctx, tx, in, key)
	if err != nil {
		return nil, false, err
	}
	if err := tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("commit transaction: %w", err)
	}
	return mapToKey(res, key), true, nil
}

func (r *CMSRepository) GetKey(ctx context.Context, id string) (*api.Key, error) {
	res, err := loadKeyRecord(ctx, r.db, id)
	if err != nil {
		return nil, mapErr(err)
	}
	return mapToKey(res, ""), nil
}

func (r *CMSRepository) GetKeyID(ctx context.Context, id string) (*api.KeyID, error) {
	hash := utils.HashKey(id)
	res, err := loadKeyRecordByHash(ctx, r.db, hash)
	if err != nil {
		return nil, mapErr(err)
	}
	return &api.KeyID{ID: res.ID, Service: res.Project}, nil
}

func (r *CMSRepository) AddCredits(ctx context.Context, id string, in *api.CreditsInput) (*api.Key, error) {
	if err := validateCreditsInput(in); err != nil {
		return nil, err
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer roolback(tx)

	res, err := r.addQuota(ctx, tx, id, in)
	if err != nil {
		if errors.Is(err, utils.ErrOperationExists) && res != nil {
			return mapToKey(res, ""), nil
		}
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}
	return mapToKey(res, ""), nil
}

func loadKeyRecord(ctx context.Context, db dbTx, id string) (*keyRecord, error) {
	var res keyRecord
	err := db.GetContext(ctx, &res, `
		SELECT id, project, manual, quota_limit, quota_value, valid_to, disabled, 
			ip_white_list, tags, created, updated, last_used, last_ip, 
			quota_value_failed, description, external_id 
		FROM keys 
		WHERE id = $1 LIMIT 1`, id)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func loadKeyRecordByHash(ctx context.Context, db dbTx, hash string) (*keyRecord, error) {
	var res keyRecord
	err := db.GetContext(ctx, &res, `
		SELECT id, project, manual, quota_limit, quota_value, valid_to, disabled, 
			ip_white_list, tags, created, updated, last_used, last_ip, 
			quota_value_failed, description, external_id 
		FROM keys 
		WHERE key_hash = $1 AND
			manual = TRUE		
		LIMIT 1`, hash)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

const (
	_saveRequestTag = "x-tts-collect-data:always"
)

// GetKeyID returns keyID by key value
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

// Change generates new key for keyID, disables the old one, returns new key
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

// Update updates key table fields
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

// Usage returns usage information for the key
// //}

// Changes returns changed keys information
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

func (r *CMSRepository) addQuota(ctx context.Context, db dbTx, id string, in *api.CreditsInput) (*keyRecord, error) {
	log.Ctx(ctx).Trace().Str("id", id).Str("operationID", in.OperationID).Float64("quota", in.Credits).Msg("Add credits")

	key, err := loadKeyRecord(ctx, db, id)
	if err != nil {
		return nil, fmt.Errorf("load key: %w", mapErr(err))
	}

	if in.Credits < 0 && key.Limit - in.Credits < key.QuotaValue {
		return nil, &api.ErrField{Field: "credits", Msg: "(limit - change) is less than used"}
	}

	now := time.Now()

	has, err := newOperation(ctx, db, &createOperationInput{opID: in.OperationID, key_id: id, date: now, quota_value: in.Credits, msg: "Add Credits"})
	if err != nil {
		return nil, err
	}
	if has {
		return key, utils.ErrOperationExists
	}

	var limit float64
	err = db.QueryRowContext(ctx, `
	UPDATE keys
	SET 
		quota_limit = quota_limit + $1, 
		updated = $2
	WHERE 
		id = $3
	RETURNING quota_limit
	`, in.Credits, now, id).Scan(&limit)
	if err != nil {
		return nil, fmt.Errorf("update key: %w", mapErr(err))
	}
	key.Limit = limit
	return key, nil
}

func mapToKey(keyR *keyRecord, key string) *api.Key {
	res := &api.Key{Service: keyR.Project,
		ID:            keyR.ID,
		ValidTo:       toTimePtr(&keyR.ValidTo),
		LastUsed:      toTimePtr(keyR.LastUsed),
		LastIP:        keyR.LastIP.String,
		TotalCredits:  keyR.Limit,
		UsedCredits:   keyR.QuotaValue,
		FailedCredits: keyR.QuotaValueFailed,
		Disabled:      keyR.Disabled,
		Created:       toTimePtr(&keyR.Created),
		Updated:       toTimePtr(&keyR.Updated),
		IPWhiteList:   keyR.IPWhiteList.String,
		SaveRequests:  mapToSaveRequests(keyR.Tags),
		Key:           key,
	}
	return res
}

func mapToSaveRequests(tags *[]string) bool {
	if tags == nil {
		return false
	}
	for _, s := range *tags {
		if s == _saveRequestTag {
			return true
		}
	}
	return false
}

func (r *CMSRepository) createKeyWithQuota(ctx context.Context, tx dbTx, in *api.CreateInput, key string) (*keyRecord, error) {
	log.Ctx(ctx).Trace().Str("id", in.ID).Str("operationID", in.OperationID).Str("service", in.Service).Msg("Create operation record")

	now := time.Now()
	hash := utils.HashKey(key)
	log.Ctx(ctx).Trace().Str("id", in.ID).Str("key", key).Msg("Create key record")
	_, err := tx.ExecContext(ctx, `
	INSERT INTO keys (id, project, key_hash, manual, quota_limit, valid_to, created, updated, disabled)
	VALUES ($1, $2, $3, TRUE, $4, $5, $6, $6, FALSE)
	`, in.ID, in.Service, hash, in.Credits, now.Add(r.defaultValidToDuration), now)
	if err != nil {
		return nil, fmt.Errorf("create key: %w", mapErr(err))
	}

	has, err := newOperation(ctx, tx, &createOperationInput{opID: in.OperationID, key_id: in.ID, date: now, quota_value: in.Credits, msg: "Create Key"})
	if err != nil {
		return nil, err
	}
	if has {
		return nil, utils.ErrOperationExists
	}

	return &keyRecord{
		ID:      in.ID,
		Project: in.Service,
		Limit:   in.Credits,
		ValidTo: now.Add(r.defaultValidToDuration),
		Created: now,
		Updated: now,
	}, nil
}

type createOperationInput struct {
	opID        string
	key_id      string
	date        time.Time
	quota_value float64
	msg         string
}

func newOperation(ctx context.Context, tx sqlx.ExecerContext, in *createOperationInput) (bool /*exists operation*/, error) {
	_, err := tx.ExecContext(ctx, `
	INSERT INTO operations (id, key_id, date, quota_value, msg)
	VALUES ($1, $2, $3, $4, $5)
	`, in.opID, in.key_id, in.date, in.quota_value, in.msg)
	if err != nil {
		if isDuplicate(err) {
			return true, nil
		}
		return false, fmt.Errorf("create operation: %w", mapErr(err))
	}
	return false, nil
}

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

func validateInput(input *api.CreateInput) error {
	// if input == nil {
	// 	return &api.ErrField{Field: "id", Msg: "missing"}
	// }
	// if strings.TrimSpace(input.ID) == "" {
	// 	return &api.ErrField{Field: "id", Msg: "missing"}
	// }
	// if strings.TrimSpace(input.OperationID) == "" {
	// 	return &api.ErrField{Field: "operationID", Msg: "missing"}
	// }
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
	if math.Abs(input.Credits) <= 0.1 {
		return &api.ErrField{Field: "credits", Msg: "less than 0.1"}
	}
	return nil
}
