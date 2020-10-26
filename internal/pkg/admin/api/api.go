package api

import "time"

// Key structure for key data
type Key struct {
	Key         string    `json:"key,omitempty"`
	Manual      bool      `json:"manual,omitempty"`
	ValidTo     time.Time `json:"validTo"`
	Limit       float64   `json:"limit,omitempty"`
	QuotaValue  float64   `json:"quotaValue,omitempty"`
	QuotaFailed float64   `json:"quotaFailed,omitempty"`
	Created     time.Time `json:"created,omitempty"`
	Updated     time.Time `json:"updated,omitempty"`
	LastUsed    time.Time `json:"lastUsed,omitempty"`
	LastIP      string    `json:"lastIP,omitempty"`
	Disabled    bool      `json:"disabled,omitempty"`
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
}
