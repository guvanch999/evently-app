package handlers

import (
	"fmt"
	"net/http"

	"github.com/guvanch999/evently-app/internal/dto"
	"github.com/guvanch999/evently-app/internal/services"
	"github.com/guvanch999/evently-app/utils/exceptions"
	"github.com/guvanch999/evently-app/utils/util"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	service       *services.AuthService
	secureCookies bool
}

// NewAuthHandler creates the Google sign-in handler. secureCookies should be
// true whenever the callback is served over HTTPS.
func NewAuthHandler(service *services.AuthService, secureCookies bool) *AuthHandler {
	return &AuthHandler{
		service:       service,
		secureCookies: secureCookies,
	}
}

// @Summary Sign in with Google
// @Description Starts the Google OAuth2 flow: sets a short-lived HttpOnly state cookie and redirects to Google's consent screen.
// @Tags Auth
// @Success 307 "Redirect to Google"
// @Router /auth/google/login [get]
func (h *AuthHandler) HandleGoogleLogin(c echo.Context) error {
	state, err := setOAuthStateCookie(c, h.secureCookies)
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusTemporaryRedirect, h.service.AuthCodeURL(state))
}

// @Summary Google OAuth callback
// @Description Verifies the state cookie, exchanges the code with Google, upserts the user and returns an access token.
// @Tags Auth
// @Produce json
// @Param code query string true "Authorization code from Google"
// @Param state query string true "State value that must match the state cookie"
// @Success 200 {object} util.DataWrapper[dto.AuthResponseDto]
// @Failure 400 {object} exceptions.ResponseError
// @Failure 403 {object} exceptions.ResponseError "Google email is not verified"
// @Failure 502 {object} exceptions.ResponseError
// @Router /auth/google/callback [get]
func (h *AuthHandler) HandleGoogleCallback(c echo.Context) error {
	if err := verifyOAuthState(c, h.secureCookies); err != nil {
		return exceptions.NewResponseError(exceptions.ErrBadRequest, err)
	}

	if googleErr := c.QueryParam("error"); googleErr != "" {
		return exceptions.NewResponseError(exceptions.ErrBadRequest, fmt.Errorf("google sign-in was not completed: %s", googleErr))
	}

	code := c.QueryParam("code")
	if code == "" {
		return exceptions.NewResponseError(exceptions.ErrBadRequest, fmt.Errorf("code not provided"))
	}

	result, err := h.service.SignInWithGoogle(c.Request().Context(), code)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, util.DataWrapper[*dto.AuthResponseDto]{Data: result})
}
