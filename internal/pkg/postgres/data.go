package postgres

import (
	"database/sql"
	"time"
)

type keyRecord struct {
	ID               string
	Project          string
	Key              string
	Manual           bool
	ValidTo          time.Time `db:"valid_to"`
	Limit            float64   `db:"quota_limit"`
	QuotaValue       float64   `db:"quota_value"`
	QuotaValueFailed float64   `db:"quota_value_failed"`
	Created          time.Time
	Updated          time.Time
	LastUsed         time.Time `db:"last_used"`
	ResetAt          time.Time `db:"reset_at"`
	LastIP           string    `db:"last_ip"`
	Disabled         bool
	IPWhiteList      sql.NullString `db:"ip_white_list"`
	Description      string
	Tags             *[]string `db:"tags,omitempty"`
	ExternalID       string    `db:"external_id"`
}

// type logRecord struct {
// 	Key          string    `bson:"key,omitempty"`
// 	KeyID        string    `bson:"keyID,omitempty"`
// 	URL          string    `bson:"url,omitempty"`
// 	QuotaValue   float64   `bson:"quotaValue,omitempty"`
// 	Date         time.Time `bson:"date,omitempty"`
// 	IP           string    `bson:"ip,omitempty"`
// 	Value        string    `bson:"value,omitempty"`
// 	Fail         bool      `bson:"fail,omitempty"`
// 	ResponseCode int       `bson:"responseCode,omitempty"`

// 	RequestID string `bson:"requestID,omitempty"`
// 	ErrorMsg  string `bson:"errorMsg,omitempty"`
// }

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
