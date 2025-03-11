package dto

import "github.com/guvanch999/evently-app/internal/models"

type OAuthUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

func UserFromOAuthInfo(info OAuthUserInfo) *models.User {
	return &models.User{
		FirstName: info.GivenName,
		LastName:  info.FamilyName,
		Email:     info.Email,
		GoogleId:  info.ID,
	}
}
