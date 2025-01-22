package model

import (
	"fmt"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/model/permission"
	"github.com/airenas/api-doorman/internal/pkg/utils/tag"
)

type User struct {
	ID          string
	Name        string
	Disabled    bool
	MaxValidTo  time.Time
	MaxLimit    float64
	Projects    []string
	Permissions map[permission.Enum]bool
	AllowedTags map[string]bool
	CurrentIP   string
}

func (u *User) ValidateDate(to *time.Time) (time.Time, error) {
	if to == nil {
		return u.MaxValidTo, nil
	}
	if to != nil && (*to).After(u.MaxValidTo) {
		return time.Time{}, NewWrongFieldError("valid_to", fmt.Sprintf("must be before %s", u.MaxValidTo.Format(time.RFC3339)))
	}
	return *to, nil
}

func (u *User) ValidateProject(p string) error {
	if u.HasPermission(permission.Everything) {
		return nil
	}
	for _, project := range u.Projects {
		if project == p {
			return nil
		}
	}
	return NewNoAccessError("project", p)
}

func (u *User) ValidateID(id string) error {
	if u.HasPermission(permission.Everything) {
		return nil
	}
	if u.ID != id {
		return ErrNoAccess
	}
	return nil
}

func (u *User) ValidateTags(tags []string) error {
	if u.HasPermission(permission.Everything) {
		return nil
	}
	for _, t := range tags {
		k, _, err := tag.Parse(t)
		if err != nil {
			return err
		}
		if !u.AllowedTags[k] {
			return NewNoAccessError("tag", k)
		}
	}
	return nil
}

func (u *User) HasPermission(perm permission.Enum) bool {
	return u.Permissions != nil && (u.Permissions[perm] || u.Permissions[permission.Everything])
}
