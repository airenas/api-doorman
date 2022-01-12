package mongodb

import (
	"time"
)

type keyRecord struct {
	Key              string    `bson:"key"`
	Manual           bool      `bson:"manual"`
	ValidTo          time.Time `bson:"validTo,omitempty"`
	Limit            float64   `bson:"limit,omitempty"`
	QuotaValue       float64   `bson:"quotaValue"`
	QuotaValueFailed float64   `bson:"quotaValueFailed,omitempty"`
	Created          time.Time `bson:"created,omitempty"`
	Updated          time.Time `bson:"updated,omitempty"`
	LastUsed         time.Time `bson:"lastUsed,omitempty"`
	LastIP           string    `bson:"lastIP,omitempty"`
	Disabled         bool      `bson:"disabled,omitempty"`
	IPWhiteList      string    `bson:"IPWhiteList,omitempty"`
	Description      string    `bson:"description,omitempty"`
	Tags             []string  `bson:"tags,omitempty"`
}

type logRecord struct {
	Key          string    `bson:"key,omitempty"`
	URL          string    `bson:"url,omitempty"`
	QuotaValue   float64   `bson:"quotaValue,omitempty"`
	Date         time.Time `bson:"date,omitempty"`
	IP           string    `bson:"ip,omitempty"`
	Value        string    `bson:"value,omitempty"`
	Fail         bool      `bson:"fail,omitempty"`
	ResponseCode int       `bson:"responseCode,omitempty"`
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
}
