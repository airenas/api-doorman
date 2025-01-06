package postgres

import (
	"database/sql"
	"time"
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
	Tags             *[]string      `db:"tags,omitempty"`
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

// type keyMapRecord struct {
// 	KeyID      string    `bson:"keyID"`
// 	KeyHash    string    `bson:"keyHash"`
// 	ExternalID string    `bson:"externalID"`
// 	Project    string    `bson:"project"`
// 	Created    time.Time `bson:"created,omitempty"`
// 	Old        []oldKey  `bson:"old,omitempty"`
// }

// type oldKey struct {
// 	KeyHash   string    `bson:"keyHash"`
// 	ChangedOn time.Time `bson:"changedOn,omitempty"`
// }

// type operationRecord struct {
// 	KeyID       string    `bson:"keyID"`
// 	OperationID string    `bson:"operationID"`
// 	Date        time.Time `bson:"date,omitempty"`
// 	QuotaValue  float64   `bson:"quotaValue,omitempty"`
// 	Msg         string    `bson:"msg,omitempty"`
// }

// type settingsRecord struct {
// 	ResetStarted time.Time `bson:"resetStarted,omitempty"`
// 	NextReset    time.Time `bson:"nextReset,omitempty"`
// 	Updated      time.Time `bson:"updated,omitempty"`
// }
