package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/guvanch999/evently-app/internal/models"
)

const (
	DefaultPageLimit = 20
	MaxPageLimit     = 100
)

type CreateEventDto struct {
	Title       string     `json:"title" validate:"required,min=1,max=200" example:"Go meetup"`
	Description string     `json:"description" validate:"max=5000" example:"Talks about Go and Postgres"`
	Location    string     `json:"location" validate:"max=255" example:"Online"`
	StartsAt    time.Time  `json:"starts_at" example:"2026-12-01T18:00:00Z"`
	EndsAt      *time.Time `json:"ends_at" example:"2026-12-01T21:00:00Z"`
}

// UpdateEventDto is a partial update: omitted fields keep their current value.
type UpdateEventDto struct {
	Title       *string    `json:"title" validate:"omitempty,min=1,max=200"`
	Description *string    `json:"description" validate:"omitempty,max=5000"`
	Location    *string    `json:"location" validate:"omitempty,max=255"`
	StartsAt    *time.Time `json:"starts_at"`
	EndsAt      *time.Time `json:"ends_at"`
}

type PaginationQuery struct {
	Page  int `query:"page" validate:"omitempty,min=1" example:"1"`
	Limit int `query:"limit" validate:"omitempty,min=1,max=100" example:"20"`
}

// Normalize applies defaults so the repository always receives sane values.
func (p PaginationQuery) Normalize() PaginationQuery {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = DefaultPageLimit
	}
	if p.Limit > MaxPageLimit {
		p.Limit = MaxPageLimit
	}
	return p
}

func (p PaginationQuery) Offset() int {
	return (p.Page - 1) * p.Limit
}

type PaginationDto struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type EventResponseDto struct {
	Id          uuid.UUID  `json:"id"`
	OwnerId     uuid.UUID  `json:"owner_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Location    string     `json:"location"`
	StartsAt    time.Time  `json:"starts_at"`
	EndsAt      *time.Time `json:"ends_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type EventListDto struct {
	List       []*EventResponseDto `json:"list"`
	Pagination PaginationDto       `json:"pagination"`
}

func EventToResponseDto(event *models.Event) *EventResponseDto {
	return &EventResponseDto{
		Id:          event.Id,
		OwnerId:     event.OwnerId,
		Title:       event.Title,
		Description: event.Description,
		Location:    event.Location,
		StartsAt:    event.StartsAt,
		EndsAt:      event.EndsAt,
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   event.UpdatedAt,
	}
}
