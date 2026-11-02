package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/guvanch999/evently-app/internal/middleware"
	"github.com/guvanch999/evently-app/internal/services"
	"github.com/guvanch999/evently-app/utils/errorHandler"
	"github.com/guvanch999/evently-app/utils/util"
	"github.com/labstack/echo/v4"
)

const secret = "middleware-test-secret-of-32-bytes!!"

func newEcho(tokens *services.TokenService) *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = errorHandler.ResponseHTTPErrorHandler
	e.GET("/private", func(c echo.Context) error {
		userId, err := util.GetUserId(c)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, userId.String())
	}, middleware.AuthMiddleware(tokens))
	return e
}

func TestAuthMiddleware(t *testing.T) {
	tokens := services.NewTokenService(secret, "evently-test", time.Hour)
	e := newEcho(tokens)
	userId := uuid.New()
	valid, err := tokens.IssueAccessToken(userId)
	if err != nil {
		t.Fatal(err)
	}
	foreign, _ := services.NewTokenService("a-different-secret-of-32-bytes-long!", "evently-test", time.Hour).IssueAccessToken(userId)

	cases := []struct {
		name   string
		header string
		status int
	}{
		{"no header", "", http.StatusUnauthorized},
		{"wrong scheme", "Basic " + valid, http.StatusUnauthorized},
		{"bearer without token", "Bearer ", http.StatusUnauthorized},
		{"invalid token", "Bearer abc.def.ghi", http.StatusUnauthorized},
		{"token signed with another key", "Bearer " + foreign, http.StatusUnauthorized},
		{"valid token", "Bearer " + valid, http.StatusOK},
		{"case-insensitive scheme", "bearer " + valid, http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/private", nil)
			if tc.header != "" {
				req.Header.Set(echo.HeaderAuthorization, tc.header)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tc.status {
				t.Fatalf("expected %d, got %d: %s", tc.status, rec.Code, rec.Body.String())
			}
			if tc.status == http.StatusOK && rec.Body.String() != userId.String() {
				t.Fatalf("user id not propagated, got %q", rec.Body.String())
			}
		})
	}
}
