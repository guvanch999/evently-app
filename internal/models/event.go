package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Event struct {
	Id          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	OwnerId     uuid.UUID  `gorm:"type:uuid;not null;index" json:"owner_id"`
	Owner       *User      `gorm:"foreignKey:OwnerId;constraint:OnDelete:CASCADE" json:"-"`
	Title       string     `gorm:"type:varchar(200);not null" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	Location    string     `gorm:"type:varchar(255)" json:"location"`
	StartsAt    time.Time  `gorm:"type:timestamptz;not null;index" json:"starts_at"`
	EndsAt      *time.Time `gorm:"type:timestamptz" json:"ends_at"`
	CreatedAt   time.Time  `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"type:timestamptz;not null" json:"updated_at"`
}

func (e *Event) TableName() string {
	return "events"
}

func (e *Event) BeforeCreate(_ *gorm.DB) error {
	if e.Id == uuid.Nil {
		e.Id = uuid.New()
	}
	return nil
}
