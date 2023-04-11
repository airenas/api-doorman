package api

import (
	"errors"
	"fmt"
	"time"
)

// CreateInput for create key request
type CreateInput struct {
	ID           string     `json:"id,omitempty"`
	OperationID  string     `json:"operationID,omitempty"`
	Service      string     `json:"service,omitempty"`
	Credits      float64    `json:"credits,omitempty"`
	ValidTo      *time.Time `json:"validTo,omitempty"`
	SaveRequests bool       `json:"saveRequests,omitempty"`
}

// CreditsInput for add credits
type CreditsInput struct {
	OperationID string  `json:"operationID,omitempty"`
	Credits     float64 `json:"credits,omitempty"`
	Msg         string  `json:"msg,omitempty"`
}

// Key structure for key data
type Key struct {
	ID           string     `json:"id,omitempty"`
	Key          string     `json:"key,omitempty"`
	Service      string     `json:"service,omitempty"`
	ValidTo      *time.Time `json:"validTo,omitempty"`
	Disabled     bool       `json:"disabled,omitempty"`
	IPWhiteList  string     `json:"IPWhiteList,omitempty"`
	SaveRequests bool       `json:"saveRequests,omitempty"`

	TotalCredits  float64 `json:"totalCredits,omitempty"`
	UsedCredits   float64 `json:"usedCredits,omitempty"`
	FailedCredits float64 `json:"failedCredits,omitempty"`

	Created  *time.Time `json:"created,omitempty"`
	Updated  *time.Time `json:"updated,omitempty"`
	LastUsed *time.Time `json:"lastUsed,omitempty"`
	LastIP   string     `json:"lastIP,omitempty"`
}

// KeyID provides key ID by key, response structure
type KeyID struct {
	ID      string `json:"id,omitempty"`
	Service string `json:"service,omitempty"`
}

// Usage response
type Usage struct {
	RequestCount  int     `json:"requestCount"`
	UsedCredits   float64 `json:"usedCredits,omitempty"`
	FailedCredits float64 `json:"failedCredits,omitempty"`
	Logs          []*Log  `json:"logs,omitempty"`
}

// Changes response
type Changes struct {
	From *time.Time `json:"from,omitempty"`
	Till *time.Time `json:"till,omitempty"`
	Data []*Key     `json:"data,omitempty"`
}

// Log detailed usage record
type Log struct {
	UsedCredits float64    `json:"usedCredits,omitempty"`
	Date        *time.Time `json:"date,omitempty"`
	IP          string     `json:"ip,omitempty"`
	Fail        bool       `json:"fail,omitempty"`
	Response    int        `json:"response,omitempty"`
}

// ErrNoRecord indicates no record found error
var ErrNoRecord = errors.New("no record found")

// ErrOperationExists indicates existing operation for the record
var ErrOperationExists = errors.New("operation exists")

// ErrField error indicating input field problem
type ErrField struct {
	Field, Msg string
}

func (r *ErrField) Error() string {
	return fmt.Sprintf("wrong field '%s' - %s", r.Field, r.Msg)
}
