package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
	"github.com/airenas/api-doorman/internal/pkg/model"
	"github.com/airenas/api-doorman/internal/pkg/model/usage"
	"github.com/airenas/api-doorman/internal/pkg/randkey"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/airenas/api-doorman/internal/pkg/utils/tag"
	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

type CMSRepository struct {
	db         *sqlx.DB
	newKeySize int
	hasher     Hasher
}

const (
	_keyFields = `id, project, manual, quota_limit, 
	quota_value, valid_to, disabled, ip_white_list, tags, created, updated, 
	last_used, last_ip, quota_value_failed, description, external_id, adm_id`
)

func NewCMSRepository(ctx context.Context, db *sqlx.DB, keySize int, hasher Hasher) (*CMSRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	if keySize < 10 || keySize > 100 {
		return nil, fmt.Errorf("wrong keySize")
	}
	if hasher == nil {
		return nil, fmt.Errorf("hasher is nil")
	}
	f := CMSRepository{db: db, newKeySize: keySize, hasher: hasher}
	return &f, nil
}

func (r *CMSRepository) Create(ctx context.Context, user *model.User, in *api.CreateInput) (*api.Key, bool /*created*/, error) {
	if err := validateInput(in); err != nil {
		return nil, false, err
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, false, fmt.Errorf("begin transaction: %w", err)
	}
	defer rollback(tx)

	if err := user.ValidateProject(in.Service); err != nil {
		return nil, false, err
	}
	if err := user.ValidateTags(in.Tags); err != nil {
		return nil, false, err
	}
	validTo, err := user.ValidateDate(in.ValidTo)
	if err != nil {
		return nil, false, err
	}
	if err := r.validateQuota(ctx, tx, user, in.Credits); err != nil {
		return nil, false, err
	}

	if in.ID == "" {
		in.ID = ulid.Make().String()
	}
	key, err := randkey.Generate(r.newKeySize)
	if err != nil {
		return nil, false, fmt.Errorf("generate key: %w", err)
	}

	res, err := r.createKeyWithQuota(ctx, tx, user, in, key, validTo)
	if err != nil {
		return nil, false, err
	}
	if err := tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("commit transaction: %w", err)
	}
	return mapToKey(res, key), true, nil
}

func (r *CMSRepository) GetKey(ctx context.Context, user *model.User, id string) (*api.Key, error) {
	res, err := loadKeyRecord(ctx, r.db, id)
	if err != nil {
		return nil, mapErr(err)
	}
	if err := validateKeyAccess(user, res); err != nil {
		return nil, err
	}
	return mapToKey(res, ""), nil
}

func (r *CMSRepository) GetKeyID(ctx context.Context, user *model.User, id string) (*api.KeyID, error) {
	hash := r.hasher.HashKey(id)
	res, err := loadKeyRecordByHash(ctx, r.db, hash)
	if err != nil {
		return nil, mapErr(err)
	}
	if err := validateKeyAccess(user, res); err != nil {
		return nil, err
	}
	return &api.KeyID{ID: res.ID, Service: res.Project}, nil
}

func (r *CMSRepository) AddCredits(ctx context.Context, user *model.User, id string, in *api.CreditsInput) (*api.Key, error) {
	if err := validateCreditsInput(in); err != nil {
		return nil, err
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer rollback(tx)

	res, err := r.addQuota(ctx, tx, user, id, in)
	if err != nil {
		if errors.Is(err, model.ErrOperationExists) && res != nil {
			return mapToKey(res, ""), nil
		}
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}
	return mapToKey(res, ""), nil
}

func (r *CMSRepository) Update(ctx context.Context, user *model.User, id string, in *api.UpdateInput) (*api.Key, error) {
	if err := validateUpdate(in); err != nil {
		return nil, err
	}
	if _, err := user.ValidateDate(in.ValidTo); err != nil {
		return nil, err
	}
	if err := user.ValidateTags(in.Tags); err != nil {
		return nil, err
	}
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer rollback(tx)

	res, err := r.update(ctx, tx, user, id, in)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}
	return mapToKey(res, ""), nil
}

func (r *CMSRepository) Change(ctx context.Context, user *model.User, id string) (*api.Key, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer rollback(tx)

	key, err := randkey.Generate(r.newKeySize)
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	res, err := r.changeKey(ctx, tx, user, id, key)
	if err != nil {
		return nil, err
	}

	if err := validateKeyAccess(user, res); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}
	return mapToKey(res, key), nil
}

