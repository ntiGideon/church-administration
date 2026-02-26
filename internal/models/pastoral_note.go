package models

import (
	"context"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/contact"
	"github.com/ntiGideon/ent/pastoralnote"
)

type PastoralNoteModel struct {
	Db *ent.Client
}

func (m *PastoralNoteModel) Create(ctx context.Context, dto *PastoralNoteDto, contactID, churchID, userID int) (*ent.PastoralNote, error) {
	visitDate, err := time.Parse("2006-01-02", dto.VisitDate)
	if err != nil {
		visitDate = time.Now()
	}

	b := m.Db.PastoralNote.Create().
		SetVisitDate(visitDate).
		SetCareType(pastoralnote.CareType(dto.CareType)).
		SetNotes(dto.Notes).
		SetNeedsFollowUp(dto.NeedsFollowUp).
		SetFollowUpDone(false).
		SetContactID(contactID).
		SetChurchID(churchID)

	if dto.NeedsFollowUp && dto.FollowUpDate != "" {
		if t, err := time.Parse("2006-01-02", dto.FollowUpDate); err == nil {
			b = b.SetFollowUpDate(t)
		}
	}

	if userID > 0 {
		b = b.SetRecordedByID(userID)
	}

	n, err := b.Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return n, nil
}

func (m *PastoralNoteModel) GetByID(ctx context.Context, id int) (*ent.PastoralNote, error) {
	n, err := m.Db.PastoralNote.Query().
		Where(pastoralnote.IDEQ(id)).
		WithMember().
		WithRecorder(func(uq *ent.UserQuery) { uq.WithContact() }).
		First(ctx)
	if err != nil {
		return nil, ErrRecordNotFound
	}
	return n, nil
}

func (m *PastoralNoteModel) ListByChurch(ctx context.Context, churchID int) ([]*ent.PastoralNote, error) {
	return m.Db.PastoralNote.Query().
		Where(pastoralnote.HasChurchWith(church.IDEQ(churchID))).
		Order(ent.Desc(pastoralnote.FieldVisitDate)).
		WithMember().
		WithRecorder(func(uq *ent.UserQuery) { uq.WithContact() }).
		All(ctx)
}

func (m *PastoralNoteModel) ListByContact(ctx context.Context, contactID int) ([]*ent.PastoralNote, error) {
	return m.Db.PastoralNote.Query().
		Where(pastoralnote.HasMemberWith(contact.IDEQ(contactID))).
		Order(ent.Desc(pastoralnote.FieldVisitDate)).
		WithRecorder(func(uq *ent.UserQuery) { uq.WithContact() }).
		All(ctx)
}

func (m *PastoralNoteModel) Update(ctx context.Context, id int, dto *PastoralNoteDto) (*ent.PastoralNote, error) {
	visitDate, err := time.Parse("2006-01-02", dto.VisitDate)
	if err != nil {
		visitDate = time.Now()
	}

	b := m.Db.PastoralNote.UpdateOneID(id).
		SetVisitDate(visitDate).
		SetCareType(pastoralnote.CareType(dto.CareType)).
		SetNotes(dto.Notes).
		SetNeedsFollowUp(dto.NeedsFollowUp)

	if dto.NeedsFollowUp && dto.FollowUpDate != "" {
		if t, err := time.Parse("2006-01-02", dto.FollowUpDate); err == nil {
			b = b.SetFollowUpDate(t)
		}
	} else {
		b = b.ClearFollowUpDate()
	}

	n, err := b.Save(ctx)
	if err != nil {
		return nil, ErrRecordNotFound
	}
	return n, nil
}

func (m *PastoralNoteModel) MarkFollowUpDone(ctx context.Context, id int) error {
	return m.Db.PastoralNote.UpdateOneID(id).
		SetFollowUpDone(true).
		SetNeedsFollowUp(false).
		Exec(ctx)
}

func (m *PastoralNoteModel) Delete(ctx context.Context, id int) error {
	return m.Db.PastoralNote.DeleteOneID(id).Exec(ctx)
}

func (m *PastoralNoteModel) CountPendingFollowUp(ctx context.Context, churchID int) int {
	n, _ := m.Db.PastoralNote.Query().
		Where(
			pastoralnote.HasChurchWith(church.IDEQ(churchID)),
			pastoralnote.NeedsFollowUp(true),
			pastoralnote.FollowUpDone(false),
		).
		Count(ctx)
	return n
}
