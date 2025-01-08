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
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type CMSRepository struct {
	db                     *sqlx.DB
	newKeySize             int
	defaultValidToDuration time.Duration
}

// Changes implements cms.Integrator.
func (c *CMSRepository) Changes(ctx context.Context, from *time.Time, projects []string) (*api.Changes, error) {
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

func (r *CMSRepository) Update(ctx context.Context, id string, in map[string]interface{}) (*api.Key, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer roolback(tx)

	res, err := r.update(ctx, tx, id, in)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}
	return mapToKey(res, ""), nil
}

func (r *CMSRepository) Change(ctx context.Context, id string) (*api.Key, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer roolback(tx)

	key, err := randkey.Generate(r.newKeySize)
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	res, err := r.changeKey(ctx, tx, id, key)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}
	return mapToKey(res, key), nil
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
		return nil, mapErr(err)
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
		return nil, mapErr(err)
	}
	return &res, nil
}

const (
	_saveRequestTag = "x-tts-collect-data:always"
)

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

func prepareKeyUpdates(input map[string]interface{}, now time.Time) (string, []interface{}, error) {
	var values []interface{}
	var updates []string

	for k, v := range input {
		var err error
		ok := true
		if k == "validTo" {
			var uv time.Time
			uv, ok = v.(time.Time)
			if !ok {
				uv, err = time.Parse(time.RFC3339, v.(string))
				ok = err == nil
			}
			if ok {
				ok = uv.After(now)
				if !ok {
					return "", nil, utils.NewWrongFieldError(k, "past date")
				}
				updates = append(updates, "valid_to")
				values = append(values, uv)
			}
		} else if k == "disabled" {
			var uv bool
			uv, ok = v.(bool)
			if ok {
				updates = append(updates, "disabled")
				values = append(values, uv)
			}
		} else if k == "IPWhiteList" {
			var s string
			s, ok = v.(string)
			if ok {
				if err := utils.ValidateIPsCIDR(s); err != nil {
					return "", nil, utils.NewWrongFieldError(k, "wrong IP CIDR format")
				}
				updates = append(updates, "ip_white_list")
				values = append(values, s)
			}
		} else if k == "description" {
			var uv string
			uv, ok = v.(string)
			if ok {
				updates = append(updates, "description")
				values = append(values, uv)
			}
		} else {
			err = utils.NewWrongFieldError(k, "unknown or unsuported update")
		}
		if !ok || err != nil {
			return "", nil, utils.NewWrongFieldError(k, "can't parse")
		}
	}
	if len(updates) == 0 {
		return "", nil, utils.NewWrongFieldError("", "no updates")
	}
	return makeUpdateSQL(updates), values, nil
}

func makeUpdateSQL(updates []string) string {
	res := make([]string, 0, len(updates))
	for i, u := range updates {
		res = append(res, fmt.Sprintf("%s = $%d", u, i+3)) // 1 and 2 are id and updated
	}
	return strings.Join(res, ", ")
}

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

	if in.Credits < 0 && key.Limit+in.Credits < key.QuotaValue {
		return nil, utils.NewWrongFieldError("credits", "(limit - change) is less than used")
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

func (r *CMSRepository) update(ctx context.Context, db dbTx, id string, in map[string]interface{}) (*keyRecord, error) {
	log.Ctx(ctx).Trace().Str("id", id).Any("in", in).Msg("Update")

	now := time.Now()
	updates, values, err := prepareKeyUpdates(in, now)
	if err != nil {
		return nil, err
	}
	upValues := make([]interface{}, 0, len(values)+1)
	upValues = append(upValues, id)
	upValues = append(upValues, now)
	upValues = append(upValues, values...)

	res, err := db.ExecContext(ctx, `
	UPDATE keys
	SET 
		`+updates+`,
		updated = $2
	WHERE 
		id = $1
	`, upValues...)
	if err != nil {
		return nil, fmt.Errorf("update key: %w", mapErr(err))
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return nil, utils.ErrNoRecord
	}
	return loadKeyRecord(ctx, db, id)
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
		Description:   keyR.Description.String,
		Key:           key,
	}
	return res
}

func mapToSaveRequests(tags []string) bool {
	for _, s := range tags {
		if s == _saveRequestTag {
			return true
		}
	}
	return false
}

func (r *CMSRepository) createKeyWithQuota(ctx context.Context, tx dbTx, in *api.CreateInput, key string) (*keyRecord, error) {
	log.Ctx(ctx).Trace().Str("id", in.ID).Str("operationID", in.OperationID).Str("service", in.Service).Msg("Create operation record")

	var tags []string
	if in.SaveRequests {
		tags = append(tags, _saveRequestTag)
	}

	now := time.Now()
	hash := utils.HashKey(key)
	log.Ctx(ctx).Trace().Str("id", in.ID).Str("key", key).Msg("Create key record")
	_, err := tx.ExecContext(ctx, `
	INSERT INTO keys (id, project, key_hash, manual, quota_limit, valid_to, created, updated, disabled, tags)
	VALUES ($1, $2, $3, TRUE, $4, $5, $6, $6, FALSE, $7)
	`, in.ID, in.Service, hash, in.Credits, now.Add(r.defaultValidToDuration), now, tags)
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
		Tags:    pq.StringArray(tags),
	}, nil
}

func (r *CMSRepository) changeKey(ctx context.Context, tx dbTx, id string, key string) (*keyRecord, error) {
	log.Ctx(ctx).Trace().Str("id", id).Msg("Change key")

	now := time.Now()
	hash := utils.HashKey(key)
	sRes, err := tx.ExecContext(ctx, `
	UPDATE keys 
	SET key_hash = $1, 
		updated = $2
	WHERE id = $3
	`, hash, now, id)
	if err != nil {
		return nil, fmt.Errorf("create key: %w", mapErr(err))
	}
	if rows, _ := sRes.RowsAffected(); rows == 0 {
		return nil, utils.ErrNoRecord
	}

	_, err = newOperation(ctx, tx, &createOperationInput{opID: uuid.NewString(), key_id: id, date: now, quota_value: 0, msg: "Change Key"})
	if err != nil {
		return nil, err
	}
	
	return loadKeyRecord(ctx, tx, id)
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
		return utils.NewWrongFieldError("service", "missing")
	}
	if input.ValidTo != nil && input.ValidTo.Before(time.Now()) {
		return utils.NewWrongFieldError("validTo", "past date")
	}
	if input.Credits <= 0.1 {
		return utils.NewWrongFieldError("credits", "less than 0.1")
	}
	return nil
}

func validateCreditsInput(input *api.CreditsInput) error {
	if input == nil {
		return utils.NewWrongFieldError("operationID", "missing")
	}
	if strings.TrimSpace(input.OperationID) == "" {
		return utils.NewWrongFieldError("operationID", "missing")
	}
	if math.Abs(input.Credits) <= 0.1 {
		return utils.NewWrongFieldError("credits", "less than 0.1")
	}
	return nil
}
