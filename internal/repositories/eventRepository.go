package repositories

import (
	"github.com/google/uuid"
	"github.com/guvanch999/evently-app/internal/models"
	"gorm.io/gorm"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{
		db,
	}
}

func (r *EventRepository) CreateEvent(event *models.Event) error {
	return r.db.Create(event).Error
}

func (r *EventRepository) GetEventById(id uuid.UUID) (*models.Event, error) {
	var event models.Event
	err := r.db.Where("id = ?", id).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// ListEvents returns one page of events ordered by start time, plus the total count.
func (r *EventRepository) ListEvents(offset, limit int) ([]models.Event, int64, error) {
	var total int64
	if err := r.db.Model(&models.Event{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	events := make([]models.Event, 0, limit)
	err := r.db.
		Order("starts_at ASC").
		Order("id ASC").
		Offset(offset).
		Limit(limit).
		Find(&events).Error
	if err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (r *EventRepository) UpdateEvent(event *models.Event) error {
	return r.db.Save(event).Error
}

func (r *EventRepository) DeleteEvent(id uuid.UUID) error {
	return r.db.Delete(&models.Event{}, "id = ?", id).Error
}
