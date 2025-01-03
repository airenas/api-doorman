package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

// Repository communicates with postgres
type Repository struct {
	db      *sqlx.DB
	project string
}

func NewRepository(ctx context.Context, db *sqlx.DB, project string) (*Repository, error) {
	pr := strings.TrimSpace(project)
	if pr == "" {
		return nil, fmt.Errorf("project is empty")
	}
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	f := Repository{db: db, project: pr}
	return &f, nil
}

// IsValid validates key
func (r *Repository) IsValid(ctx context.Context, key string, IP string, manual bool) (bool, string, []string, error) {
	log.Debug().Msg("Validating key")

	var res keyRecord
	err := r.db.GetContext(ctx, &res, `
		SELECT id, disabled, valid_to, ip_white_list, tags 
		FROM keys 
		WHERE project = $1 AND 
			key = $2 AND 
			manual = $3`, r.project, key, manual)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info().Msg("No key")
			return false, "", nil, nil
		}
		return false, "", nil, fmt.Errorf("can't get key: %w", err)
	}
	ok, err := validateKey(&res, IP)
	if err != nil {
		return ok, "", nil, err
	}
	return ok, res.ID, toStrArray(res.Tags), nil
}

func toStrArray(in *[]string) []string {
	if in == nil {
		return nil
	}
	return *in
}

func validateKey(key *keyRecord, IP string) (bool, error) {
	if key.Disabled {
		log.Info().Msg("Key disabled")
		return false, nil
	}
	if !key.ValidTo.After(time.Now()) {
		log.Info().Msg("Key expired")
		return false, nil
	}
	res, err := utils.ValidateIPInWhiteList(key.IPWhiteList.String, IP)
	if !res {
		log.Info().Str("whiteList", key.IPWhiteList.String).Str("ip", IP).Msg("IP white list does not allow IP")
		if err != nil {
			log.Error().Err(err).Send()
		}
	}
	return res, err
}

// SaveValidate add qv to quota and validates with quota limit
func (r *Repository) SaveValidate(ctx context.Context, key string, ip string, manual bool, qv float64) (bool, float64, float64, error) {
	log.Debug().Msg("Validating key")

	tx, err := r.db.Beginx()
	if err != nil {
		return false, 0, 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer roolback(tx)

	var res keyRecord
	err = tx.GetContext(ctx, &res, `
		SELECT id, project, manual, quota_limit, quota_value 
		FROM keys 
		WHERE project = $1 AND 
			key = $2 AND 
			manual = $3 LIMIT 1`, r.project, key, manual)
	if err != nil {
		return false, 0, 0, err
	}

	remRequired := res.Limit - res.QuotaValue - qv
	if remRequired < 0 {
		return false, res.Limit - res.QuotaValue, res.Limit, r.updateFailed(ctx, tx, res.ID, ip, qv)
	}
	now := time.Now()
	var limit, quotaValue float64
	err = tx.QueryRowContext(ctx, `
		UPDATE keys 
		SET last_used = $1, 
			updated = $1, 
			last_ip = $2, 
			quota_value = quota_value + $3 
		WHERE id = $4
		RETURNING quota_limit, quota_value`,
		now, ip, qv, res.ID).Scan(&limit, &quotaValue)
	if err != nil {
		return false, 0, 0, fmt.Errorf("update key record: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, 0, 0, fmt.Errorf("commit transaction: %w", err)
	}
	remainingQuota := limit - quotaValue

	return true, remainingQuota, limit, nil
}

func roolback(tx *sqlx.Tx) {
	err := tx.Rollback()
	if err != nil && err != sql.ErrTxDone {
		log.Warn().Err(err).Msg("rollback failed")
	}
}

// Restore restores quota value after failed service call
func (r *Repository) Restore(ctx context.Context, key string, manual bool, qv float64) (float64, float64, error) {
	log.Ctx(ctx).Debug().Float64("quota", qv).Msg("Restoring quota for key")

	now := time.Now()

	updateQuery := `
		UPDATE keys
		SET updated = $1, 
			quota_value_failed = quota_value_failed + $2, 
			quota_value = quota_value - $2
		WHERE 
			project = $3 AND
			key = $4 AND 
			manual = $5
		RETURNING quota_limit, quota_value
	`
	var limit, quotaValue float64
	err := r.db.QueryRowContext(ctx, updateQuery, now, qv, r.project, key, manual).Scan(&limit, &quotaValue)
	if err != nil {
		return 0, 0, fmt.Errorf("update quota: %w", err)
	}

	remainingQuota := limit - quotaValue
	return remainingQuota, limit, nil
}

func (r *Repository) CheckCreateIPKey(ctx context.Context, ip string, limit float64) (string, error) {
	log.Ctx(ctx).Debug().Str("ip", ip).Msg("Validating IP")

	tx, err := r.db.Beginx()
	if err != nil {
		return "", fmt.Errorf("begin transaction: %w", err)
	}
	defer roolback(tx)

	var res keyRecord
	err = tx.GetContext(ctx, &res, `
		SELECT id 
		FROM keys 
		WHERE project = $1 AND 
			key = $2 AND 
			manual = $3 LIMIT 1`, r.project, ip, false)
	if err == nil {
		return res.ID, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}

	id := uuid.NewString()
	log.Ctx(ctx).Debug().Str("ip", ip).Msg("insert new key for IP")
	_, err = tx.ExecContext(ctx, `
	INSERT INTO keys (id, project, key, manual, quota_limit, valid_to, created, updated)
	VALUES ($1, $2, $3, FALSE, $4, $5, $6, $6)
	`, id, r.project, ip, limit, time.Date(2100, time.Month(1), 1, 01, 0, 0, 0, time.UTC), time.Now())
	if err != nil {
		return "", fmt.Errorf("create key: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit transaction: %w", err)
	}
	return id, nil
}

func (r *Repository) SaveLog(ctx context.Context, data *api.Log) error {
	log.Ctx(ctx).Trace().Any("data", data).Msg("Insert log")

	_, err := r.db.ExecContext(ctx, `
	INSERT INTO logs (key_id, url, quota_value, date, ip, value, fail, response_code, request_id, error_msg)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, data.KeyID, data.URL, data.QuotaValue, data.Date, data.IP, data.Value, data.Fail, data.ResponseCode, data.RequestID, data.ErrorMsg)
	if err != nil {
		return fmt.Errorf("insert log: %w", err)
	}
	return nil
}

func (r *Repository) updateFailed(ctx context.Context, db *sqlx.Tx, id, ip string, qv float64) error {
	now := time.Now()
	_, err := db.ExecContext(ctx, `
	UPDATE keys 
	SET quota_value_failed = quota_value_failed + $1, 
		last_ip = $2, 
		last_used = $3, 
		updated = $3 
	WHERE id = $4
	`, qv, ip, now, id)
	if err != nil {
		return fmt.Errorf("failed to update key record: %w", err)
	}
	return nil
}
