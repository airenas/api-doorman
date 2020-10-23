package api

import "time"

// Key structure for key data
type Key struct {
	Key         string    `json:"key,omitempty"`
	ValidTo     time.Time `json:"validTo"`
	Limit       float64   `json:"limit,omitempty"`
	QuotaValue  float64   `json:"quotaValue,omitempty"`
	QuotaFailed float64   `json:"quotaFailed,omitempty"`
	Created     time.Time `json:"created,omitempty"`
	LastUsed    time.Time `json:"lastUsed,omitempty"`
	LastIP      string    `json:"lastIP,omitempty"`
}
