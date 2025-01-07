package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

// Repository communicates with postgres
type AdmimRepository struct {
	db *sqlx.DB
}

// DeleteLogs implements admin.LogProvider.
func (a *AdmimRepository) DeleteLogs(ctx context.Context, project string, to time.Time) (int, error) {
	panic("unimplemented")
}

// ListLogs implements admin.LogProvider.
func (a *AdmimRepository) ListLogs(ctx context.Context, project string, to time.Time) ([]*api.Log, error) {
	panic("unimplemented")
}

// Create implements admin.KeyCreator.
func (a *AdmimRepository) Create(ctx context.Context, project string, data *api.Key) (*api.Key, error) {
	panic("unimplemented")
}

// Reset implements reset.Reseter.
func (a *AdmimRepository) Reset(ctx context.Context, project string, since time.Time, limit float64) error {
	panic("unimplemented")
}

// Update implements admin.KeyUpdater.
func (a *AdmimRepository) Update(ctx context.Context, project string, id string, data map[string]interface{}) (*api.Key, error) {
	panic("unimplemented")
}

// List implements admin.KeyRetriever.
func (a *AdmimRepository) List(ctx context.Context, project string) ([]*api.Key, error) {
	panic("unimplemented")
}

func NewAdmimRepository(ctx context.Context, db *sqlx.DB) (*AdmimRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	f := AdmimRepository{db: db}
	return &f, nil
}

func (r *AdmimRepository) Get(ctx context.Context, project string, id string) (*api.Key, error) {
	log.Ctx(ctx).Debug().Str("id", id).Str("project", project).Msg("Get key")
	res, err := loadKeyRecord(ctx, r.db, id)
	if err != nil {
		return nil, mapErr(err)
	}
	return mapToAdminKey(res, ""), nil
}

func (r *AdmimRepository) GetLogs(ctx context.Context, project string, keyID string) ([]*api.Log, error) {
	log.Ctx(ctx).Debug().Str("id", keyID).Str("project", project).Msg("Get key")

	var res []*logRecord
	err := r.db.SelectContext(ctx, &res, `
		SELECT * 
		FROM logs 
		WHERE key_id = $1
		`, keyID)
	if err != nil {
		return nil, mapErr(err)
	}
	log.Ctx(ctx).Debug().Int("count", len(res)).Msg("Got logs")
	apiRes := make([]*api.Log, 0, len(res))
	for _, r := range res {
		apiRes = append(apiRes, mapToLog(r))
	}
	return apiRes, nil
}

func (r *AdmimRepository) RestoreUsage(ctx context.Context, project string, manual bool, request string, errorMsg string) error {
	log.Ctx(ctx).Debug().Str("requestID", request).Msg("Restoring usage")

	now := time.Now()
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer roolback(tx)

	var res logRecord
	err = tx.GetContext(ctx, &res, `
		SELECT * 
		FROM logs 
		WHERE request_id = $1
		`, request)
	if err != nil {
		return mapErr(err)
	}
	if res.Fail {
		return utils.ErrLogRestored
	}
	updateQuery := `
		UPDATE logs
		SET fail = TRUE,
			error_msg = $1 
		WHERE 
			request_id = $2 
	`
	_, err = tx.ExecContext(ctx, updateQuery, errorMsg, request)
	if err != nil {
		return fmt.Errorf("update logs: %w", mapErr(err))
	}

	log.Ctx(ctx).Trace().Str("id", res.KeyID).Float64("quota", res.QuotaValue).Msg("Restoring usage")

	updateQuery = `
		UPDATE keys
		SET updated = $1, 
			quota_value_failed = quota_value_failed + $2, 
			quota_value = quota_value - $2
		WHERE 
			id = $3
	`
	_, err = tx.ExecContext(ctx, updateQuery, now, res.QuotaValue, res.KeyID)
	if err != nil {
		return fmt.Errorf("update quota: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func mapToLog(v *logRecord) *api.Log {
	res := &api.Log{
		Date:         v.Date,
		Fail:         v.Fail,
		IP:           v.IP,
		URL:          v.URL,
		QuotaValue:   v.QuotaValue,
		Value:        v.Value,
		ResponseCode: v.ResponseCode,
		RequestID:    v.RequestID,
		ErrorMsg:     v.ErrorMsg,
	}
	return res
}

func mapToAdminKey(keyR *keyRecord, key string) *api.Key {
	res := &api.Key{
		ID:          keyR.ID,
		ValidTo:     toTimePtr(&keyR.ValidTo),
		LastUsed:    toTimePtr(keyR.LastUsed),
		LastIP:      keyR.LastIP.String,
		Limit:       keyR.Limit,
		QuotaValue:  keyR.QuotaValue,
		QuotaFailed: keyR.QuotaValueFailed,
		Disabled:    keyR.Disabled,
		Created:     toTimePtr(&keyR.Created),
		Updated:     toTimePtr(&keyR.Updated),
		IPWhiteList: keyR.IPWhiteList.String,
		Tags:        keyR.Tags,
		Key:         key,
	}
	return res
}
