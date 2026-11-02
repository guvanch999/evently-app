package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/guvanch999/evently-app/internal/models"
)

const DateLayout = "2006-01-02"

type UserResponseDto struct {
	Id                uuid.UUID `json:"id"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	InstagramNickname string    `json:"instagram_nickname"`
	BirthDate         *string   `json:"birth_date" example:"1999-05-21"`
	Email             string    `json:"email"`
	CreatedAt         time.Time `json:"created_at"`
}

// UpdateProfileDto is a partial update: omitted fields keep their current value.
type UpdateProfileDto struct {
	FirstName         *string `json:"first_name" validate:"omitempty,min=1,max=100"`
	LastName          *string `json:"last_name" validate:"omitempty,max=255"`
	InstagramNickname *string `json:"instagram_nickname" validate:"omitempty,max=100"`
	// BirthDate uses the YYYY-MM-DD format; an empty string clears it.
	BirthDate *string `json:"birth_date" example:"1999-05-21"`
}

func UserToResponseDto(user *models.User) *UserResponseDto {
	var birthDate *string
	if user.BirthDate != nil {
		formatted := user.BirthDate.Format(DateLayout)
		birthDate = &formatted
	}
	return &UserResponseDto{
		Id:                user.Id,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		InstagramNickname: user.InstagramNickname,
		BirthDate:         birthDate,
		Email:             user.Email,
		CreatedAt:         user.CreatedAt,
	}
}
