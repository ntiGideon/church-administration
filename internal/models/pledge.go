package models

import (
	"context"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/contact"
	"github.com/ntiGideon/ent/pledge"
)

type PledgeModel struct {
	Db *ent.Client
}

// Create records a new pledge for a contact.
func (m *PledgeModel) Create(ctx context.Context, dto *PledgeDto, contactID, churchID int) (*ent.Pledge, error) {
	startDate, err := time.Parse("2006-01-02", dto.StartDate)
	if err != nil {
		startDate = time.Now()
	}

	currency := dto.Currency
	if currency == "" {
		currency = "GHS"
	}
	category := dto.Category
	if category == "" {
		category = "General Fund"
	}

	b := m.Db.Pledge.Create().
		SetContactID(contactID).
		SetChurchID(churchID).
		SetTitle(dto.Title).
		SetCategory(category).
		SetAmount(dto.Amount).
		SetCurrency(currency).
		SetStartDate(startDate).
		SetNillableNotes(nullStr(dto.Notes))

	if dto.Frequency != "" {
		b = b.SetFrequency(pledge.Frequency(dto.Frequency))
	}
	if dto.EndDate != "" {
		if endDate, err := time.Parse("2006-01-02", dto.EndDate); err == nil {
			b = b.SetEndDate(endDate)
		}
	}

	p, err := b.Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return p, nil
}

// ListByContact returns all pledges for a specific contact, newest first.
func (m *PledgeModel) ListByContact(ctx context.Context, contactID int) ([]*ent.Pledge, error) {
	return m.Db.Pledge.Query().
		Where(pledge.HasContactWith(contact.IDEQ(contactID))).
		Order(ent.Desc(pledge.FieldCreatedAt)).
		All(ctx)
}

// Delete removes a pledge by ID.
func (m *PledgeModel) Delete(ctx context.Context, id int) error {
	return m.Db.Pledge.DeleteOneID(id).Exec(ctx)
}
