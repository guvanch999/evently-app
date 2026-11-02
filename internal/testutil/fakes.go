// Package testutil holds in-memory fakes used by unit and HTTP tests so the
// suite runs without PostgreSQL or any other external service.
package testutil

import (
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/guvanch999/evently-app/internal/models"
	"gorm.io/gorm"
)

// UserStore is an in-memory implementation of services.UserStore.
type UserStore struct {
	mu    sync.Mutex
	users map[uuid.UUID]models.User
}

func NewUserStore() *UserStore {
	return &UserStore{users: map[uuid.UUID]models.User{}}
}

func (s *UserStore) find(match func(models.User) bool) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.users {
		if match(u) {
			user := u
			return &user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *UserStore) GetUserByEmail(email string) (*models.User, error) {
	return s.find(func(u models.User) bool { return u.Email == email })
}

func (s *UserStore) GetUserByGoogleId(googleId string) (*models.User, error) {
	return s.find(func(u models.User) bool { return u.GoogleId == googleId })
}

func (s *UserStore) GetUserById(id uuid.UUID) (*models.User, error) {
	return s.find(func(u models.User) bool { return u.Id == id })
}

func (s *UserStore) CreateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if user.Id == uuid.Nil {
		user.Id = uuid.New()
	}
	now := time.Now().UTC()
	user.CreatedAt, user.UpdatedAt = now, now
	s.users[user.Id] = *user
	return nil
}

func (s *UserStore) UpdateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[user.Id]; !ok {
		return gorm.ErrRecordNotFound
	}
	user.UpdatedAt = time.Now().UTC()
	s.users[user.Id] = *user
	return nil
}

// Count returns the number of stored users.
func (s *UserStore) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.users)
}

// EventStore is an in-memory implementation of services.EventStore.
type EventStore struct {
	mu     sync.Mutex
	events map[uuid.UUID]models.Event
}

func NewEventStore() *EventStore {
	return &EventStore{events: map[uuid.UUID]models.Event{}}
}

func (s *EventStore) CreateEvent(event *models.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if event.Id == uuid.Nil {
		event.Id = uuid.New()
	}
	now := time.Now().UTC()
	event.CreatedAt, event.UpdatedAt = now, now
	s.events[event.Id] = *event
	return nil
}

func (s *EventStore) GetEventById(id uuid.UUID) (*models.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	event, ok := s.events[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &event, nil
}

// ListEvents mirrors the repository: ordered by starts_at, then id.
func (s *EventStore) ListEvents(offset, limit int) ([]models.Event, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	all := make([]models.Event, 0, len(s.events))
	for _, e := range s.events {
		all = append(all, e)
	}
	sort.Slice(all, func(i, j int) bool {
		if !all[i].StartsAt.Equal(all[j].StartsAt) {
			return all[i].StartsAt.Before(all[j].StartsAt)
		}
		return all[i].Id.String() < all[j].Id.String()
	})
	total := int64(len(all))
	if offset >= len(all) {
		return []models.Event{}, total, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], total, nil
}

func (s *EventStore) UpdateEvent(event *models.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.events[event.Id]; !ok {
		return gorm.ErrRecordNotFound
	}
	event.UpdatedAt = time.Now().UTC()
	s.events[event.Id] = *event
	return nil
}

func (s *EventStore) DeleteEvent(id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.events, id)
	return nil
}
