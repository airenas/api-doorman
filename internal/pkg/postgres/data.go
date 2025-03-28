package postgres

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
)

type keyRecord struct {
	ID               string
	Project          string
	AdminID          sql.NullString `db:"adm_id"`
	KeyHash          string         `db:"key_hash"` // or IP
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

type operationRecord struct {
	ID         string
	KeyID      string  `db:"key_id"`
	QuotaValue float64 `db:"quota_value"`
	Date       time.Time
	Msg        sql.NullString
	Data       *operationData
}

type operationData struct {
	IP      string `json:"ip,omitempty"`
	AdminID string `json:"adm_id,omitempty"`
}

type ProjectSettings struct {
	Project      string
	ResetStarted time.Time `json:"resetStarted,omitempty"`
	NextReset    time.Time `json:"nextReset,omitempty"`
	Updated      time.Time `json:"updated,omitempty"`
}

type administratorRecord struct {
	ID          string
	Projects    pq.StringArray
	Permissions pq.StringArray
	KeyHash     string         `db:"key_hash"`
	MaxValidTo  time.Time      `db:"max_valid_to"`
	MaxLimit    float64        `db:"max_limit"`
	IPWhiteList sql.NullString `db:"ip_white_list"`
	AllowedTags pq.StringArray `db:"allowed_tags"`
	Name        string
	Disabled    bool
	Description sql.NullString
	Created     time.Time
	Updated     time.Time
}

type bucketRecord struct {
	At             time.Time         `db:"at"`
	RequestCount   sql.Null[int]     `db:"request_count"`
	FailedQuota    sql.Null[float64] `db:"failed_quota"`
	UsedQuota      sql.Null[float64] `db:"used_quota"`
	FailedRequests sql.Null[int]     `db:"failed_requests"`
}

// //////////////////////////
// conversion for operationData
// //////////////////////////
func (od operationData) Value() (driver.Value, error) {
	return json.Marshal(od)
}

func (od *operationData) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &od)
}
