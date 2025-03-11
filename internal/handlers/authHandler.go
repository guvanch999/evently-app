package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/guvanch999/evently-app/internal/config"
	"github.com/guvanch999/evently-app/utils/exceptions"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthHandler struct {
	cfg         *config.Schema
	oAuthConfig *oauth2.Config
}

func NewAuthHandler(cfg *config.Schema) *AuthHandler {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GAuthClientId,
		ClientSecret: cfg.GAuthClientSecret,
		RedirectURL:  cfg.GAuthRedirectUrl,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	return &AuthHandler{
		cfg,
		oauthConfig,
	}
}

// @Summary Sign in with Google
// @Description Initiates Google OAuth2 login flow
// @Tags Auth
// @Accept json
// @Produce json
// @Success 302
// @Router /auth/google/login [get]
func (h *AuthHandler) HandleGoogleLogin(c echo.Context) error {
	state := uuid.New().String()
	url := h.oAuthConfig.AuthCodeURL(state)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// @Summary Google OAuth callback
// @Description Handles the Google OAuth2 callback
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /auth/google/callback [get]
func (h *AuthHandler) HandleGoogleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return exceptions.NewResponseError(exceptions.ErrBadRequest, fmt.Errorf("code no provided"))
	}

	token, err := h.oAuthConfig.Exchange(c.Request().Context(), code)
	if err != nil {
		return exceptions.NewResponseError(exceptions.ErrBadRequest, fmt.Errorf("failed to exchange token: %s", err.Error()))
	}

	response, err := http.Get(h.cfg.GAuthUserInfoUrl + token.AccessToken)
	if err != nil {
		return exceptions.NewResponseError(exceptions.ErrBadRequest, fmt.Errorf("failed getting user info: %s", err.Error()))
	}

	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return exceptions.NewResponseError(exceptions.ErrBadRequest, fmt.Errorf("failed read body: %s", err.Error()))
	}

	return c.JSON(http.StatusOK, string(contents))
}
