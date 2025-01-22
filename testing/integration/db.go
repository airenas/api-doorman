package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/postgres"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func NewDB() (*sqlx.DB, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("no DB_DSN set")
	}
	return postgres.NewDB(context.Background(), dsn)
}

func ResetSettings(t *testing.T, db *sqlx.DB, key string) {
	t.Helper()

	data := postgres.ProjectSettings{
		NextReset: time.Now().AddDate(0, -1, 0),
	}
	_, err := db.Exec(`UPDATE settings 
	SET data = $1
	WHERE id = $2`, data, key)
	require.NoError(t, err)
}

func InsertIPKey(t *testing.T, db *sqlx.DB, project string) string {
	t.Helper()

	id := ulid.Make().String()
	ip := ulid.Make().String()
	_, err := db.Exec(`INSERT INTO keys (id, project, key_hash, quota_limit, manual, valid_to) VALUES ($1, $2, $3, 10000, FALSE, $4)`, id, project, ip, time.Now().AddDate(0, 0, 1))
	require.NoError(t, err)
	return id
}

type InsertAdminParams struct {
	Projects    []string
	KeyHash     string
	Permissions []string
	MaxLimit    float64
	MaxValidTo  time.Time
	Disabled    bool
	IPWhiteList string
}

func InsertAdmin(t *testing.T, db *sqlx.DB, params *InsertAdminParams) {
	t.Helper()

	now := time.Now()
	_, err := db.Exec(`
		INSERT INTO administrators
			(id, key_hash, projects, max_valid_to, max_limit, name, created, updated, permissions, ip_white_list)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $7, $8, $9)
		`, ulid.Make().String(), params.KeyHash, pq.Array(params.Projects), params.MaxValidTo, params.MaxLimit, "test", now, pq.Array(params.Permissions), params.IPWhiteList)
	require.NoError(t, err)
}

// func RefreshView(t *testing.T, db *sqlx.DB, view string) {
// 	t.Helper()
// 	// need to add some months to the date to refresh monthly logs
// 	_, err := db.Exec(`CALL refresh_continuous_aggregate($1, '2025-01-01', $2::timestamptz)`, view, time.Now().AddDate(0, 2, 0))
// 	require.NoError(t, err)
// }
