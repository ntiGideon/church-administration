package models

import (
	"context"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/contact"
	"github.com/ntiGideon/ent/milestone"
)

type MilestoneModel struct {
	Db *ent.Client
}

func (m *MilestoneModel) Create(ctx context.Context, dto *MilestoneDto, contactID, churchID int) (*ent.Milestone, error) {
	eventDate, err := time.Parse("2006-01-02", dto.EventDate)
	if err != nil {
		eventDate = time.Now()
	}

	b := m.Db.Milestone.Create().
		SetMilestoneType(milestone.MilestoneType(dto.MilestoneType)).
		SetEventDate(eventDate).
		SetContactID(contactID).
		SetChurchID(churchID)

	if dto.Description != "" {
		b = b.SetDescription(dto.Description)
	}
	if dto.OfficiatedBy != "" {
		b = b.SetOfficiatedBy(dto.OfficiatedBy)
	}

	ms, err := b.Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return ms, nil
}

func (m *MilestoneModel) ListByContact(ctx context.Context, contactID int) ([]*ent.Milestone, error) {
	return m.Db.Milestone.Query().
		Where(milestone.HasMemberWith(contact.IDEQ(contactID))).
		Order(ent.Desc(milestone.FieldEventDate)).
		All(ctx)
}

func (m *MilestoneModel) ListByChurch(ctx context.Context, churchID int) ([]*ent.Milestone, error) {
	return m.Db.Milestone.Query().
		Where(milestone.HasChurchWith(church.IDEQ(churchID))).
		Order(ent.Desc(milestone.FieldEventDate)).
		WithMember().
		All(ctx)
}

func (m *MilestoneModel) Delete(ctx context.Context, id int) error {
	return m.Db.Milestone.DeleteOneID(id).Exec(ctx)
}

func (m *MilestoneModel) CountByChurch(ctx context.Context, churchID int) int {
	n, _ := m.Db.Milestone.Query().
		Where(milestone.HasChurchWith(church.IDEQ(churchID))).
		Count(ctx)
	return n
}
