package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/guvanch999/evently-app/internal/config"
	"github.com/guvanch999/evently-app/internal/dto"
	"github.com/guvanch999/evently-app/internal/models"
	"github.com/guvanch999/evently-app/utils/exceptions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

const maxUserInfoBytes = 1 << 20

// UserStore is the persistence contract the services need for users.
// *repositories.UserRepository satisfies it; tests use an in-memory fake.
type UserStore interface {
	GetUserByEmail(email string) (*models.User, error)
	GetUserByGoogleId(googleId string) (*models.User, error)
	GetUserById(id uuid.UUID) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
}

type AuthService struct {
	oAuthConfig    *oauth2.Config
	userInfoUrl    string
	httpClient     *http.Client
	userRepository UserStore
	tokens         *TokenService
}

// AuthOption customises the AuthService (used by tests to point at a fake Google).
type AuthOption func(*AuthService)

// WithOAuthEndpoint overrides Google's authorization and token endpoints.
func WithOAuthEndpoint(endpoint oauth2.Endpoint) AuthOption {
	return func(s *AuthService) { s.oAuthConfig.Endpoint = endpoint }
}

// WithHTTPClient sets the HTTP client used for the token exchange and userinfo calls.
func WithHTTPClient(client *http.Client) AuthOption {
	return func(s *AuthService) { s.httpClient = client }
}

func NewAuthService(cfg *config.Schema, userRepository UserStore, tokens *TokenService, opts ...AuthOption) *AuthService {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GAuthClientId,
		ClientSecret: cfg.GAuthClientSecret,
		RedirectURL:  cfg.GAuthRedirectUrl,
		Scopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	s := &AuthService{
		oAuthConfig:    oauthConfig,
		userInfoUrl:    cfg.GAuthUserInfoUrl,
		httpClient:     http.DefaultClient,
		userRepository: userRepository,
		tokens:         tokens,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// AuthCodeURL builds the Google consent URL bound to the given anti-CSRF state.
func (s *AuthService) AuthCodeURL(state string) string {
	return s.oAuthConfig.AuthCodeURL(state)
}

// SignInWithGoogle exchanges the authorization code, loads the Google profile,
// creates or updates the local user and issues the application's access token.
func (s *AuthService) SignInWithGoogle(ctx context.Context, code string) (*dto.AuthResponseDto, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, s.httpClient)

	token, err := s.oAuthConfig.Exchange(ctx, code)
	if err != nil {
		log.Printf("google oauth: code exchange failed: %v", err)
		return nil, exceptions.NewResponseError(exceptions.ErrBadRequest, errors.New("failed to exchange authorization code"))
	}

	userInfo, err := s.fetchUserInfo(ctx, token)
	if err != nil {
		log.Printf("google oauth: fetching userinfo failed: %v", err)
		return nil, exceptions.NewResponseError(exceptions.ErrBadGateway, errors.New("failed to fetch Google user profile"))
	}
	if !userInfo.VerifiedEmail {
		return nil, exceptions.NewResponseError(exceptions.ErrForbidden, errors.New("google account email is not verified"))
	}

	user, err := s.upsertUser(userInfo)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.tokens.IssueAccessToken(user.Id)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponseDto{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.tokens.TTL().Seconds()),
		User:        dto.UserToResponseDto(user),
	}, nil
}

func (s *AuthService) fetchUserInfo(ctx context.Context, token *oauth2.Token) (*dto.OAuthUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.userInfoUrl, nil)
	if err != nil {
		return nil, err
	}
	// The oauth2 client sends the Google access token in the Authorization header.
	response, err := s.oAuthConfig.Client(ctx, token).Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo returned status %d", response.StatusCode)
	}

	var userInfo dto.OAuthUserInfo
	if err := json.NewDecoder(io.LimitReader(response.Body, maxUserInfoBytes)).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("decode userinfo: %w", err)
	}
	userInfo.Email = strings.ToLower(strings.TrimSpace(userInfo.Email))
	if userInfo.ID == "" || userInfo.Email == "" {
		return nil, errors.New("userinfo is missing id or email")
	}
	return &userInfo, nil
}

// upsertUser finds the user by Google id (falling back to email) and keeps the
// Google identity in sync. Names edited by the user are not overwritten.
func (s *AuthService) upsertUser(info *dto.OAuthUserInfo) (*models.User, error) {
	user, err := s.userRepository.GetUserByGoogleId(info.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user, err = s.userRepository.GetUserByEmail(info.Email)
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if user == nil {
		user = dto.UserFromOAuthInfo(*info)
		if err := s.userRepository.CreateUser(user); err != nil {
			return nil, err
		}
		return user, nil
	}

	user.GoogleId = info.ID
	user.Email = info.Email
	if user.FirstName == "" {
		user.FirstName = info.GivenName
	}
	if user.LastName == "" {
		user.LastName = info.FamilyName
	}
	if err := s.userRepository.UpdateUser(user); err != nil {
		return nil, err
	}
	return user, nil
}
