package models

import (
	"context"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/roster"
	"github.com/ntiGideon/ent/rosterentry"
)

type RosterModel struct {
	Db *ent.Client
}

// Create creates a new roster for a church.
func (m *RosterModel) Create(ctx context.Context, dto *RosterDto, churchID int) (*ent.Roster, error) {
	svcDate, err := time.Parse("2006-01-02", dto.ServiceDate)
	if err != nil {
		svcDate = time.Now()
	}

	b := m.Db.Roster.Create().
		SetTitle(dto.Title).
		SetServiceDate(svcDate).
		SetNillableDepartment(nullStr(dto.Department)).
		SetNillableNotes(nullStr(dto.Notes)).
		SetChurchID(churchID)

	r, err := b.Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return r, nil
}

// GetByID returns a roster with its entries and each entry's contact loaded.
func (m *RosterModel) GetByID(ctx context.Context, id int) (*ent.Roster, error) {
	r, err := m.Db.Roster.Query().
		Where(roster.IDEQ(id)).
		WithEntries(func(eq *ent.RosterEntryQuery) {
			eq.WithContact()
		}).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return r, nil
}

// ListByChurch returns all rosters for a church, newest first.
func (m *RosterModel) ListByChurch(ctx context.Context, churchID int) ([]*ent.Roster, error) {
	q := m.Db.Roster.Query().
		WithEntries()
	if churchID > 0 {
		q = q.Where(roster.HasChurchWith(church.IDEQ(churchID)))
	}
	return q.Order(ent.Desc(roster.FieldServiceDate)).All(ctx)
}

// Update saves editable fields on a roster.
func (m *RosterModel) Update(ctx context.Context, id int, dto *RosterDto) error {
	svcDate, err := time.Parse("2006-01-02", dto.ServiceDate)
	if err != nil {
		svcDate = time.Now()
	}
	_, err = m.Db.Roster.UpdateOneID(id).
		SetTitle(dto.Title).
		SetServiceDate(svcDate).
		SetNillableDepartment(nullStr(dto.Department)).
		SetNillableNotes(nullStr(dto.Notes)).
		Save(ctx)
	return err
}

// Delete removes a roster and all its entries (cascade via FK).
func (m *RosterModel) Delete(ctx context.Context, id int) error {
	return m.Db.Roster.DeleteOneID(id).Exec(ctx)
}

// AddEntry assigns a volunteer to a roster in a specific role.
func (m *RosterModel) AddEntry(ctx context.Context, rosterID, contactID int, role, notes string) (*ent.RosterEntry, error) {
	b := m.Db.RosterEntry.Create().
		SetRosterID(rosterID).
		SetContactID(contactID).
		SetRole(role).
		SetNillableNotes(nullStr(notes))

	e, err := b.Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return e, nil
}

// RemoveEntry deletes a single roster assignment.
func (m *RosterModel) RemoveEntry(ctx context.Context, entryID int) error {
	return m.Db.RosterEntry.DeleteOneID(entryID).Exec(ctx)
}

// IsAssigned reports whether a contact already has an entry in the roster.
func (m *RosterModel) IsAssigned(ctx context.Context, rosterID, contactID int) (bool, error) {
	return m.Db.RosterEntry.Query().
		Where(
			rosterentry.RosterIDEQ(rosterID),
			rosterentry.ContactIDEQ(contactID),
		).
		Exist(ctx)
}