func validateKeyAccess(user *model.User, res *keyRecord) error {
	if err := user.ValidateID(res.AdminID.String); err != nil {
		return err
	}
	return user.ValidateProject(res.Project)
}

func (r *CMSRepository) Usage(ctx context.Context, user *model.User, id string, from *time.Time, to *time.Time, full bool) (*api.Usage, error) {
	log.Ctx(ctx).Debug().Str("id", id).Msg("Get logs")

	where, values := makeDatesFilter("date", 2 /*start id*/, from, to)
	sValues := make([]interface{}, 0, len(values)+1)
	sValues = append(sValues, id)
	sValues = append(sValues, values...)

	key, err := loadKeyRecord(ctx, r.db, id)
	if err != nil {
		return nil, mapErr(err)
	}
	if err := validateKeyAccess(user, key); err != nil {
		return nil, err
	}

	var res []*logRecord
	err = r.db.SelectContext(ctx, &res, `
		SELECT * 
		FROM logs 
		WHERE 
			key_id = $1 `+where+`			
		`, sValues...)
	if err != nil {
		return nil, mapErr(err)
	}
	log.Ctx(ctx).Debug().Int("count", len(res)).Msg("Got logs")
	apiRes := &api.Usage{
		Logs: make([]*api.Log, 0, len(res)),
	}
	for _, r := range res {
		if r.Fail {
			apiRes.FailedCredits += r.QuotaValue
		} else {
			apiRes.UsedCredits += r.QuotaValue
		}
		apiRes.RequestCount++
		if full {
			apiRes.Logs = append(apiRes.Logs, mapLog(r))
		}
	}
	return apiRes, nil
}

// func (r *CMSRepository) Changes(ctx context.Context, user *model.User, from *time.Time, projects []string) (*api.Changes, error) {
// 	log.Ctx(ctx).Debug().Msg("Get changes")

// 	where, values := makeDatesFilter("updated", 2 /*start id*/, from, nil)
// 	sValues := make([]interface{}, 0, len(values)+1)
// 	sValues = append(sValues, user.ID)
// 	sValues = append(sValues, values...)

// 	to := time.Now().Add(-time.Millisecond) // make sure we will not loose some updates, so add -1 ms
// 	var res []*keyRecord
// 	err := r.db.SelectContext(ctx, &res, `
// 		SELECT `+_keyFields+`
// 		FROM keys
// 		WHERE manual = TRUE AND
// 			adm_id = $1
// 		`+where+`
// 		`, sValues...)
// 	if err != nil {
// 		return nil, mapErr(err)
// 	}
// 	log.Ctx(ctx).Debug().Int("count", len(res)).Msg("Got keys")
// 	apiRes := &api.Changes{
// 		Data: make([]*api.Key, 0, len(res)),
// 		From: from,
// 		Till: &to,
// 	}
// 	for _, r := range res {
// 		apiRes.Data = append(apiRes.Data, mapToKey(r, ""))
// 	}
// 	return apiRes, nil
// }

func (r *CMSRepository) Stats(ctx context.Context, user *model.User, in *api.StatParams) ([]*api.Bucket, error) {
	log.Ctx(ctx).Trace().Any("data", in).Msg("Get stats")
	key, err := loadKeyRecord(ctx, r.db, in.ID)
	if err != nil {
		return nil, mapErr(err)
	}
	if err := validateKeyAccess(user, key); err != nil {
		return nil, err
	}
	tbl, dField, err := getStatsTableField(in.Type)
	if err != nil {
		return nil, err
	}

	where, values := makeDatesFilter(dField, 2 /*start id*/, in.From, in.To)
	sValues := make([]interface{}, 0, len(values)+1)
	sValues = append(sValues, in.ID)
	sValues = append(sValues, values...)

	var res []*bucketRecord
	err = r.db.SelectContext(ctx, &res, `
		SELECT `+dField+` as at,
			request_count, failed_quota, used_quota, failed_requests
		FROM `+tbl+`
		WHERE 
			key_id = $1
		`+where+`			
		`, sValues...)
	if err != nil {
		return nil, mapErr(err)
	}
	log.Ctx(ctx).Debug().Int("count", len(res)).Msg("Got buckets")
	apiRes := make([]*api.Bucket, 0, len(res))
	for _, r := range res {
		apiRes = append(apiRes, mapToBucket(r))
	}
	return apiRes, nil
}

