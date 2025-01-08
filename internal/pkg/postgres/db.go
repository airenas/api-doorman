package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type dbTx interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// NewKeyValidator creates KeyValidator instance
func NewDB(ctx context.Context, dsn string) (*sqlx.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty DSN")
	}
	log.Ctx(ctx).Debug().Msg("Connecting to DB")
	sqlxDB, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to DB: %w", err)
	}
	sqlxDB.SetMaxOpenConns(10)
	sqlxDB.SetMaxIdleConns(10)
	if err := sqlxDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping DB: %w", err)
	}
	log.Ctx(ctx).Debug().Msg("Connected to DB")
	return sqlxDB, nil
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if err == sql.ErrNoRows {
		return fmt.Errorf("%w: %w", utils.ErrNoRecord, err)
	}
	if isDuplicate(err) {
		return fmt.Errorf("%w: %w", utils.ErrDuplicate, err)
	}
	return err
}

func isDuplicate(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func toTimePtr(time *time.Time) *time.Time {
	if time == nil || time.IsZero() {
		return nil
	}
	return time
}
