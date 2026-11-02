package handlers

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"time"

	"github.com/guvanch999/evently-app/utils/util"
	"github.com/labstack/echo/v4"
)

const (
	OAuthStateCookieName = "evently_oauth_state"
	oauthStateCookiePath = "/v1/auth/google"
	oauthStateTTL        = 10 * time.Minute
	oauthStateBytes      = 32
)

var errInvalidOAuthState = errors.New("invalid or missing OAuth state")

// setOAuthStateCookie creates a random state value and stores it in a
// short-lived HttpOnly cookie scoped to the Google auth routes. SameSite=Lax is
// required because Google redirects back with a top-level cross-site GET.
func setOAuthStateCookie(c echo.Context, secure bool) (string, error) {
	state, err := util.SecureToken(oauthStateBytes)
	if err != nil {
		return "", err
	}
	c.SetCookie(&http.Cookie{
		Name:     OAuthStateCookieName,
		Value:    state,
		Path:     oauthStateCookiePath,
		MaxAge:   int(oauthStateTTL.Seconds()),
		Expires:  time.Now().Add(oauthStateTTL),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	return state, nil
}

// verifyOAuthState checks the state query parameter against the cookie in
// constant time. The cookie is cleared whatever the outcome, so a state value
// can only be used once.
func verifyOAuthState(c echo.Context, secure bool) error {
	c.SetCookie(&http.Cookie{
		Name:     OAuthStateCookieName,
		Value:    "",
		Path:     oauthStateCookiePath,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})

	cookie, err := c.Cookie(OAuthStateCookieName)
	if err != nil || cookie.Value == "" {
		return errInvalidOAuthState
	}
	state := c.QueryParam("state")
	if state == "" || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(state)) != 1 {
		return errInvalidOAuthState
	}
	return nil
}
