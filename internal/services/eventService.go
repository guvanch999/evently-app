package services

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/guvanch999/evently-app/internal/dto"
	"github.com/guvanch999/evently-app/internal/models"
	"github.com/guvanch999/evently-app/utils/exceptions"
	"github.com/guvanch999/evently-app/utils/util"
	"gorm.io/gorm"
)

// EventStore is the persistence contract for events.
// *repositories.EventRepository satisfies it; tests use an in-memory fake.
type EventStore interface {
	CreateEvent(event *models.Event) error
	GetEventById(id uuid.UUID) (*models.Event, error)
	ListEvents(offset, limit int) ([]models.Event, int64, error)
	UpdateEvent(event *models.Event) error
	DeleteEvent(id uuid.UUID) error
}

type EventService struct {
	eventRepository EventStore
}

func NewEventService(eventRepository EventStore) *EventService {
	return &EventService{eventRepository: eventRepository}
}

func (s *EventService) CreateEvent(ownerId uuid.UUID, body dto.CreateEventDto) (*dto.EventResponseDto, error) {
	event := &models.Event{
		OwnerId:     ownerId,
		Title:       strings.TrimSpace(body.Title),
		Description: strings.TrimSpace(body.Description),
		Location:    strings.TrimSpace(body.Location),
		StartsAt:    body.StartsAt.UTC(),
		EndsAt:      utcPtr(body.EndsAt),
	}
	if err := validateEvent(event); err != nil {
		return nil, err
	}
	if err := s.eventRepository.CreateEvent(event); err != nil {
		return nil, err
	}
	return dto.EventToResponseDto(event), nil
}

func (s *EventService) ListEvents(query dto.PaginationQuery) (*dto.EventListDto, error) {
	query = query.Normalize()
	events, total, err := s.eventRepository.ListEvents(query.Offset(), query.Limit)
	if err != nil {
		return nil, err
	}
	list := util.Map(events, func(e models.Event) *dto.EventResponseDto {
		return dto.EventToResponseDto(&e)
	})
	return &dto.EventListDto{
		List: list,
		Pagination: dto.PaginationDto{
			Page:  query.Page,
			Limit: query.Limit,
			Total: total,
		},
	}, nil
}

func (s *EventService) GetEvent(id uuid.UUID) (*dto.EventResponseDto, error) {
	event, err := s.getEvent(id)
	if err != nil {
		return nil, err
	}
	return dto.EventToResponseDto(event), nil
}

// UpdateEvent applies a partial update; only the event owner may change it.
func (s *EventService) UpdateEvent(actorId, id uuid.UUID, body dto.UpdateEventDto) (*dto.EventResponseDto, error) {
	event, err := s.getOwnedEvent(actorId, id)
	if err != nil {
		return nil, err
	}

	if body.Title != nil {
		event.Title = strings.TrimSpace(*body.Title)
	}
	if body.Description != nil {
		event.Description = strings.TrimSpace(*body.Description)
	}
	if body.Location != nil {
		event.Location = strings.TrimSpace(*body.Location)
	}
	if body.StartsAt != nil {
		event.StartsAt = body.StartsAt.UTC()
	}
	if body.EndsAt != nil {
		event.EndsAt = utcPtr(body.EndsAt)
	}
	if err := validateEvent(event); err != nil {
		return nil, err
	}

	if err := s.eventRepository.UpdateEvent(event); err != nil {
		return nil, err
	}
	return dto.EventToResponseDto(event), nil
}

// DeleteEvent removes an event; only the event owner may delete it.
func (s *EventService) DeleteEvent(actorId, id uuid.UUID) error {
	if _, err := s.getOwnedEvent(actorId, id); err != nil {
		return err
	}
	return s.eventRepository.DeleteEvent(id)
}

func (s *EventService) getEvent(id uuid.UUID) (*models.Event, error) {
	event, err := s.eventRepository.GetEventById(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exceptions.NewResponseError(exceptions.ErrNotFound, errors.New("event not found"))
	}
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (s *EventService) getOwnedEvent(actorId, id uuid.UUID) (*models.Event, error) {
	event, err := s.getEvent(id)
	if err != nil {
		return nil, err
	}
	if event.OwnerId != actorId {
		return nil, exceptions.NewResponseError(exceptions.ErrForbidden, errors.New("only the event owner can modify this event"))
	}
	return event, nil
}

func validateEvent(event *models.Event) error {
	switch {
	case event.Title == "":
		return exceptions.NewResponseError(exceptions.ErrUnprocessableEntity, errors.New("title is required"))
	case event.StartsAt.IsZero():
		return exceptions.NewResponseError(exceptions.ErrUnprocessableEntity, errors.New("starts_at is required"))
	case event.EndsAt != nil && !event.EndsAt.After(event.StartsAt):
		return exceptions.NewResponseError(exceptions.ErrUnprocessableEntity, errors.New("ends_at must be after starts_at"))
	}
	return nil
}

func utcPtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	utc := t.UTC()
	return &utc
}
