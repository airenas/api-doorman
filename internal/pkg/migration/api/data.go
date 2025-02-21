package api

import (
	"time"
)

type (
	Key struct {
		Key         string    `json:"key"`
		KeyID       string    `json:"keyID"`
		Manual      bool      `json:"manual"`
		ValidTo     time.Time `json:"validTo"`
		Limit       float64   `json:"limit"`
		QuotaValue  float64   `json:"quotaValue"`
		QuotaFailed float64   `json:"quotaFailed"`
		Created     time.Time `json:"created"`
		Updated     time.Time `json:"updated"`
		LastUsed    time.Time `json:"lastUsed"`
		LastIP      string    `json:"lastIP"`
		Description string    `json:"description"`
		IPWhiteList string    `json:"IPWhiteList"`
		Disabled    bool      `json:"disabled"`
		Tags        []string  `json:"tags"`
	}
)
