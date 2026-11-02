package dto

import "github.com/guvanch999/evently-app/internal/models"

// OAuthUserInfo is the payload returned by Google's userinfo endpoint (v2).
type OAuthUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// AuthResponseDto is returned after a successful Google sign-in.
type AuthResponseDto struct {
	AccessToken string           `json:"access_token"`
	TokenType   string           `json:"token_type" example:"Bearer"`
	ExpiresIn   int64            `json:"expires_in" example:"3600"`
	User        *UserResponseDto `json:"user"`
}

func UserFromOAuthInfo(info OAuthUserInfo) *models.User {
	return &models.User{
		FirstName: info.GivenName,
		LastName:  info.FamilyName,
		Email:     info.Email,
		GoogleId:  info.ID,
	}
}
