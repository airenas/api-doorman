package mongodb

import "time"

type keyRecord struct {
	Key              string    `json:"key,omitempty"`
	ValidTo          time.Time `json:"validTo"`
	Limit            float64   `json:"limit,omitempty"`
	QuotaValue       float64   `json:"quotaValue,omitempty"`
	QuotaValueFailed float64   `json:"quotaValueFailed,omitempty"`
	Created          time.Time `json:"created,omitempty"`
	LastUsed         time.Time `json:"lastUsed,omitempty"`
	LastIP           string    `json:"lastIP,omitempty"`
}
