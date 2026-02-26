package models

import (
	"context"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/programentry"
)

type ProgramModel struct {
	Db *ent.Client
}

// ListByMonthYear returns all program entries for a church in the given month/year, ordered by date.
// If churchID is 0, returns across all churches (super_admin view).
func (m *ProgramModel) ListByMonthYear(ctx context.Context, churchID, year, month int) ([]*ent.ProgramEntry, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	q := m.Db.ProgramEntry.Query().
		Where(
			programentry.DateGTE(start),
			programentry.DateLT(end),
		)
	if churchID > 0 {
		q = q.Where(programentry.HasChurchWith(church.IDEQ(churchID)))
	} else {
		q = q.Where(programentry.HasChurchWith())
	}
	return q.WithChurch().Order(ent.Asc(programentry.FieldDate)).All(ctx)
}

// Create saves a new program entry.
func (m *ProgramModel) Create(ctx context.Context, dto *ProgramDto, churchID int) (*ent.ProgramEntry, error) {
	date, err := time.Parse("2006-01-02", dto.Date)
	if err != nil {
		date = time.Now()
	}

	p, err := m.Db.ProgramEntry.Create().
		SetTitle(dto.Title).
		SetProgramType(programentry.ProgramType(dto.ProgramType)).
		SetDate(date).
		SetNillableTheme(nullStr(dto.Theme)).
		SetNillableSermonTopic(nullStr(dto.SermonTopic)).
		SetNillableVisionGoals(nullStr(dto.VisionGoals)).
		SetNillablePreacher(nullStr(dto.Preacher)).
		SetNillableOpeningPrayerBy(nullStr(dto.OpeningPrayerBy)).
		SetNillableClosingPrayerBy(nullStr(dto.ClosingPrayerBy)).
		SetNillableWorshipLeader(nullStr(dto.WorshipLeader)).
		SetNillableResponsiblePerson(nullStr(dto.ResponsiblePerson)).
		SetNillableNotes(nullStr(dto.Notes)).
		SetIsPublished(dto.IsPublished).
		SetChurchID(churchID).
		Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return p, nil
}

// GetByID fetches a single program entry with its church edge.
func (m *ProgramModel) GetByID(ctx context.Context, id int) (*ent.ProgramEntry, error) {
	p, err := m.Db.ProgramEntry.Query().
		Where(programentry.IDEQ(id)).
		WithChurch().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return p, nil
}

// Update saves changes to an existing program entry.
func (m *ProgramModel) Update(ctx context.Context, id int, dto *ProgramDto) (*ent.ProgramEntry, error) {
	date, err := time.Parse("2006-01-02", dto.Date)
	if err != nil {
		date = time.Now()
	}

	p, err := m.Db.ProgramEntry.UpdateOneID(id).
		SetTitle(dto.Title).
		SetProgramType(programentry.ProgramType(dto.ProgramType)).
		SetDate(date).
		SetNillableTheme(nullStr(dto.Theme)).
		SetNillableSermonTopic(nullStr(dto.SermonTopic)).
		SetNillableVisionGoals(nullStr(dto.VisionGoals)).
		SetNillablePreacher(nullStr(dto.Preacher)).
		SetNillableOpeningPrayerBy(nullStr(dto.OpeningPrayerBy)).
		SetNillableClosingPrayerBy(nullStr(dto.ClosingPrayerBy)).
		SetNillableWorshipLeader(nullStr(dto.WorshipLeader)).
		SetNillableResponsiblePerson(nullStr(dto.ResponsiblePerson)).
		SetNillableNotes(nullStr(dto.Notes)).
		SetIsPublished(dto.IsPublished).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Delete removes a program entry.
func (m *ProgramModel) Delete(ctx context.Context, id int) error {
	return m.Db.ProgramEntry.DeleteOneID(id).Exec(ctx)
}
