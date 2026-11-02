package util

import (
	"github.com/google/uuid"
	"github.com/guvanch999/evently-app/utils/exceptions"
	"github.com/labstack/echo/v4"
)

const userIdContextKey = "userId"

// SetUserId stores the authenticated user's id on the request context.
func SetUserId(c echo.Context, userId uuid.UUID) {
	c.Set(userIdContextKey, userId)
}

// GetUserId returns the authenticated user's id set by the JWT middleware.
func GetUserId(c echo.Context) (uuid.UUID, error) {
	userId, ok := c.Get(userIdContextKey).(uuid.UUID)
	if !ok || userId == uuid.Nil {
		return uuid.Nil, exceptions.ErrUnauthorized
	}
	return userId, nil
}
