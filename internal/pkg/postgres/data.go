package postgres

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type keyRecord struct {
	ID               string
	Project          string
	KeyHash          string `db:"key_hash"` // or IP
	Manual           bool
	ValidTo          time.Time `db:"valid_to"`
	Limit            float64   `db:"quota_limit"`
	QuotaValue       float64   `db:"quota_value"`
	QuotaValueFailed float64   `db:"quota_value_failed"`
	Created          time.Time
	Updated          time.Time
	LastUsed         *time.Time     `db:"last_used"`
	ResetAt          *time.Time     `db:"reset_at"`
	LastIP           sql.NullString `db:"last_ip"`
	Disabled         bool
	IPWhiteList      sql.NullString `db:"ip_white_list"`
	Description      sql.NullString
	Tags             pq.StringArray `db:"tags,omitempty"`
	ExternalID       sql.NullString `db:"external_id"`
}

type logRecord struct {
	KeyID        string `db:"key_id"`
	URL          string
	QuotaValue   float64 `db:"quota_value"`
	Date         time.Time
	IP           string
	Value        string
	Fail         bool
	ResponseCode int `db:"response_code"`

	RequestID string `db:"request_id"`
	ErrorMsg  string `db:"error_msg"`
}

// type settingsRecord struct {
// 	ResetStarted time.Time `bson:"resetStarted,omitempty"`
// 	NextReset    time.Time `bson:"nextReset,omitempty"`
// 	Updated      time.Time `bson:"updated,omitempty"`
// }
