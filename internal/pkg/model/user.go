package model

import "time"

type User struct {
	ID         string
	Name       string
	Disabled   bool
	MaxValidTo time.Time
	MaxLimit   float64
	Projects   []string
}
