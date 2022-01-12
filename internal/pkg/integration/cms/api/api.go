package api

import (
	"errors"
	"fmt"
	"time"
)

type CreateInput struct {
	ID           string     `json:"id,omitempty"`
	OperationID  string     `json:"operationID,omitempty"`
	Service      string     `json:"service,omitempty"`
	Credits      float64    `json:"credits,omitempty"`
	ValidTo      *time.Time `json:"validTo,omitempty"`
	SaveRequests bool       `json:"saveRequests,omitempty"`
}

type CreditsInput struct {
	OperationID string  `json:"operationID,omitempty"`
	Credits     float64 `json:"credits,omitempty"`
}

// Key structure for key data
type Key struct {
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

type KeyID struct {
	ID      string `json:"id,omitempty"`
	Service string `json:"service,omitempty"`
}

type Usage struct {
	RequestCount  int     `json:"requestCount"`
	UsedCredits   float64 `json:"usedCredits,omitempty"`
	FailedCredits float64 `json:"failedCredits,omitempty"`
	Logs          []*Log  `json:"logs,omitempty"`
}

type Log struct {
	UsedCredits float64    `json:"usedCredits,omitempty"`
	Date        *time.Time `json:"date,omitempty"`
	IP          string     `json:"ip,omitempty"`
	Fail        bool       `json:"fail,omitempty"`
	Response    int        `json:"response,omitempty"`
}

//ErrNoRecord indicates no record found error
var ErrNoRecord = errors.New("no record found")

type ErrField struct {
	Field, Msg string
}

func (r *ErrField) Error() string {
	return fmt.Sprintf("wrong field '%s' - %s", r.Field, r.Msg)
}
