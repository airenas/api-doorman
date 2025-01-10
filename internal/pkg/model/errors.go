package model

import (
	"errors"
	"fmt"
)

// ErrNoRecord indicates no record found error
var ErrNoRecord = errors.New("no record found")

// ErrLogRestored indicates conflict call for restoring usage by requestID
var ErrLogRestored = errors.New("already restored")

// ErrDuplicate indicates duplicate key record
var ErrDuplicate = errors.New("duplicate record")

// ErrOperationExists indicates existing operation for the record
var ErrOperationExists = errors.New("operation exists")

var ErrUnauthorized = errors.New("unauthorized")
var ErrNoAccess = errors.New("no access")

type NoAccessError struct {
	Resource string
}

type WrongFieldError struct {
	Field   string
	Message string
}

func (e *WrongFieldError) Error() string {
	return fmt.Sprintf("wrong %s: %s", e.Field, e.Message)
}

func NewWrongFieldError(field, message string) *WrongFieldError {
	return &WrongFieldError{Field: field, Message: message}
}

func (e *NoAccessError) Error() string {
	return fmt.Sprintf("no access to %s", e.Resource)
}

func NewNoAccessError(resource, name string) *NoAccessError {
	return &NoAccessError{Resource: fmt.Sprintf("%s %s", resource, name)}
}
