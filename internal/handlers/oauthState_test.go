package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestSetOAuthStateCookie(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/v1/auth/google/login", nil), rec)

	state, err := setOAuthStateCookie(c, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(state) < 40 {
		t.Fatalf("state should carry 32 random bytes, got %q", state)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one cookie, got %d", len(cookies))
	}
	cookie := cookies[0]
	switch {
	case cookie.Name != OAuthStateCookieName || cookie.Value != state:
		t.Fatalf("cookie does not carry the state: %+v", cookie)
	case !cookie.HttpOnly || !cookie.Secure:
		t.Fatalf("cookie must be HttpOnly and Secure: %+v", cookie)
	case cookie.SameSite != http.SameSiteLaxMode:
		t.Fatalf("cookie must be SameSite=Lax, got %v", cookie.SameSite)
	case cookie.Path != "/v1/auth/google" || cookie.MaxAge != 600:
		t.Fatalf("unexpected path/max-age: %+v", cookie)
	}

	other, _ := setOAuthStateCookie(e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder()), true)
	if other == state {
		t.Fatal("state values must be random")
	}
}

func TestVerifyOAuthState(t *testing.T) {
	cases := []struct {
		name   string
		cookie string
		query  string
		ok     bool
	}{
		{"matching", "abc123", "abc123", true},
		{"missing cookie", "", "abc123", false},
		{"missing query", "abc123", "", false},
		{"mismatch", "abc123", "abc124", false},
		{"prefix only", "abc123", "abc", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/auth/google/callback?state="+tc.query, nil)
			if tc.cookie != "" {
				req.AddCookie(&http.Cookie{Name: OAuthStateCookieName, Value: tc.cookie})
			}
			rec := httptest.NewRecorder()
			err := verifyOAuthState(echo.New().NewContext(req, rec), false)
			if (err == nil) != tc.ok {
				t.Fatalf("ok=%v, err=%v", tc.ok, err)
			}

			// The state cookie is always cleared, so it cannot be replayed.
			cleared := false
			for _, c := range rec.Result().Cookies() {
				if c.Name == OAuthStateCookieName && c.MaxAge < 0 {
					cleared = true
				}
			}
			if !cleared {
				t.Fatal("state cookie should be cleared")
			}
		})
	}
}
