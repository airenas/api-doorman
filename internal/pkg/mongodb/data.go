package mongodb

import "time"

type keyRecord struct {
	Key              string    `json:"key,omitempty"`
	Manual           bool      `json:"manual,omitempty"`
	ValidTo          time.Time `json:"validTo"`
	Limit            float64   `json:"limit,omitempty"`
	QuotaValue       float64   `json:"quotaValue,omitempty"`
	QuotaValueFailed float64   `json:"quotaValueFailed,omitempty"`
	Created          time.Time `json:"created,omitempty"`
	LastUsed         time.Time `json:"lastUsed,omitempty"`
	LastIP           string    `json:"lastIP,omitempty"`
	Disabled         bool      `json:"disabled,omitempty"`
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
