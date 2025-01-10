package utils

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// ParseDateParam parse query param or returns echo error
func ParseDateParam(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	res, err := time.Parse(time.RFC3339, s)
	if err != nil {
		log.Error().Err(err).Str("str", s).Msg("can't parse as date")
		return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("can't parse as date '%s'", s))
	}
	return &res, nil
}

// StartOfMonth returns the time when month starts
func StartOfMonth(now time.Time, next int) time.Time {
	y, m := 0, int(now.Month())+next
	if m > 12 {
		y, m = 1, 1
	}
	return time.Date(now.Year()+y, time.Month(m), 1, 0, 0, 0, 0, now.Location())
}
