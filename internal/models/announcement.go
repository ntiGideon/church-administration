package models

import (
	"context"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/announcement"
)

type AnnouncementModel struct {
	Db *ent.Client
}

// Create publishes or saves a draft announcement.
func (m *AnnouncementModel) Create(ctx context.Context, dto *AnnouncementDto, churchID, authorID int) (*ent.Announcement, error) {
	b := m.Db.Announcement.Create().
		SetTitle(dto.Title).
		SetContent(dto.Content).
		SetCategory(announcement.Category(dto.Category)).
		SetChurchID(churchID)

	if authorID > 0 {
		b = b.SetAuthorID(authorID)
	}
	if dto.IsPublished {
		b = b.SetIsPublished(true).SetPublishedAt(time.Now())
	}

	a, err := b.Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return a, nil
}

// ListByChurch returns all published announcements for a church, newest first.
func (m *AnnouncementModel) ListByChurch(ctx context.Context, churchID int) ([]*ent.Announcement, error) {
	return m.Db.Announcement.Query().
		Where(announcement.HasChurchWith()).
		WithAuthor(func(q *ent.UserQuery) { q.WithContact() }).
		Order(ent.Desc(announcement.FieldCreatedAt)).
		All(ctx)
}

// GetByID fetches a single announcement.
func (m *AnnouncementModel) GetByID(ctx context.Context, id int) (*ent.Announcement, error) {
	a, err := m.Db.Announcement.Query().
		Where(announcement.IDEQ(id)).
		WithAuthor(func(q *ent.UserQuery) { q.WithContact() }).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return a, nil
}

// Publish marks an announcement as published.
func (m *AnnouncementModel) Publish(ctx context.Context, id int) error {
	_, err := m.Db.Announcement.UpdateOneID(id).
		SetIsPublished(true).
		SetPublishedAt(time.Now()).
		Save(ctx)
	return err
}
