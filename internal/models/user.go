package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id                uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	FirstName         string    `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName          string    `gorm:"type:varchar(255);not null" json:"last_name"`
	InstagramNickname string    `gorm:"type:varchar(100)" json:"instagram_nickname"`
	BirthDate         time.Time `gorm:"type:timestamp" json:"birth_date"`
	GoogleId          string    `gorm:"type:varchar(100)" json:"google_id"`
	Email             string    `gorm:"type:varchar(100);not null;unique" json:"email"`
	CreatedAt         time.Time `gorm:"type:timestamp;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt         time.Time `gorm:"type:timestamp;not null;default:current_timestamp" json:"updated_at"`
}

func (u *User) TableName() string {
	return "users"
}
