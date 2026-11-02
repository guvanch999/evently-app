package internal_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/guvanch999/evently-app/internal"
	"github.com/guvanch999/evently-app/internal/config"
	"github.com/guvanch999/evently-app/internal/dto"
	"github.com/guvanch999/evently-app/internal/handlers"
	"github.com/guvanch999/evently-app/internal/middleware"
	"github.com/guvanch999/evently-app/internal/models"
	"github.com/guvanch999/evently-app/internal/services"
	"github.com/guvanch999/evently-app/internal/testutil"
	"github.com/guvanch999/evently-app/utils/errorHandler"
	"github.com/guvanch999/evently-app/utils/validator"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

const (
	testClientId     = "test-client-id"
	testClientSecret = "test-client-secret"
	testRedirectUrl  = "http://localhost:8080/v1/auth/google/callback"
	validCode        = "valid-code"
	googleToken      = "google-access-token"
)

// fakeGoogle emulates Google's token and userinfo endpoints.
type fakeGoogle struct {
	server     *httptest.Server
	profile    atomic.Value // dto.OAuthUserInfo
	tokenCalls atomic.Int32
}

func newFakeGoogle(t *testing.T) *fakeGoogle {
	g := &fakeGoogle{}
	g.profile.Store(dto.OAuthUserInfo{
		ID: "google-123", Email: "Ada@Example.com", VerifiedEmail: true, GivenName: "Ada", FamilyName: "Lovelace",
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		g.tokenCalls.Add(1)
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost ||
			r.PostForm.Get("grant_type") != "authorization_code" ||
			r.PostForm.Get("code") != validCode ||
			r.PostForm.Get("redirect_uri") != testRedirectUrl ||
			r.PostForm.Get("client_id") != testClientId ||
			r.PostForm.Get("client_secret") != testClientSecret {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"error":"invalid_grant"}`)
			return
		}
		_, _ = io.WriteString(w, `{"access_token":"`+googleToken+`","token_type":"Bearer","expires_in":3600}`)
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+googleToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(g.profile.Load().(dto.OAuthUserInfo))
	})

	g.server = httptest.NewServer(mux)
	t.Cleanup(g.server.Close)
	return g
}

type testApp struct {
	echo   *echo.Echo
	google *fakeGoogle
	users  *testutil.UserStore
	tokens *services.TokenService
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()
	google := newFakeGoogle(t)
	cfg := &config.Schema{
		GAuthClientId:     testClientId,
		GAuthClientSecret: testClientSecret,
		GAuthRedirectUrl:  testRedirectUrl,
		GAuthUserInfoUrl:  google.server.URL + "/userinfo",
	}

	users := testutil.NewUserStore()
	tokens := services.NewTokenService("router-test-secret-of-at-least-32-bytes", "evently-test", time.Hour)
	authService := services.NewAuthService(cfg, users, tokens,
		services.WithOAuthEndpoint(oauth2.Endpoint{
			AuthURL:   google.server.URL + "/auth",
			TokenURL:  google.server.URL + "/token",
			AuthStyle: oauth2.AuthStyleInParams,
		}),
		services.WithHTTPClient(google.server.Client()),
	)

	e := echo.New()
	e.Validator = validator.NewValidator()
	e.HTTPErrorHandler = errorHandler.ResponseHTTPErrorHandler
	internal.RegisterRoutes(e, internal.Handlers{
		Auth:  handlers.NewAuthHandler(authService, false),
		User:  handlers.NewUserHandler(services.NewUserService(users)),
		Event: handlers.NewEventHandler(services.NewEventService(testutil.NewEventStore())),
	}, middleware.AuthMiddleware(tokens))

	return &testApp{echo: e, google: google, users: users, tokens: tokens}
}

type request struct {
	method, path, body, token string
	cookies                   []*http.Cookie
}

func (a *testApp) do(r request) *httptest.ResponseRecorder {
	var body io.Reader
	if r.body != "" {
		body = strings.NewReader(r.body)
	}
	req := httptest.NewRequest(r.method, r.path, body)
	if r.body != "" {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	if r.token != "" {
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+r.token)
	}
	for _, c := range r.cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	a.echo.ServeHTTP(rec, req)
	return rec
}

// startLogin hits the login endpoint and returns the state cookie.
func (a *testApp) startLogin(t *testing.T) *http.Cookie {
	t.Helper()
	rec := a.do(request{method: http.MethodGet, path: "/v1/auth/google/login"})
	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("login: expected 307, got %d", rec.Code)
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == handlers.OAuthStateCookieName {
			return c
		}
	}
	t.Fatal("login did not set the state cookie")
	return nil
}

// signIn runs the full Google flow and returns the app's access token.
func (a *testApp) signIn(t *testing.T, profile dto.OAuthUserInfo) (string, *dto.UserResponseDto) {
	t.Helper()
	a.google.profile.Store(profile)
	state := a.startLogin(t)
	rec := a.do(request{
		method:  http.MethodGet,
		path:    "/v1/auth/google/callback?code=" + validCode + "&state=" + url.QueryEscape(state.Value),
		cookies: []*http.Cookie{state},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("callback: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data dto.AuthResponseDto `json:"data"`
	}
	decode(t, rec, &resp)
	return resp.Data.AccessToken, resp.Data.User
}

func decode(t *testing.T, rec *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
}

func expectStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("expected %d, got %d: %s", want, rec.Code, rec.Body.String())
	}
}

func TestGoogleLoginRedirectsWithState(t *testing.T) {
	app := newTestApp(t)
	rec := app.do(request{method: http.MethodGet, path: "/v1/auth/google/login"})
	expectStatus(t, rec, http.StatusTemporaryRedirect)

	location, err := url.Parse(rec.Header().Get(echo.HeaderLocation))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(location.String(), app.google.server.URL+"/auth?") {
		t.Fatalf("unexpected redirect target %s", location)
	}
	q := location.Query()
	if q.Get("client_id") != testClientId || q.Get("redirect_uri") != testRedirectUrl || q.Get("response_type") != "code" {
		t.Fatalf("unexpected auth URL params: %v", q)
	}
	if !strings.Contains(q.Get("scope"), "userinfo.email") {
		t.Fatalf("email scope missing: %q", q.Get("scope"))
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Value == "" || cookies[0].Value != q.Get("state") {
		t.Fatalf("state cookie must match the state sent to Google: %+v vs %q", cookies, q.Get("state"))
	}
}

func TestGoogleCallbackSignsInAndCreatesUser(t *testing.T) {
	app := newTestApp(t)
	profile := dto.OAuthUserInfo{ID: "google-1", Email: "Ada@Example.com", VerifiedEmail: true, GivenName: "Ada", FamilyName: "Lovelace"}

	token, user := app.signIn(t, profile)
	if user.Email != "ada@example.com" || user.FirstName != "Ada" {
		t.Fatalf("unexpected user: %+v", user)
	}
	userId, err := app.tokens.VerifyAccessToken(token)
	if err != nil || userId != user.Id {
		t.Fatalf("issued token should identify the user: id=%s err=%v", userId, err)
	}

	rec := app.do(request{method: http.MethodGet, path: "/v1/me", token: token})
	expectStatus(t, rec, http.StatusOK)
	var me struct {
		Data dto.UserResponseDto `json:"data"`
	}
	decode(t, rec, &me)
	if me.Data.Id != user.Id {
		t.Fatalf("/me returned another user: %+v", me.Data)
	}

	// Signing in again with the same Google account reuses the user.
	_, again := app.signIn(t, profile)
	if again.Id != user.Id || app.users.Count() != 1 {
		t.Fatalf("expected the same user, got %s (count %d)", again.Id, app.users.Count())
	}
}

func TestGoogleCallbackLinksExistingUserByEmail(t *testing.T) {
	app := newTestApp(t)
	existing := &models.User{FirstName: "Countess", LastName: "Lovelace", Email: "ada@example.com"}
	if err := app.users.CreateUser(existing); err != nil {
		t.Fatal(err)
	}

	_, user := app.signIn(t, dto.OAuthUserInfo{ID: "google-9", Email: "ada@example.com", VerifiedEmail: true, GivenName: "Ada"})
	if user.Id != existing.Id || user.FirstName != "Countess" {
		t.Fatalf("existing user should be linked without overwriting names: %+v", user)
	}
	stored, err := app.users.GetUserByGoogleId("google-9")
	if err != nil || stored.Id != existing.Id {
		t.Fatalf("google id should be linked: %v", err)
	}
}

func TestGoogleCallbackRejectsInvalidState(t *testing.T) {
	app := newTestApp(t)
	state := app.startLogin(t)

	cases := []request{
		{method: http.MethodGet, path: "/v1/auth/google/callback?code=" + validCode + "&state=" + state.Value},
		{method: http.MethodGet, path: "/v1/auth/google/callback?code=" + validCode + "&state=forged", cookies: []*http.Cookie{state}},
		{method: http.MethodGet, path: "/v1/auth/google/callback?code=" + validCode, cookies: []*http.Cookie{state}},
	}
	for _, r := range cases {
		expectStatus(t, app.do(r), http.StatusBadRequest)
	}
	if calls := app.google.tokenCalls.Load(); calls != 0 {
		t.Fatalf("the code must not be exchanged when state is invalid, got %d token calls", calls)
	}
}

func TestGoogleCallbackFailures(t *testing.T) {
	t.Run("provider error", func(t *testing.T) {
		app := newTestApp(t)
		state := app.startLogin(t)
		rec := app.do(request{method: http.MethodGet, path: "/v1/auth/google/callback?error=access_denied&state=" + state.Value, cookies: []*http.Cookie{state}})
		expectStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("invalid code", func(t *testing.T) {
		app := newTestApp(t)
		state := app.startLogin(t)
		rec := app.do(request{method: http.MethodGet, path: "/v1/auth/google/callback?code=wrong&state=" + state.Value, cookies: []*http.Cookie{state}})
		expectStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("unverified email", func(t *testing.T) {
		app := newTestApp(t)
		app.google.profile.Store(dto.OAuthUserInfo{ID: "google-2", Email: "eve@example.com", VerifiedEmail: false})
		state := app.startLogin(t)
		rec := app.do(request{method: http.MethodGet, path: "/v1/auth/google/callback?code=" + validCode + "&state=" + state.Value, cookies: []*http.Cookie{state}})
		expectStatus(t, rec, http.StatusForbidden)
		if app.users.Count() != 0 {
			t.Fatal("no user should be created for an unverified email")
		}
	})
}

func TestUpdateMe(t *testing.T) {
	app := newTestApp(t)
	token, _ := app.signIn(t, dto.OAuthUserInfo{ID: "g", Email: "a@example.com", VerifiedEmail: true, GivenName: "A"})

	rec := app.do(request{method: http.MethodPatch, path: "/v1/me", token: token, body: `{"instagram_nickname":"@ada","birth_date":"1990-12-10"}`})
	expectStatus(t, rec, http.StatusOK)
	var resp struct {
		Data dto.UserResponseDto `json:"data"`
	}
	decode(t, rec, &resp)
	if resp.Data.InstagramNickname != "ada" || resp.Data.BirthDate == nil || *resp.Data.BirthDate != "1990-12-10" {
		t.Fatalf("profile not updated: %+v", resp.Data)
	}

	expectStatus(t, app.do(request{method: http.MethodPatch, path: "/v1/me", token: token, body: `{"birth_date":"10/12/1990"}`}), http.StatusUnprocessableEntity)
	expectStatus(t, app.do(request{method: http.MethodPatch, path: "/v1/me", token: token, body: `{"first_name":""}`}), http.StatusBadRequest)
	expectStatus(t, app.do(request{method: http.MethodGet, path: "/v1/me"}), http.StatusUnauthorized)
}

func TestEventsAPI(t *testing.T) {
	app := newTestApp(t)
	ownerToken, owner := app.signIn(t, dto.OAuthUserInfo{ID: "g-owner", Email: "owner@example.com", VerifiedEmail: true, GivenName: "O"})
	otherToken, _ := app.signIn(t, dto.OAuthUserInfo{ID: "g-other", Email: "other@example.com", VerifiedEmail: true, GivenName: "X"})

	body := `{"title":"Go meetup","location":"Online","starts_at":"2026-12-01T18:00:00Z","ends_at":"2026-12-01T20:00:00Z"}`
	expectStatus(t, app.do(request{method: http.MethodPost, path: "/v1/events", body: body}), http.StatusUnauthorized)

	rec := app.do(request{method: http.MethodPost, path: "/v1/events", body: body, token: ownerToken})
	expectStatus(t, rec, http.StatusCreated)
	var created struct {
		Data dto.EventResponseDto `json:"data"`
	}
	decode(t, rec, &created)
	if created.Data.OwnerId != owner.Id || created.Data.Title != "Go meetup" {
		t.Fatalf("unexpected event: %+v", created.Data)
	}
	eventPath := "/v1/events/" + created.Data.Id.String()

	rec = app.do(request{method: http.MethodGet, path: "/v1/events?page=1&limit=10"})
	expectStatus(t, rec, http.StatusOK)
	var list struct {
		Data dto.EventListDto `json:"data"`
	}
	decode(t, rec, &list)
	if list.Data.Pagination.Total != 1 || len(list.Data.List) != 1 {
		t.Fatalf("unexpected list: %+v", list.Data)
	}

	expectStatus(t, app.do(request{method: http.MethodGet, path: eventPath}), http.StatusOK)
	expectStatus(t, app.do(request{method: http.MethodGet, path: "/v1/events/not-a-uuid"}), http.StatusBadRequest)
	expectStatus(t, app.do(request{method: http.MethodGet, path: "/v1/events/" + uuid.NewString()}), http.StatusNotFound)
	expectStatus(t, app.do(request{method: http.MethodGet, path: "/v1/events?limit=1000"}), http.StatusBadRequest)
	expectStatus(t, app.do(request{method: http.MethodPost, path: "/v1/events", body: `{"title":`, token: ownerToken}), http.StatusBadRequest)
	expectStatus(t, app.do(request{method: http.MethodPost, path: "/v1/events", body: `{"title":"No start"}`, token: ownerToken}), http.StatusUnprocessableEntity)

	expectStatus(t, app.do(request{method: http.MethodPatch, path: eventPath, body: `{"title":"Mine now"}`, token: otherToken}), http.StatusForbidden)
	expectStatus(t, app.do(request{method: http.MethodDelete, path: eventPath, token: otherToken}), http.StatusForbidden)

	rec = app.do(request{method: http.MethodPatch, path: eventPath, body: `{"title":"Go meetup #2"}`, token: ownerToken})
	expectStatus(t, rec, http.StatusOK)
	decode(t, rec, &created)
	if created.Data.Title != "Go meetup #2" || created.Data.Location != "Online" {
		t.Fatalf("partial update failed: %+v", created.Data)
	}

	expectStatus(t, app.do(request{method: http.MethodDelete, path: eventPath, token: ownerToken}), http.StatusNoContent)
	expectStatus(t, app.do(request{method: http.MethodGet, path: eventPath}), http.StatusNotFound)
}

func TestUnknownRouteReturns404(t *testing.T) {
	app := newTestApp(t)
	rec := app.do(request{method: http.MethodGet, path: "/v1/nope"})
	expectStatus(t, rec, http.StatusNotFound)
	if !strings.Contains(rec.Body.String(), `"code":404`) {
		t.Fatalf("expected the JSON error envelope, got %s", rec.Body.String())
	}
}