func getStatsTableField(enum usage.Enum) (string /*table*/, string /*field*/, error) {
	switch enum {
	case usage.Daily:
		return "daily_logs", "day", nil
	case usage.Monthly:
		return "monthly_logs", "month", nil
	}
	return "", "", model.NewWrongFieldError("type", "wrong type")
}

func mapToBucket(r *bucketRecord) *api.Bucket {
	return &api.Bucket{
		At:             r.At,
		RequestCount:   r.RequestCount.V,
		UsedQuota:      r.UsedQuota.V,
		FailedQuota:    r.FailedQuota.V,
		FailedRequests: r.FailedRequests.V,
	}
}

func (r *CMSRepository) validateQuota(ctx context.Context, tx dbTx, user *model.User, credits float64) error {
	assigned, err := r.getAssignedQuota(ctx, tx, user.ID)
	if err != nil {
		return err
	}
	if assigned+credits > user.MaxLimit {
		return model.NewWrongFieldError("credits", fmt.Sprintf("over limit, max %f, assigned %f", user.MaxLimit, assigned))
	}
	return nil
}

func (r *CMSRepository) getAssignedQuota(ctx context.Context, tx dbTx, id string) (float64, error) {
	var res float64
	err := tx.GetContext(ctx, &res, `
		SELECT COALESCE(SUM(quota_limit), 0)
		FROM keys
		WHERE adm_id = $1
	`, id)
	if err != nil {
		return 0, fmt.Errorf("get all assigned quotas: %w", mapErr(err))
	}
	return res, nil
}

func makeDatesFilter(field string, startIndex int, from *time.Time, to *time.Time) (string, []interface{}) {
	var values []interface{}
	var res string
	if from != nil {
		res += fmt.Sprintf(" AND %s >= $%d", field, startIndex)
		values = append(values, *from)
		startIndex++
	}
	if to != nil {
		res += fmt.Sprintf(" AND %s < $%d", field, startIndex)
		values = append(values, *to)
	}
	return res, values
}

func loadKeyRecord(ctx context.Context, db dbTx, id string) (*keyRecord, error) {
	var res keyRecord
	err := db.GetContext(ctx, &res, `
		SELECT `+_keyFields+` 
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
		SELECT `+_keyFields+` 
		FROM keys 
		WHERE key_hash = $1 AND
			manual = TRUE		
		LIMIT 1`, hash)
	if err != nil {
		return nil, mapErr(err)
	}
	return &res, nil
}

func prepareKeyUpdates(in *api.UpdateInput, key *keyRecord) (string, []interface{}, error) {
	var values []interface{}
	var updates []string

	if in.Description != nil {
		updates = append(updates, "description")
		values = append(values, *in.Description)
	}
	if in.Disabled != nil {
		updates = append(updates, "disabled")
		values = append(values, *in.Disabled)
	}
	if in.IPWhiteList != nil {
		updates = append(updates, "ip_white_list")
		values = append(values, *in.IPWhiteList)
	}
	if in.ValidTo != nil {
		updates = append(updates, "valid_to")
		values = append(values, *in.ValidTo)
	}
	if len(in.Tags) > 0 {
		tags, err := mergeTags(key.Tags, in.Tags)
		if err != nil {
			return "", nil, fmt.Errorf("merge tags: %w", err)
		}
		updates = append(updates, "tags")
		values = append(values, tags)
	}
	if len(updates) == 0 {
		return "", nil, model.NewWrongFieldError("", "no updates")
	}
	return makeUpdateSQL(updates), values, nil
}

func mergeTags(old, update []string) ([]string, error) {
	all := make(map[string]string)
	for _, t := range old {
		k, v, err := tag.Parse(t)
		if err != nil {
			return nil, err
		}
		all[k] = v
	}
	for _, t := range update {
		k, v, err := tag.Parse(t)
		if err != nil {
			return nil, err
		}
		if v == "" {
			delete(all, k)
		} else {
			all[k] = v
		}
	}
	res := make([]string, 0, len(all))
	for k, v := range all {
		res = append(res, fmt.Sprintf("%s:%s", k, v))
	}
	return res, nil
}

