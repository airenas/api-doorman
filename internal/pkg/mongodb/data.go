package mongodb

import (
	"time"
)

type keyRecord struct {
	Key              string    `bson:"key"`
	KeyID            string    `bson:"keyID,omitempty"`
	Manual           bool      `bson:"manual"`
	ValidTo          time.Time `bson:"validTo,omitempty"`
	Limit            float64   `bson:"limit,omitempty"`
	QuotaValue       float64   `bson:"quotaValue"`
	QuotaValueFailed float64   `bson:"quotaValueFailed,omitempty"`
	Created          time.Time `bson:"created,omitempty"`
	Updated          time.Time `bson:"updated,omitempty"`
	LastUsed         time.Time `bson:"lastUsed,omitempty"`
	ResetAt          time.Time `bson:"resetAt,omitempty"`
	LastIP           string    `bson:"lastIP,omitempty"`
	Disabled         bool      `bson:"disabled,omitempty"`
	IPWhiteList      string    `bson:"IPWhiteList,omitempty"`
	Description      string    `bson:"description,omitempty"`
	Tags             []string  `bson:"tags,omitempty"`
	ExternalID       string    `bson:"externalID,omitempty"`
}

type logRecord struct {
	Key          string    `bson:"key,omitempty"`
	KeyID        string    `bson:"keyID,omitempty"`
	URL          string    `bson:"url,omitempty"`
	QuotaValue   float64   `bson:"quotaValue,omitempty"`
	Date         time.Time `bson:"date,omitempty"`
	IP           string    `bson:"ip,omitempty"`
	Value        string    `bson:"value,omitempty"`
	Fail         bool      `bson:"fail,omitempty"`
	ResponseCode int       `bson:"responseCode,omitempty"`

	RequestID string `bson:"requestID,omitempty"`
	ErrorMsg  string `bson:"errorMsg,omitempty"`
}

type keyMapRecord struct {
	Key        string    `bson:"key"`
	ExternalID string    `bson:"externalID"`
	Project    string    `bson:"project"`
	Created    time.Time `bson:"created,omitempty"`
	Old        []oldKey  `bson:"old,omitempty"`
}

type oldKey struct {
	Key       string    `bson:"key"`
	ChangedOn time.Time `bson:"changedOn,omitempty"`
}

type operationRecord struct {
	Key         string    `bson:"key"`
	OperationID string    `bson:"operationID"`
	Date        time.Time `bson:"date,omitempty"`
	QuotaValue  float64   `bson:"quotaValue,omitempty"`
	Msg         string    `bson:"msg,omitempty"`
}

type settingsRecord struct {
	ResetStarted time.Time `bson:"resetStarted,omitempty"`
	NextReset    time.Time `bson:"nextReset,omitempty"`
	Updated      time.Time `bson:"updated,omitempty"`
}
