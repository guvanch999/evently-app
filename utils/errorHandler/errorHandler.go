package errorHandler

import (
	"errors"
	"net/http"

	"github.com/guvanch999/evently-app/utils/exceptions"
	"github.com/labstack/echo/v4"
)

func ResponseHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	code := http.StatusInternalServerError
	var he *exceptions.ResponseError
	if errors.As(err, &he) {
		code = he.Code
	} else {
		he = exceptions.ErrInternalServerError
	}
	if responseErr := c.JSON(code, he); responseErr != nil {
		c.Logger().Error(responseErr)
	}
}
