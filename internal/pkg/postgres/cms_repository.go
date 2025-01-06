package postgres

import (
	"context"
	"fmt"
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

// AddCredits implements cms.Integrator.
func (c *CMSRepository) AddCredits(ctx context.Context, id string, in *api.CreditsInput) (*api.Key, error) {
	panic("unimplemented")
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

func loadKeyRecord(ctx context.Context, db* sqlx.DB, id string) (*keyRecord, error) {
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

func (r *CMSRepository) GetKeyID(ctx context.Context, id string) (*api.KeyID, error) {
	panic("unimplemented")
}

// // IsValid validates key
// func (r *Repository) IsValid(ctx context.Context, key string, IP string, manual bool) (bool, string, []string, error) {
// 	log.Debug().Msg("Validating key")

// 	var res keyRecord
// 	err := r.db.GetContext(ctx, &res, `
// 		SELECT id, disabled, valid_to, ip_white_list, tags
// 		FROM keys
// 		WHERE project = $1 AND
// 			key = $2 AND
// 			manual = $3`, r.project, key, manual)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			log.Info().Msg("No key")
// 			return false, "", nil, nil
// 		}
// 		return false, "", nil, fmt.Errorf("can't get key: %w", err)
// 	}
// 	ok, err := validateKey(&res, IP)
// 	if err != nil {
// 		return ok, "", nil, err
// 	}
// 	return ok, res.ID, toStrArray(res.Tags), nil
// }

// func toStrArray(in *[]string) []string {
// 	if in == nil {
// 		return nil
// 	}
// 	return *in
// }

// func validateKey(key *keyRecord, IP string) (bool, error) {
// 	if key.Disabled {
// 		log.Info().Msg("Key disabled")
// 		return false, nil
// 	}
// 	if !key.ValidTo.After(time.Now()) {
// 		log.Info().Msg("Key expired")
// 		return false, nil
// 	}
// 	res, err := utils.ValidateIPInWhiteList(key.IPWhiteList.String, IP)
// 	if !res {
// 		log.Info().Str("whiteList", key.IPWhiteList.String).Str("ip", IP).Msg("IP white list does not allow IP")
// 		if err != nil {
// 			log.Error().Err(err).Send()
// 		}
// 	}
// 	return res, err
// }

// // SaveValidate add qv to quota and validates with quota limit
// func (r *Repository) SaveValidate(ctx context.Context, key string, ip string, manual bool, qv float64) (bool, float64, float64, error) {
// 	log.Debug().Msg("Validating key")

// 	tx, err := r.db.Beginx()
// 	if err != nil {
// 		return false, 0, 0, fmt.Errorf("begin transaction: %w", err)
// 	}
// 	defer roolback(tx)

// 	var res keyRecord
// 	err = tx.GetContext(ctx, &res, `
// 		SELECT id, project, manual, quota_limit, quota_value
// 		FROM keys
// 		WHERE project = $1 AND
// 			key = $2 AND
// 			manual = $3 LIMIT 1`, r.project, key, manual)
// 	if err != nil {
// 		return false, 0, 0, err
// 	}

// 	remRequired := res.Limit - res.QuotaValue - qv
// 	if remRequired < 0 {
// 		return false, res.Limit - res.QuotaValue, res.Limit, r.updateFailed(ctx, tx, res.ID, ip, qv)
// 	}
// 	now := time.Now()
// 	var limit, quotaValue float64
// 	err = tx.QueryRowContext(ctx, `
// 		UPDATE keys
// 		SET last_used = $1,
// 			updated = $1,
// 			last_ip = $2,
// 			quota_value = quota_value + $3
// 		WHERE id = $4
// 		RETURNING quota_limit, quota_value`,
// 		now, ip, qv, res.ID).Scan(&limit, &quotaValue)
// 	if err != nil {
// 		return false, 0, 0, fmt.Errorf("update key record: %w", err)
// 	}
// 	if err := tx.Commit(); err != nil {
// 		return false, 0, 0, fmt.Errorf("commit transaction: %w", err)
// 	}
// 	remainingQuota := limit - quotaValue

// 	return true, remainingQuota, limit, nil
// }

// func roolback(tx *sqlx.Tx) {
// 	err := tx.Rollback()
// 	if err != nil && err != sql.ErrTxDone {
// 		log.Warn().Err(err).Msg("rollback failed")
// 	}
// }

// // Restore restores quota value after failed service call
// func (r *Repository) Restore(ctx context.Context, key string, manual bool, qv float64) (float64, float64, error) {
// 	log.Ctx(ctx).Debug().Float64("quota", qv).Msg("Restoring quota for key")

// 	now := time.Now()

// 	updateQuery := `
// 		UPDATE keys
// 		SET updated = $1,
// 			quota_value_failed = quota_value_failed + $2,
// 			quota_value = quota_value - $2
// 		WHERE
// 			project = $3 AND
// 			key = $4 AND
// 			manual = $5
// 		RETURNING quota_limit, quota_value
// 	`
// 	var limit, quotaValue float64
// 	err := r.db.QueryRowContext(ctx, updateQuery, now, qv, r.project, key, manual).Scan(&limit, &quotaValue)
// 	if err != nil {
// 		return 0, 0, fmt.Errorf("update quota: %w", err)
// 	}

// 	remainingQuota := limit - quotaValue
// 	return remainingQuota, limit, nil
// }

// func (r *Repository) CheckCreateIPKey(ctx context.Context, ip string, limit float64) (string, error) {
// 	log.Ctx(ctx).Debug().Str("ip", ip).Msg("Validating IP")

// 	tx, err := r.db.Beginx()
// 	if err != nil {
// 		return "", fmt.Errorf("begin transaction: %w", err)
// 	}
// 	defer roolback(tx)

// 	var res keyRecord
// 	err = tx.GetContext(ctx, &res, `
// 		SELECT id
// 		FROM keys
// 		WHERE project = $1 AND
// 			key = $2 AND
// 			manual = $3 LIMIT 1`, r.project, ip, false)
// 	if err == nil {
// 		return res.ID, nil
// 	}
// 	if err != sql.ErrNoRows {
// 		return "", err
// 	}

// 	id := uuid.NewString()
// 	log.Ctx(ctx).Debug().Str("ip", ip).Msg("insert new key for IP")
// 	_, err = tx.ExecContext(ctx, `
// 	INSERT INTO keys (id, project, key, manual, quota_limit, valid_to, created, updated)
// 	VALUES ($1, $2, $3, FALSE, $4, $5, $6, $6)
// 	`, id, r.project, ip, limit, time.Date(2100, time.Month(1), 1, 01, 0, 0, 0, time.UTC), time.Now())
// 	if err != nil {
// 		return "", fmt.Errorf("create key: %w", err)
// 	}
// 	if err := tx.Commit(); err != nil {
// 		return "", fmt.Errorf("commit transaction: %w", err)
// 	}
// 	return id, nil
// }

// func (r *Repository) SaveLog(ctx context.Context, data *api.Log) error {
// 	log.Ctx(ctx).Trace().Any("data", data).Msg("Insert log")

// 	_, err := r.db.ExecContext(ctx, `
// 	INSERT INTO logs (key_id, url, quota_value, date, ip, value, fail, response_code, request_id, error_msg)
// 	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
// 	`, data.KeyID, data.URL, data.QuotaValue, data.Date, data.IP, data.Value, data.Fail, data.ResponseCode, data.RequestID, data.ErrorMsg)
// 	if err != nil {
// 		return fmt.Errorf("insert log: %w", err)
// 	}
// 	return nil
// }

// func (r *Repository) updateFailed(ctx context.Context, db *sqlx.Tx, id, ip string, qv float64) error {
// 	now := time.Now()
// 	_, err := db.ExecContext(ctx, `
// 	UPDATE keys
// 	SET quota_value_failed = quota_value_failed + $1,
// 		last_ip = $2,
// 		last_used = $3,
// 		updated = $3
// 	WHERE id = $4
// 	`, qv, ip, now, id)
// 	if err != nil {
// 		return fmt.Errorf("failed to update key record: %w", err)
// 	}
// 	return nil
// }

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

// AddCredits to the key
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

func (r *CMSRepository) createKeyWithQuota(ctx context.Context, tx *sqlx.Tx, in *api.CreateInput, key string) (*keyRecord, error) {
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

	_, err = tx.ExecContext(ctx, `
	INSERT INTO operations (id, key_id, date, quota_value, msg)
	VALUES ($1, $2, $3, $4, $5)
	`, in.OperationID, in.ID, now, in.Credits, "Initial credits")
	if err != nil {
		if isDuplicate(err) {
			return nil, utils.ErrOperationExists
		}
		return nil, fmt.Errorf("create operation: %w", mapErr(err))
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
	if input.Credits <= 0.1 {
		return &api.ErrField{Field: "credits", Msg: "less than 0.1"}
	}
	return nil
}