func makeUpdateSQL(updates []string) string {
	res := make([]string, 0, len(updates))
	for i, u := range updates {
		res = append(res, fmt.Sprintf("%s = $%d", u, i+3)) // 1 and 2 are id and updated
	}
	return strings.Join(res, ", ")
}

func (r *CMSRepository) addQuota(ctx context.Context, db dbTx, user *model.User, id string, in *api.CreditsInput) (*keyRecord, error) {
	log.Ctx(ctx).Trace().Str("id", id).Str("operationID", in.OperationID).Float64("quota", in.Credits).Msg("Add credits")

	key, err := loadKeyRecord(ctx, db, id)
	if err != nil {
		return nil, fmt.Errorf("load key: %w", mapErr(err))
	}

	if err := validateKeyAccess(user, key); err != nil {
		return nil, err
	}

	if in.Credits > 0 {
		if err := r.validateQuota(ctx, db, user, in.Credits); err != nil {
			return nil, err
		}
	}

	if in.Credits < 0 && key.Limit+in.Credits < key.QuotaValue {
		return nil, model.NewWrongFieldError("credits", "(limit - change) is less than used")
	}

	now := time.Now()

	has, err := newOperation(ctx, db, &createOperationInput{opID: in.OperationID, keyID: id, date: now, quotaValue: in.Credits, msg: "Add Credits", opData: newOpData(user)})
	if err != nil {
		return nil, err
	}
	if has {
		return key, model.ErrOperationExists
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

func (r *CMSRepository) update(ctx context.Context, db dbTx, user *model.User, id string, in *api.UpdateInput) (*keyRecord, error) {
	log.Ctx(ctx).Trace().Str("id", id).Any("in", in).Msg("Update")

	now := time.Now()

	key, err := loadKeyRecord(ctx, db, id)
	if err != nil {
		return nil, fmt.Errorf("load key: %w", mapErr(err))
	}
	if err := validateKeyAccess(user, key); err != nil {
		return nil, err
	}

	updates, values, err := prepareKeyUpdates(in, key)
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
		return nil, model.ErrNoRecord
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
		Description:   keyR.Description.String,
		Manual:        keyR.Manual,
		Tags:          keyR.Tags,

		Key: key,
	}
	return res
}

func mapLog(log *logRecord) *api.Log {
	res := &api.Log{}
	res.Date = toTimePtr(&log.Date)
	res.Fail = log.Fail
	res.Response = log.ResponseCode
	res.IP = log.IP
	res.UsedCredits = log.QuotaValue
	return res
}

func (r *CMSRepository) createKeyWithQuota(ctx context.Context, tx dbTx, user *model.User, in *api.CreateInput, key string, validTo time.Time) (*keyRecord, error) {
	log.Ctx(ctx).Trace().Str("id", in.ID).Str("operationID", in.OperationID).Str("service", in.Service).Msg("Create operation record")

	now := time.Now()
	if in.OperationID != "" {
		has, err := validateOperation(ctx, tx, &createOperationInput{
			opID:       in.OperationID,
			keyID:      "",
			date:       now,
			quotaValue: in.Credits,
			msg:        "Create Key",
			opData:     newOpData(user),
		})
		if err != nil {
			return nil, err
		}
		if has {
			return nil, model.ErrOperationExists
		}
	} else {
		in.OperationID = ulid.Make().String()
	}

	hash := r.hasher.HashKey(key)
	log.Ctx(ctx).Trace().Str("id", in.ID).Str("key", key).Msg("Create key record")
	_, err := tx.ExecContext(ctx, `
	INSERT INTO keys (id, project, key_hash, manual, quota_limit, valid_to, created, updated, disabled, tags, description, adm_id, ip_white_list)
	VALUES ($1, $2, $3, TRUE, $4, $5, $6, $6, $7, $8, $9, $10, $11)
	`, in.ID, in.Service, hash, in.Credits, validTo, now, in.Disabled, in.Tags, in.Description, user.ID, in.IPWhiteList)
	if err != nil {
		return nil, fmt.Errorf("create key: %w", mapErr(err))
	}

	has, err := newOperation(ctx, tx, &createOperationInput{opID: in.OperationID, keyID: in.ID, date: now,
		quotaValue: in.Credits, msg: "Create Key", opData: newOpData(user)})
	if err != nil {
		return nil, err
	}
	if has {
		return nil, model.ErrOperationExists
	}
	return loadKeyRecord(ctx, tx, in.ID)
}

func (r *CMSRepository) changeKey(ctx context.Context, tx dbTx, user *model.User, id string, key string) (*keyRecord, error) {
	log.Ctx(ctx).Trace().Str("id", id).Msg("Change key")

	keyRec, err := loadKeyRecord(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if !keyRec.Manual {
		return nil, model.NewWrongFieldError("manual", "not manual key")
	}

	now := time.Now()
	hash := r.hasher.HashKey(key)
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
		return nil, model.ErrNoRecord
	}

	_, err = newOperation(ctx, tx, &createOperationInput{opID: ulid.Make().String(), keyID: id, date: now, quotaValue: 0, msg: "Change Key", opData: newOpData(user)})
	if err != nil {
		return nil, err
	}

	return keyRec, nil
}

func newOpData(user *model.User) *operationData {
	return &operationData{IP: user.CurrentIP, AdminID: user.ID}
}

type createOperationInput struct {
	opID       string
	keyID      string
	date       time.Time
	quotaValue float64
	msg        string
	opData     *operationData
}

func validateOperation(ctx context.Context, tx dbTx, in *createOperationInput) (bool /*exists operation*/, error) {
	log.Ctx(ctx).Trace().Any("data", in).Msg("Validate operation record")

	var opRec operationRecord
	err := tx.GetContext(ctx, &opRec, `
	SELECT id, key_id, quota_value, msg
	FROM operations
	WHERE id = $1
	`, in.opID)
	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("get operation: %w", mapErr(err))
	} else if err == nil {
		if (in.keyID == "" || opRec.KeyID == in.keyID) && opRec.QuotaValue == in.quotaValue && opRec.Msg.String == in.msg {
			return true, nil
		}
		return false, model.ErrOperationDiffers
	}
	return false, nil
}

