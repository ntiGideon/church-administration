package models

import (
	"context"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/prayerrequest"
)

type PrayerRequestModel struct {
	Db *ent.Client
}

// Create adds a new prayer request for a church.
func (m *PrayerRequestModel) Create(ctx context.Context, dto *PrayerRequestDto, churchID int) (*ent.PrayerRequest, error) {
	b := m.Db.PrayerRequest.Create().
		SetTitle(dto.Title).
		SetBody(dto.Body).
		SetIsAnonymous(dto.IsAnonymous).
		SetIsPrivate(dto.IsPrivate).
		SetChurchID(churchID)

	if !dto.IsAnonymous && dto.RequesterName != "" {
		b = b.SetRequesterName(dto.RequesterName)
	}
	if dto.ContactID > 0 {
		b = b.SetContactID(dto.ContactID)
	}

	pr, err := b.Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return pr, nil
}

// GetByID returns a single prayer request with its contact edge loaded.
func (m *PrayerRequestModel) GetByID(ctx context.Context, id int) (*ent.PrayerRequest, error) {
	pr, err := m.Db.PrayerRequest.Query().
		Where(prayerrequest.IDEQ(id)).
		WithContact().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return pr, nil
}

// ListByChurch returns all prayer requests for a church, newest first.
func (m *PrayerRequestModel) ListByChurch(ctx context.Context, churchID int) ([]*ent.PrayerRequest, error) {
	q := m.Db.PrayerRequest.Query().WithContact()
	if churchID > 0 {
		q = q.Where(prayerrequest.HasChurchWith(church.IDEQ(churchID)))
	}
	return q.Order(ent.Desc(prayerrequest.FieldCreatedAt)).All(ctx)
}

// Update saves editable fields on a prayer request.
func (m *PrayerRequestModel) Update(ctx context.Context, id int, dto *PrayerRequestDto) error {
	u := m.Db.PrayerRequest.UpdateOneID(id).
		SetTitle(dto.Title).
		SetBody(dto.Body).
		SetIsAnonymous(dto.IsAnonymous).
		SetIsPrivate(dto.IsPrivate)

	if !dto.IsAnonymous && dto.RequesterName != "" {
		u = u.SetRequesterName(dto.RequesterName)
	} else {
		u = u.ClearRequesterName()
	}

	_, err := u.Save(ctx)
	return err
}

// UpdateStatus changes only the status of a prayer request.
func (m *PrayerRequestModel) UpdateStatus(ctx context.Context, id int, status string) error {
	_, err := m.Db.PrayerRequest.UpdateOneID(id).
		SetStatus(prayerrequest.Status(status)).
		Save(ctx)
	return err
}

// CountByStatus returns a map of status → count for a church.
func (m *PrayerRequestModel) CountByStatus(ctx context.Context, churchID int) (map[string]int, error) {
	all, err := m.ListByChurch(ctx, churchID)
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	for _, pr := range all {
		counts[string(pr.Status)]++
	}
	return counts, nil
}

// Delete removes a prayer request.
func (m *PrayerRequestModel) Delete(ctx context.Context, id int) error {
	return m.Db.PrayerRequest.DeleteOneID(id).Exec(ctx)
}
