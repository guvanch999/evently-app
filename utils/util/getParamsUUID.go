package util

import (
	"github.com/google/uuid"
	"github.com/guvanch999/evently-app/utils/exceptions"
	"github.com/labstack/echo/v4"
)

func GetParamsUUID(c echo.Context, name string) (uuid.UUID, error) {
	userGuid, err := uuid.Parse(c.Param(name))
	if err != nil {
		return uuid.Nil, exceptions.NewResponseError(exceptions.ErrBadRequest, err)
	}

	return userGuid, nil
}
