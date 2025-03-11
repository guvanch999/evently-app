package middleware

import (
	"github.com/guvanch999/evently-app/utils/exceptions"
	"github.com/labstack/echo/v4"
)

func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenHeader := c.Request().Header.Get("Authorization")
			if tokenHeader == "" {
				return exceptions.ErrBadRequest
			}
			// fixme write auth logic
			return next(c)
		}
	}
}
