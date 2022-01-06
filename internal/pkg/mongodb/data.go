package mongodb

import (
	"time"
)

type keyRecord struct {
	Key              string    `json:"key,omitempty"`
	Manual           bool      `json:"manual,omitempty"`
	ValidTo          time.Time `json:"validTo"`
	Limit            float64   `json:"limit,omitempty"`
	QuotaValue       float64   `json:"quotaValue"`
	QuotaValueFailed float64   `json:"quotaValueFailed,omitempty"`
	Created          time.Time `json:"created,omitempty"`
	Updated          time.Time `json:"updated,omitempty"`
	LastUsed         time.Time `json:"lastUsed,omitempty"`
	LastIP           string    `json:"lastIP,omitempty"`
	Disabled         bool      `json:"disabled,omitempty"`
	IPWhiteList      string    `json:"IPWhiteList,omitempty"`
	Description      string    `json:"description,omitempty"`
	Tags             []string  `json:"tags,omitempty"`
}

type logRecord struct {
	Key          string    `json:"key,omitempty"`
	URL          string    `json:"url,omitempty"`
	QuotaValue   float64   `json:"quotaValue,omitempty"`
	Date         time.Time `json:"date,omitempty"`
	IP           string    `json:"ip,omitempty"`
	Value        string    `json:"value,omitempty"`
	Fail         bool      `json:"fail,omitempty"`
	ResponseCode int       `json:"response,omitempty"`
}

type keyMapRecord struct {
	Key        string    `bson:"key"`
	ExternalID string    `bson:"externalID"`
	Project    string    `bson:"project"`
	Created    time.Time `bson:"created,omitempty"`
}

type operationRecord struct {
	Key         string    `bson:"key"`
	OperationID string    `bson:"operationID"`
	Date        time.Time `bson:"date,omitempty"`
	QuotaValue  float64   `bson:"quotaValue,omitempty"`
}
