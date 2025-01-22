package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/airenas/api-doorman/internal/pkg/model"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type (
	Auth interface {
		ValidateToken(ctx context.Context, token, ip string) (*model.User, error)
	}

	AuthMiddleware struct {
		auth Auth
	}
)

func NewAuthMiddleware(auth Auth) (*AuthMiddleware, error) {
	if auth == nil {
		return nil, fmt.Errorf("auth is nil")
	}
	return &AuthMiddleware{auth: auth}, nil
}

// Handle is the method that implements the middleware logic
func (a *AuthMiddleware) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		r := c.Request()
		ctx := r.Context()
		log.Ctx(ctx).Trace().Msg("Auth middleware")
		key, err := extractKey(r.Header.Get(authHeader))
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("can't extract key from header")
			return c.String(http.StatusUnauthorized, "Wrong auth header")
		}
		if key != "" {
			user, err := a.auth.ValidateToken(ctx, key, utils.ExtractIP(r))
			if err != nil {
				log.Ctx(ctx).Error().Err(err).Msg("can't validate token")
				if errors.Is(err, model.ErrUnauthorized) {
					return echo.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
				}
				return echo.NewHTTPError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			}
			log.Ctx(ctx).Trace().Str("user", user.Name).Msg("Logged as")
			c.SetRequest(r.WithContext(context.WithValue(ctx, model.CtxUser, user)))
		}
		return next(c)
	}
}
