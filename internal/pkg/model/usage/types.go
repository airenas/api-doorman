//go:generate stringer -type=Enum
package usage

import (
	"fmt"
	"strings"
)

type Enum int

const (
	Unknown Enum = iota
	Monthly
	Daily
)

func Parse(s string) (Enum, error) {
	switch strings.ToLower(s) {
	case "monthly":
		return Monthly, nil
	case "daily":
		return Daily, nil
	default:
		return Unknown, fmt.Errorf("invalid: %s", s)
	}
}
