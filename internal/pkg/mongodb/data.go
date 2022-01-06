package mongodb

import (
	"time"
)

type keyRecord struct {
	Key              string    `bson:"key,omitempty"`
	Manual           bool      `bson:"manual,omitempty"`
	ValidTo          time.Time `bson:"validto,omitempty"`
	Limit            float64   `bson:"limit,omitempty"`
	QuotaValue       float64   `bson:"quotavalue"`
	QuotaValueFailed float64   `bson:"quotavaluefailed,omitempty"`
	Created          time.Time `bson:"created,omitempty"`
	Updated          time.Time `bson:"updated,omitempty"`
	LastUsed         time.Time `bson:"lastused,omitempty"`
	LastIP           string    `bson:"lastip,omitempty"`
	Disabled         bool      `bson:"disabled,omitempty"`
	IPWhiteList      string    `bson:"ipwhitelist,omitempty"`
	Description      string    `bson:"description,omitempty"`
	Tags             []string  `bson:"tags,omitempty"`
}

type logRecord struct {
	Key          string    `bson:"key,omitempty"`
	URL          string    `bson:"url,omitempty"`
	QuotaValue   float64   `bson:"quotavalue,omitempty"`
	Date         time.Time `bson:"date,omitempty"`
	IP           string    `bson:"ip,omitempty"`
	Value        string    `bson:"value,omitempty"`
	Fail         bool      `bson:"fail,omitempty"`
	ResponseCode int       `bson:"response,omitempty"`
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
