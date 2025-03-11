package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/guvanch999/evently-app/internal/config"
	"github.com/guvanch999/evently-app/internal/dto"
	"github.com/guvanch999/evently-app/internal/repositories"
	"github.com/guvanch999/evently-app/utils/exceptions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

type AuthService struct {
	cfg            *config.Schema
	oAuthConfig    *oauth2.Config
	userRepository *repositories.UserRepository
}

func NewAuthService(cfg *config.Schema, userRepository *repositories.UserRepository) *AuthService {
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
	return &AuthService{
		cfg:            cfg,
		oAuthConfig:    oauthConfig,
		userRepository: userRepository,
	}
}

func (s *AuthService) GenerageRedirectUrl() string {
	state := uuid.New().String()
	return s.oAuthConfig.AuthCodeURL(state)
}

func (s *AuthService) OAuthCallback(code string) (*dto.UserResponseDto, error) {
	token, err := s.oAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, exceptions.NewResponseError(exceptions.ErrBadRequest, fmt.Errorf("failed to exchange token: %s", err.Error()))
	}

	response, err := http.Get(s.cfg.GAuthUserInfoUrl + token.AccessToken)
	if err != nil {
		return nil, exceptions.NewResponseError(exceptions.ErrBadRequest, fmt.Errorf("failed getting user info: %s", err.Error()))
	}

	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, exceptions.NewResponseError(exceptions.ErrBadRequest, fmt.Errorf("failed read body: %s", err.Error()))
	}

	var userInfo dto.OAuthUserInfo
	err = json.Unmarshal(contents, &userInfo)
	if err != nil {
		return nil, exceptions.NewResponseError(exceptions.ErrBadRequest, fmt.Errorf("failed getting user info: %s", err.Error()))
	}

	user, err := s.userRepository.GetUserByEmail(userInfo.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if user != nil {
		user.FirstName = userInfo.GivenName
		user.LastName = userInfo.FamilyName
		user.GoogleId = userInfo.ID
		err = s.userRepository.UpdateUserProfile(user)
		if err != nil {
			return nil, err
		}
		return dto.UserToResponseDto(user), nil
	}
	user = dto.UserFromOAuthInfo(userInfo)
	err = s.userRepository.CreateUser(user)
	if err != nil {
		return nil, err
	}
	return dto.UserToResponseDto(user), nil
}
