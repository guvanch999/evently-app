package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	Id                uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	FirstName         string     `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName          string     `gorm:"type:varchar(255);not null" json:"last_name"`
	InstagramNickname string     `gorm:"type:varchar(100)" json:"instagram_nickname"`
	BirthDate         *time.Time `gorm:"type:date" json:"birth_date"`
	GoogleId          string     `gorm:"type:varchar(100);uniqueIndex" json:"google_id"`
	Email             string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"email"`
	CreatedAt         time.Time  `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"type:timestamptz;not null" json:"updated_at"`
}

func (u *User) TableName() string {
	return "users"
}

// BeforeCreate assigns a UUID in Go, so no Postgres extension (uuid-ossp) is needed.
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.Id == uuid.Nil {
		u.Id = uuid.New()
	}
	return nil
}
