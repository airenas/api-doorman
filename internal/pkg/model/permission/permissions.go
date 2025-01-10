//go:generate stringer -type=Enum
package permission

import "fmt"

type Enum int

const (
	Unknown Enum = iota
	Everything
	RestoreUsage
	ResetMonthlyUsage
)

func Parse(s string) (Enum, error) {
	switch s {
	case "Everything":
		return Everything, nil
	case "RestoreUsage":
		return RestoreUsage, nil
	case "ResetMonthlyUsage":
		return ResetMonthlyUsage, nil
	default:
		return Unknown, fmt.Errorf("invalid Permission: %s", s)
	}
}
