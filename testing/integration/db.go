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

func FindKeyByIP(t *testing.T, db *sqlx.DB, project, ip string) string {
	t.Helper()

	var key string
	err := db.Get(&key, `SELECT id FROM keys WHERE project = $1 AND key_hash = $2`, project, ip)
	require.NoError(t, err)
	return key
}

func ResetKey(t *testing.T, db *sqlx.DB, id string) {
	t.Helper()

	_, err := db.Exec(`UPDATE keys SET reset_at = NULL WHERE id = $1`, id)
	require.NoError(t, err)
}
