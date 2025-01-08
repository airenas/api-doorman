package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/postgres"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
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

func InsertIPKey(t *testing.T, db *sqlx.DB, project string) string {
	t.Helper()

	id := uuid.New().String()
	ip := uuid.New().String()
	_, err := db.Exec(`INSERT INTO keys (id, project, key_hash, quota_limit, manual, valid_to) VALUES ($1, $2, $3, 10000, FALSE, $4)`, id, project, ip, time.Now().AddDate(0, 0, 1))
	require.NoError(t, err)
	return id
}
