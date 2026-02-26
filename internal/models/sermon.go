package models

import (
	"context"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/sermon"
)

type SermonModel struct {
	Db *ent.Client
}

// Create adds a new sermon record for a church.
func (m *SermonModel) Create(ctx context.Context, dto *SermonDto, churchID int) (*ent.Sermon, error) {
	svcDate, err := time.Parse("2006-01-02", dto.ServiceDate)
	if err != nil {
		svcDate = time.Now()
	}

	s, err := m.Db.Sermon.Create().
		SetTitle(dto.Title).
		SetSpeaker(dto.Speaker).
		SetNillableSeries(nullStr(dto.Series)).
		SetNillableScripture(nullStr(dto.Scripture)).
		SetNillableDescription(nullStr(dto.Description)).
		SetNillableMediaURL(nullStr(dto.MediaURL)).
		SetServiceDate(svcDate).
		SetIsPublished(dto.IsPublished).
		SetChurchID(churchID).
		Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return s, nil
}

// GetByID returns a single sermon.
func (m *SermonModel) GetByID(ctx context.Context, id int) (*ent.Sermon, error) {
	s, err := m.Db.Sermon.Query().
		Where(sermon.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return s, nil
}

// ListByChurch returns all sermons for a church, newest first.
func (m *SermonModel) ListByChurch(ctx context.Context, churchID int) ([]*ent.Sermon, error) {
	q := m.Db.Sermon.Query()
	if churchID > 0 {
		q = q.Where(sermon.HasChurchWith(church.IDEQ(churchID)))
	}
	return q.Order(ent.Desc(sermon.FieldServiceDate)).All(ctx)
}

// Update saves editable fields on a sermon.
func (m *SermonModel) Update(ctx context.Context, id int, dto *SermonDto) error {
	svcDate, err := time.Parse("2006-01-02", dto.ServiceDate)
	if err != nil {
		svcDate = time.Now()
	}
	_, err = m.Db.Sermon.UpdateOneID(id).
		SetTitle(dto.Title).
		SetSpeaker(dto.Speaker).
		SetNillableSeries(nullStr(dto.Series)).
		SetNillableScripture(nullStr(dto.Scripture)).
		SetNillableDescription(nullStr(dto.Description)).
		SetNillableMediaURL(nullStr(dto.MediaURL)).
		SetServiceDate(svcDate).
		SetIsPublished(dto.IsPublished).
		Save(ctx)
	return err
}

// TogglePublish flips the published state of a sermon.
func (m *SermonModel) TogglePublish(ctx context.Context, id int, published bool) error {
	_, err := m.Db.Sermon.UpdateOneID(id).SetIsPublished(published).Save(ctx)
	return err
}

// Delete removes a sermon.
func (m *SermonModel) Delete(ctx context.Context, id int) error {
	return m.Db.Sermon.DeleteOneID(id).Exec(ctx)
}
