package models

import (
	"context"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/event"
)

type EventModel struct {
	Db *ent.Client
}

// Create adds a new event for a church.
func (m *EventModel) Create(ctx context.Context, dto *EventDto, churchID int) (*ent.Event, error) {
	start, err := time.Parse("2006-01-02T15:04", dto.StartTime)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02T15:04", dto.EndTime)
	if err != nil {
		return nil, err
	}

	e, err := m.Db.Event.Create().
		SetTitle(dto.Title).
		SetDescription(dto.Description).
		SetStartTime(start).
		SetEndTime(end).
		SetLocation(dto.Location).
		SetEventType(event.EventType(dto.EventType)).
		SetChurchID(churchID).
		Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return e, nil
}

// List returns all events for a church, newest first.
// If churchID is 0, returns events across all churches (super_admin view).
func (m *EventModel) List(ctx context.Context, churchID int) ([]*ent.Event, error) {
	q := m.Db.Event.Query()
	if churchID > 0 {
		q = q.Where(event.HasChurchWith(church.IDEQ(churchID)))
	} else {
		q = q.Where(event.HasChurchWith())
	}
	return q.Order(ent.Desc(event.FieldStartTime)).All(ctx)
}

// Upcoming returns events that start in the future.
// If churchID is 0, returns upcoming events across all churches (super_admin view).
func (m *EventModel) Upcoming(ctx context.Context, churchID int, limit int) ([]*ent.Event, error) {
	q := m.Db.Event.Query().Where(event.StartTimeGTE(time.Now()))
	if churchID > 0 {
		q = q.Where(event.HasChurchWith(church.IDEQ(churchID)))
	}
	return q.Order(ent.Asc(event.FieldStartTime)).Limit(limit).All(ctx)
}

// GetByID returns a single event.
func (m *EventModel) GetByID(ctx context.Context, id int) (*ent.Event, error) {
	e, err := m.Db.Event.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	return e, nil
}

// UpdateAttendance updates the attendance count for an event.
func (m *EventModel) UpdateAttendance(ctx context.Context, id, count int) error {
	_, err := m.Db.Event.UpdateOneID(id).SetAttendanceCount(count).Save(ctx)
	return err
}

// SetPublished toggles the published state of an event.
func (m *EventModel) SetPublished(ctx context.Context, id int, published bool) error {
	_, err := m.Db.Event.UpdateOneID(id).SetIsPublished(published).Save(ctx)
	return err
}
