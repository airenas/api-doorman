package api

import (
	"errors"
	"time"
)

// Key structure for key data
type Key struct {
	Key         string     `json:"key,omitempty"`
	KeyID       string     `json:"keyID,omitempty"`
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
	Key          string    `json:"key,omitempty"`
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

//ErrNoRecord indicates no record found error
var ErrNoRecord = errors.New("no record found")

//ErrWrongField indicates wrong passed field on update
var ErrWrongField = errors.New("wrong field")

//ErrLogRestored indicates conflict call for restoring usage by requestID
var ErrLogRestored = errors.New("already restored")
