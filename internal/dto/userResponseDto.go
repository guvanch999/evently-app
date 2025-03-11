package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/guvanch999/evently-app/internal/models"
)

type UserResponseDto struct {
	Id                uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	FirstName         string    `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName          string    `gorm:"type:varchar(255);not null" json:"last_name"`
	InstagramNickname string    `gorm:"type:varchar(100)" json:"instagram_nickname"`
	BirthDate         time.Time `gorm:"type:timestamp" json:"birth_date"`
	GoogleId          string    `gorm:"type:varchar(100)" json:"google_id"`
	Email             string    `gorm:"type:varchar(100);not null;unique" json:"email"`
}

func UserToResponseDto(user *models.User) *UserResponseDto {
	return &UserResponseDto{
		Id:                user.Id,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		InstagramNickname: user.InstagramNickname,
		BirthDate:         user.BirthDate,
		GoogleId:          user.GoogleId,
		Email:             user.Email,
	}
}
