package middleware

import (
	"strings"

	"github.com/guvanch999/evently-app/internal/services"
	"github.com/guvanch999/evently-app/utils/exceptions"
	"github.com/guvanch999/evently-app/utils/util"
	"github.com/labstack/echo/v4"
)

// AuthMiddleware requires a valid "Authorization: Bearer <access token>"
// header and stores the authenticated user's id on the request context.
func AuthMiddleware(tokens *services.TokenService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			scheme, token, found := strings.Cut(c.Request().Header.Get(echo.HeaderAuthorization), " ")
			if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
				return exceptions.ErrUnauthorized
			}

			userId, err := tokens.VerifyAccessToken(strings.TrimSpace(token))
			if err != nil {
				return exceptions.ErrUnauthorized
			}

			util.SetUserId(c, userId)
			return next(c)
		}
	}
}
