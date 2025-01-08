package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/google/uuid"
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

func (r *AdmimRepository) Reset(ctx context.Context, project string, since time.Time, limit float64) error {
	log.Ctx(ctx).Debug().Str("project", project).Str("since", since.Format(time.RFC3339)).Float64("limit", limit).Msg("reset usage")

	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer roolback(tx)

	settings, err := getProjectSetting(ctx, tx, project)
	if err != nil {
		return fmt.Errorf("get project settings: %w", mapErr(err))
	}

	if since.Before(settings.NextReset) {
		log.Info().Str("since", since.Format(time.RFC3339)).Str("next", settings.NextReset.Format(time.RFC3339)).Msg("skip reset")
		return nil
	}

	items, err := getResetableItems(ctx, tx, project, since)
	if err != nil {
		return fmt.Errorf("get items: %w", err)
	}
	ua, ta := 0, 0.0
	log.Info().Int("len", len(items)).Msg("items to check")
	settings.NextReset = utils.StartOfMonth(since, 1)
	for _, it := range items {
		if it.ResetAt != nil && !since.After(*it.ResetAt) {
			continue
		}
		if !since.After(it.Created) {
			continue
		}
		cv := it.Limit - it.QuotaValue
		if cv >= limit {
			continue
		}
		if utils.Float64Equal(cv, limit) {
			continue
		}
		ua++
		ta += limit - cv
		err = reset(ctx, tx, it.ID, settings.NextReset, limit-cv)
		if err != nil {
			return fmt.Errorf("reset: %w", err)
		}
	}

	now := time.Now()
	log.Info().Int("items", ua).Float64("quota_total", ta).Msg("updated quota")

	settings.ResetStarted = now
	settings.Updated = now
	settings.Project = project
	if err := updateProjectSetting(ctx, tx, project, settings); err != nil {
		return mapErr(err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return err
}

func reset(ctx context.Context, tx dbTx, id string, at time.Time, quotaInc float64) error {
	log.Ctx(ctx).Debug().Str("id", id).Time("at", at).Float64("quotaInc", quotaInc).Msg("reset usage")

	_, err := newOperation(ctx, tx, &createOperationInput{opID: uuid.NewString(), key_id: id, date: time.Now(), quota_value: quotaInc, msg: "monthly reset"})
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE keys
		SET updated = $1,
			reset_at = $2,
			quota_limit = quota_limit + $3	
		WHERE
			id = $4
		`, time.Now(), at, quotaInc, id)
	if err != nil {
		return fmt.Errorf("update key record: %w", err)
	}
	return nil
}

func getResetableItems(ctx context.Context, tx dbTx, project string, at time.Time) ([]*keyRecord, error) {
	var res []*keyRecord
	err := tx.SelectContext(ctx, &res, `
		SELECT `+_keyFields+` 
		FROM keys 
		WHERE manual = FALSE AND
			project = $1 AND
			created < $2 AND
			(reset_at IS NULL OR reset_at < $2)
		`, project, at)
	if err != nil {
		return nil, mapErr(err)
	}
	return res, nil
}

func updateProjectSetting(ctx context.Context, tx dbTx, project string, settings *ProjectSettings) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO settings 
			(id, data, updated)
		VALUES 
			($1, $2, $3)
		ON CONFLICT (id) DO UPDATE
		SET 
			data = EXCLUDED.data,
			updated = EXCLUDED.updated
		`, resetSettingKey(project), settings, time.Now())
	if err != nil {
		return fmt.Errorf("update project settings: %w", err)
	}
	return nil
}

func getProjectSetting(ctx context.Context, tx dbTx, project string) (*ProjectSettings, error) {
	var data json.RawMessage
	err := tx.QueryRowContext(ctx, `
		SELECT data
		FROM settings
		WHERE id = $1
		`, resetSettingKey(project)).Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return &ProjectSettings{Project: project}, nil
		}
		return nil, fmt.Errorf("get project settings: %w", err)
	}
	var res ProjectSettings
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, fmt.Errorf("unmarshal settings: %w", err)
	}
	return &res, nil
}

func resetSettingKey(project string) string {
	return "reset-" + project
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
