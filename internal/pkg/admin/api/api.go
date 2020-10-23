package api

import "time"

// Key structure for key data
type Key struct {
	Key     string
	ValidTo time.Time
	Limit   float64
}
