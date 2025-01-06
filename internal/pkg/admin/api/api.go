package api

import (
	"time"
)

// Key structure for key data
type Key struct {
	ID          string     `json:"ID,omitempty"`
	Key         string     `json:"key,omitempty"`
	Manual      bool       `json:"manual,omitempty"`
	ValidTo     *time.Time `json:"validTo,omitempty"`
	Limit       float64    `json:"limit,omitempty"`
	QuotaValue  float64    `json:"quotaValue,omitempty"`
	QuotaFailed float64    `json:"quotaFailed,omitempty"`
	Created     *time.Time `json:"created,omitempty"`
	Updated     *time.Time `json:"updated,omitempty"`
	LastUsed    *time.Time `json:"lastUsed,omitempty"`
	LastIP      string     `json:"lastIP,omitempty"`
	IPWhiteList string     `json:"IPWhiteList,omitempty"`
	Disabled    bool       `json:"disabled,omitempty"`
	Description string     `json:"description,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
}

// Log structure for log data
type Log struct {
	KeyID        string    `json:"keyID,omitempty"`
	URL          string    `json:"url,omitempty"`
	QuotaValue   float64   `json:"quotaValue,omitempty"`
	Date         time.Time `json:"date,omitempty"`
	IP           string    `json:"ip,omitempty"`
	Value        string    `json:"value,omitempty"`
	Fail         bool      `json:"fail,omitempty"`
	ResponseCode int       `json:"response,omitempty"`
	RequestID    string    `json:"requestID,omitempty"`
	ErrorMsg     string    `json:"errorMsg,omitempty"`
}

// KeyInfoResp keep key and logs data
type KeyInfoResp struct {
	Key  *Key   `json:"key,omitempty"`
	Logs []*Log `json:"logs,omitempty"`
}