func newOperation(ctx context.Context, tx dbTx, in *createOperationInput) (bool /*exists operation*/, error) {
	log.Ctx(ctx).Trace().Any("data", in).Msg("Create operation record")
	res, err := validateOperation(ctx, tx, in)
	if err != nil || res {
		return res, err
	}
	_, err = tx.ExecContext(ctx, `
	INSERT INTO operations (id, key_id, date, quota_value, msg, data)
	VALUES ($1, $2, $3, $4, $5, $6)
	`, in.opID, in.keyID, in.date, in.quotaValue, in.msg, in.opData)
	if err != nil {
		if isDuplicate(err) {
			return true, nil
		}
		return false, fmt.Errorf("create operation: %w", mapErr(err))
	}
	return false, nil
}

func validateInput(input *api.CreateInput) error {
	if strings.TrimSpace(input.Service) == "" {
		return model.NewWrongFieldError("service", "missing")
	}
	if input.ValidTo != nil && input.ValidTo.Before(time.Now()) {
		return model.NewWrongFieldError("validTo", "past date")
	}
	if input.Credits <= 0.1 {
		return model.NewWrongFieldError("credits", "less than 0.1")
	}
	if input.IPWhiteList != "" {
		if err := utils.ValidateIPsCIDR(input.IPWhiteList); err != nil {
			return model.NewWrongFieldError("IPWhiteList", "wrong IP CIDR format")
		}
	}
	return validateTags(input.Tags)
}

func validateTags(tags []string) error {
	for _, t := range tags {
		_, _, err := tag.Parse(t)
		if err != nil {
			return model.NewWrongFieldError("tags", fmt.Sprintf("wrong tag: %s", t))
		}
	}
	return nil
}

func validateUpdate(input *api.UpdateInput) error {
	if input.ValidTo != nil && input.ValidTo.Before(time.Now()) {
		return model.NewWrongFieldError("validTo", "past date")
	}
	if input.IPWhiteList != nil {
		if err := utils.ValidateIPsCIDR(*input.IPWhiteList); err != nil {
			return model.NewWrongFieldError("IPWhiteList", "wrong IP CIDR format")
		}
	}
	return validateTags(input.Tags)
}

func validateCreditsInput(input *api.CreditsInput) error {
	if input == nil {
		return model.NewWrongFieldError("operationID", "missing")
	}
	if strings.TrimSpace(input.OperationID) == "" {
		return model.NewWrongFieldError("operationID", "missing")
	}
	if math.Abs(input.Credits) <= 0.1 {
		return model.NewWrongFieldError("credits", "less than 0.1")
	}
	return nil
}
