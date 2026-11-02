package services_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/guvanch999/evently-app/internal/dto"
	"github.com/guvanch999/evently-app/internal/services"
	"github.com/guvanch999/evently-app/internal/testutil"
	"github.com/guvanch999/evently-app/utils/exceptions"
)

func ptr[T any](v T) *T { return &v }

func assertStatus(t *testing.T, err error, want int) {
	t.Helper()
	var re *exceptions.ResponseError
	if !errors.As(err, &re) {
		t.Fatalf("expected a ResponseError with status %d, got %v", want, err)
	}
	if re.Code != want {
		t.Fatalf("expected status %d, got %d (%s)", want, re.Code, re.Details)
	}
}

func newEventService() *services.EventService {
	return services.NewEventService(testutil.NewEventStore())
}

func TestCreateEvent(t *testing.T) {
	svc := newEventService()
	owner := uuid.New()
	start := time.Date(2026, 12, 1, 18, 0, 0, 0, time.FixedZone("UTC+5", 5*3600))

	event, err := svc.CreateEvent(owner, dto.CreateEventDto{
		Title:    "  Go meetup  ",
		StartsAt: start,
		EndsAt:   ptr(start.Add(3 * time.Hour)),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if event.Id == uuid.Nil || event.OwnerId != owner {
		t.Fatalf("unexpected ids: %+v", event)
	}
	if event.Title != "Go meetup" {
		t.Fatalf("title should be trimmed, got %q", event.Title)
	}
	if event.StartsAt.Location() != time.UTC || !event.StartsAt.Equal(start) {
		t.Fatalf("starts_at should be stored in UTC, got %v", event.StartsAt)
	}
}

func TestCreateEventValidation(t *testing.T) {
	svc := newEventService()
	start := time.Now().Add(24 * time.Hour)

	cases := map[string]dto.CreateEventDto{
		"blank title":      {Title: "   ", StartsAt: start},
		"missing start":    {Title: "Party"},
		"end before start": {Title: "Party", StartsAt: start, EndsAt: ptr(start.Add(-time.Hour))},
		"end equals start": {Title: "Party", StartsAt: start, EndsAt: ptr(start)},
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := svc.CreateEvent(uuid.New(), body)
			assertStatus(t, err, http.StatusUnprocessableEntity)
		})
	}
}

func TestOnlyOwnerCanUpdateOrDelete(t *testing.T) {
	svc := newEventService()
	owner, stranger := uuid.New(), uuid.New()
	start := time.Now().Add(48 * time.Hour).UTC()

	event, err := svc.CreateEvent(owner, dto.CreateEventDto{Title: "Concert", StartsAt: start})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.UpdateEvent(stranger, event.Id, dto.UpdateEventDto{Title: ptr("Hijacked")})
	assertStatus(t, err, http.StatusForbidden)
	assertStatus(t, svc.DeleteEvent(stranger, event.Id), http.StatusForbidden)

	updated, err := svc.UpdateEvent(owner, event.Id, dto.UpdateEventDto{
		Location: ptr("Main hall"),
		EndsAt:   ptr(start.Add(2 * time.Hour)),
	})
	if err != nil {
		t.Fatalf("owner update: %v", err)
	}
	if updated.Title != "Concert" || updated.Location != "Main hall" || updated.EndsAt == nil {
		t.Fatalf("partial update not applied correctly: %+v", updated)
	}

	_, err = svc.UpdateEvent(owner, event.Id, dto.UpdateEventDto{StartsAt: ptr(start.Add(5 * time.Hour))})
	assertStatus(t, err, http.StatusUnprocessableEntity) // would end before it starts

	if err := svc.DeleteEvent(owner, event.Id); err != nil {
		t.Fatalf("owner delete: %v", err)
	}
	_, err = svc.GetEvent(event.Id)
	assertStatus(t, err, http.StatusNotFound)
}

func TestMissingEventReturnsNotFound(t *testing.T) {
	svc := newEventService()
	_, err := svc.GetEvent(uuid.New())
	assertStatus(t, err, http.StatusNotFound)
	_, err = svc.UpdateEvent(uuid.New(), uuid.New(), dto.UpdateEventDto{})
	assertStatus(t, err, http.StatusNotFound)
	assertStatus(t, svc.DeleteEvent(uuid.New(), uuid.New()), http.StatusNotFound)
}

func TestListEventsPagination(t *testing.T) {
	svc := newEventService()
	owner := uuid.New()
	base := time.Now().Add(time.Hour).UTC()
	for i := 4; i >= 0; i-- { // insert out of order
		if _, err := svc.CreateEvent(owner, dto.CreateEventDto{Title: "Event", StartsAt: base.Add(time.Duration(i) * time.Hour)}); err != nil {
			t.Fatal(err)
		}
	}

	page, err := svc.ListEvents(dto.PaginationQuery{Page: 2, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if page.Pagination != (dto.PaginationDto{Page: 2, Limit: 2, Total: 5}) {
		t.Fatalf("unexpected pagination: %+v", page.Pagination)
	}
	if len(page.List) != 2 || !page.List[0].StartsAt.Equal(base.Add(2*time.Hour)) {
		t.Fatalf("expected events 3-4 ordered by start, got %+v", page.List)
	}

	defaults, err := svc.ListEvents(dto.PaginationQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if defaults.Pagination.Page != 1 || defaults.Pagination.Limit != dto.DefaultPageLimit || len(defaults.List) != 5 {
		t.Fatalf("unexpected defaults: %+v", defaults.Pagination)
	}

	beyond, err := svc.ListEvents(dto.PaginationQuery{Page: 10, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(beyond.List) != 0 || beyond.List == nil {
		t.Fatalf("a page past the end should be an empty list, got %#v", beyond.List)
	}
}

func TestPaginationNormalize(t *testing.T) {
	got := dto.PaginationQuery{Page: -1, Limit: 1000}.Normalize()
	if got.Page != 1 || got.Limit != dto.MaxPageLimit || got.Offset() != 0 {
		t.Fatalf("unexpected normalization: %+v", got)
	}
	if off := (dto.PaginationQuery{Page: 3, Limit: 10}).Offset(); off != 20 {
		t.Fatalf("expected offset 20, got %d", off)
	}
}
