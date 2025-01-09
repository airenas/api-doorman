package model

import (
	"fmt"
	"time"
)

type User struct {
	ID         string
	Name       string
	Disabled   bool
	MaxValidTo time.Time
	MaxLimit   float64
	Projects   []string
}

func (u *User) ValidateDate(to *time.Time) (time.Time, error) {
	if to == nil {
		return u.MaxValidTo, nil
	}
	if to != nil && (*to).After(u.MaxValidTo) {
		return time.Time{}, NewWrongFieldError("max_valid_to", fmt.Sprintf("must be before %s", u.MaxValidTo.Format(time.RFC3339)))
	}
	return *to, nil
}

func (u *User) ValidateProject(p string) error {
	for _, project := range u.Projects {
		if project == p {
			return nil
		}
	}
	return NewNoAccessError("project", p)
}
