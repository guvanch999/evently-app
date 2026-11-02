package errorHandler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/guvanch999/evently-app/utils/exceptions"
	"github.com/labstack/echo/v4"
)

// ResponseHTTPErrorHandler renders every error as an exceptions.ResponseError.
// Echo's own errors (unknown route, bind/validation errors, rate limiting)
// keep their status code; anything unexpected becomes a 500 and is logged.
func ResponseHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var he *exceptions.ResponseError
	var echoErr *echo.HTTPError
	switch {
	case errors.As(err, &he):
	case errors.As(err, &echoErr):
		he = &exceptions.ResponseError{
			Code:    echoErr.Code,
			Message: http.StatusText(echoErr.Code),
			Details: fmt.Sprint(echoErr.Message),
		}
	default:
		c.Logger().Error(err)
		he = exceptions.ErrInternalServerError
	}

	if responseErr := c.JSON(he.Code, he); responseErr != nil {
		c.Logger().Error(responseErr)
	}
}
