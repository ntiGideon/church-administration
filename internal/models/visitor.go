package models

import (
	"context"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/visitor"
)

type VisitorModel struct {
	Db *ent.Client
}

// Create adds a new visitor record.
func (m *VisitorModel) Create(ctx context.Context, dto *VisitorDto, churchID int) (*ent.Visitor, error) {
	visitDate, err := time.Parse("2006-01-02", dto.VisitDate)
	if err != nil {
		visitDate = time.Now()
	}

	b := m.Db.Visitor.Create().
		SetFirstName(dto.FirstName).
		SetLastName(dto.LastName).
		SetNillableEmail(nullStr(dto.Email)).
		SetNillablePhone(nullStr(dto.Phone)).
		SetNillableAddress(nullStr(dto.Address)).
		SetVisitDate(visitDate).
		SetNillableInvitedBy(nullStr(dto.InvitedBy)).
		SetNillableNotes(nullStr(dto.Notes)).
		SetChurchID(churchID)

	if dto.HowHeard != "" {
		b = b.SetHowHeard(visitor.HowHeard(dto.HowHeard))
	}
	if dto.FollowUpStatus != "" {
		b = b.SetFollowUpStatus(visitor.FollowUpStatus(dto.FollowUpStatus))
	}

	v, err := b.Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return v, nil
}

// GetByID returns a single visitor.
func (m *VisitorModel) GetByID(ctx context.Context, id int) (*ent.Visitor, error) {
	v, err := m.Db.Visitor.Query().
		Where(visitor.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return v, nil
}

// ListByChurch returns all visitors for a church, most recent visit first.
func (m *VisitorModel) ListByChurch(ctx context.Context, churchID int) ([]*ent.Visitor, error) {
	q := m.Db.Visitor.Query()
	if churchID > 0 {
		q = q.Where(visitor.HasChurchWith(church.IDEQ(churchID)))
	}
	return q.Order(ent.Desc(visitor.FieldVisitDate)).All(ctx)
}

// Update saves editable fields on a visitor record.
func (m *VisitorModel) Update(ctx context.Context, id int, dto *VisitorDto) error {
	visitDate, err := time.Parse("2006-01-02", dto.VisitDate)
	if err != nil {
		visitDate = time.Now()
	}

	u := m.Db.Visitor.UpdateOneID(id).
		SetFirstName(dto.FirstName).
		SetLastName(dto.LastName).
		SetNillableEmail(nullStr(dto.Email)).
		SetNillablePhone(nullStr(dto.Phone)).
		SetNillableAddress(nullStr(dto.Address)).
		SetVisitDate(visitDate).
		SetNillableInvitedBy(nullStr(dto.InvitedBy)).
		SetNillableNotes(nullStr(dto.Notes))

	if dto.HowHeard != "" {
		u = u.SetHowHeard(visitor.HowHeard(dto.HowHeard))
	} else {
		u = u.ClearHowHeard()
	}
	if dto.FollowUpStatus != "" {
		u = u.SetFollowUpStatus(visitor.FollowUpStatus(dto.FollowUpStatus))
	}

	_, err = u.Save(ctx)
	return err
}

// UpdateStatus changes only the follow-up status of a visitor.
func (m *VisitorModel) UpdateStatus(ctx context.Context, id int, status string) error {
	_, err := m.Db.Visitor.UpdateOneID(id).
		SetFollowUpStatus(visitor.FollowUpStatus(status)).
		Save(ctx)
	return err
}

// CountByStatus returns a map of follow_up_status → count for a church.
func (m *VisitorModel) CountByStatus(ctx context.Context, churchID int) (map[string]int, error) {
	all, err := m.ListByChurch(ctx, churchID)
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	for _, v := range all {
		counts[string(v.FollowUpStatus)]++
	}
	return counts, nil
}

// Delete removes a visitor record.
func (m *VisitorModel) Delete(ctx context.Context, id int) error {
	return m.Db.Visitor.DeleteOneID(id).Exec(ctx)
}
