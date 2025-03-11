package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/guvanch999/evently-app/utils/exceptions"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Users struct {
	Id        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Username  string
	Name      string
	Password  string
	Email     string
	Active    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u *Users) TableName() string {
	return "users"
}

func (u *Users) BeforeCreate(_ *gorm.DB) (err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10)
	if err != nil {
		return exceptions.NewResponseError(exceptions.ErrBadRequest, err)
	}
	u.Password = string(hash)
	return nil
}
