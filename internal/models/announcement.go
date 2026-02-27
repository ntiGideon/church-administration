package models

import (
	"context"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/announcement"
	"github.com/ntiGideon/ent/church"
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

// ListByChurch returns all announcements for a church, newest first.
// If churchID is 0, returns announcements across all churches (super_admin view).
func (m *AnnouncementModel) ListByChurch(ctx context.Context, churchID int) ([]*ent.Announcement, error) {
	q := m.Db.Announcement.Query()
	if churchID > 0 {
		q = q.Where(announcement.HasChurchWith(church.IDEQ(churchID)))
	} else {
		q = q.Where(announcement.HasChurchWith())
	}
	return q.WithAuthor(func(q *ent.UserQuery) { q.WithContact() }).
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

// Unpublish reverts an announcement to draft status.
func (m *AnnouncementModel) Unpublish(ctx context.Context, id int) error {
	_, err := m.Db.Announcement.UpdateOneID(id).
		SetIsPublished(false).
		ClearPublishedAt().
		Save(ctx)
	return err
}

// Update saves changes to an existing announcement.
func (m *AnnouncementModel) Update(ctx context.Context, id int, dto *AnnouncementDto) error {
	b := m.Db.Announcement.UpdateOneID(id).
		SetTitle(dto.Title).
		SetContent(dto.Content).
		SetCategory(announcement.Category(dto.Category))

	if dto.IsPublished {
		b = b.SetIsPublished(true).SetPublishedAt(time.Now())
	} else {
		b = b.SetIsPublished(false).ClearPublishedAt()
	}

	_, err := b.Save(ctx)
	return err
}

// Delete removes an announcement by ID.
func (m *AnnouncementModel) Delete(ctx context.Context, id int) error {
	return m.Db.Announcement.DeleteOneID(id).Exec(ctx)
}
