package handlers

import (
	"fmt"
	"net/http"

	"github.com/guvanch999/evently-app/internal/services"
	"github.com/guvanch999/evently-app/utils/exceptions"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{
		service,
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
	url := h.service.GenerageRedirectUrl()
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

	contens, err := h.service.OAuthCallback(code)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, contens)
}
