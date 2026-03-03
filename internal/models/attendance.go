package models

import (
	"context"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/attendance"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/contact"
	"github.com/ntiGideon/ent/event"
)

type AttendanceModel struct {
	Db *ent.Client
}

// CheckIn records a member as attending an event.
// Returns ErrDuplicateAttendance if the member is already checked in.
func (m *AttendanceModel) CheckIn(ctx context.Context, eventID, contactID int, status, notes string) (*ent.Attendance, error) {
	s := attendance.StatusPresent
	if status == "late" {
		s = attendance.StatusLate
	}

	b := m.Db.Attendance.Create().
		SetEventID(eventID).
		SetContactID(contactID).
		SetStatus(s)

	if notes != "" {
		b = b.SetNotes(notes)
	}

	a, err := b.Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return a, nil
}

// Remove deletes an attendance record by its ID.
func (m *AttendanceModel) Remove(ctx context.Context, id int) error {
	return m.Db.Attendance.DeleteOneID(id).Exec(ctx)
}

// ListByEvent returns all attendance records for an event, with contact edges loaded,
// ordered by check-in time ascending.
func (m *AttendanceModel) ListByEvent(ctx context.Context, eventID int) ([]*ent.Attendance, error) {
	return m.Db.Attendance.Query().
		Where(attendance.HasEventWith(event.IDEQ(eventID))).
		WithContact().
		Order(ent.Asc(attendance.FieldCheckInTime)).
		All(ctx)
}

// CountByEvent returns the total number of attendance records for an event.
func (m *AttendanceModel) CountByEvent(ctx context.Context, eventID int) (int, error) {
	return m.Db.Attendance.Query().
		Where(attendance.HasEventWith(event.IDEQ(eventID))).
		Count(ctx)
}

// ListByContact returns all attendance records for a contact, with event edges loaded,
// ordered by check-in time descending (most recent first).
func (m *AttendanceModel) ListByContact(ctx context.Context, contactID int) ([]*ent.Attendance, error) {
	return m.Db.Attendance.Query().
		Where(attendance.HasContactWith(contact.IDEQ(contactID))).
		WithEvent().
		Order(ent.Desc(attendance.FieldCheckInTime)).
		All(ctx)
}

// IsCheckedIn reports whether a contact already has an attendance record for an event.
func (m *AttendanceModel) IsCheckedIn(ctx context.Context, eventID, contactID int) (bool, error) {
	return m.Db.Attendance.Query().
		Where(
			attendance.HasEventWith(event.IDEQ(eventID)),
			attendance.HasContactWith(contact.IDEQ(contactID)),
		).
		Exist(ctx)
}

// MonthlyAttendanceItem holds per-month attendance totals.
type MonthlyAttendanceItem struct {
	Month   string `json:"month"`
	Present int    `json:"present"`
	Late    int    `json:"late"`
	Total   int    `json:"total"`
}

// MonthlyAttendanceTrend returns per-month attendance counts for the last n months,
// slotted by the start_time of the associated event.
func (m *AttendanceModel) MonthlyAttendanceTrend(ctx context.Context, churchID, months int) ([]MonthlyAttendanceItem, error) {
	if months <= 0 {
		months = 12
	}

	now := time.Now()
	type slotKey struct {
		year  int
		month time.Month
	}

	result := make([]MonthlyAttendanceItem, months)
	slotMap := make(map[slotKey]int, months)

	for i := 0; i < months; i++ {
		t := now.AddDate(0, -(months-1-i), 0)
		k := slotKey{t.Year(), t.Month()}
		label := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC).Format("Jan '06")
		result[i] = MonthlyAttendanceItem{Month: label}
		slotMap[k] = i
	}

	earliestT := now.AddDate(0, -(months-1), 0)
	earliest := time.Date(earliestT.Year(), earliestT.Month(), 1, 0, 0, 0, 0, time.UTC)

	records, err := m.Db.Attendance.Query().
		Where(
			attendance.HasEventWith(
				event.StartTimeGTE(earliest),
				event.HasChurchWith(church.IDEQ(churchID)),
			),
		).
		WithEvent().
		All(ctx)
	if err != nil {
		return result, nil
	}

	for _, a := range records {
		if a.Edges.Event == nil {
			continue
		}
		t := a.Edges.Event.StartTime
		k := slotKey{t.Year(), t.Month()}
		idx, ok := slotMap[k]
		if !ok {
			continue
		}
		result[idx].Total++
		if a.Status == attendance.StatusLate {
			result[idx].Late++
		} else {
			result[idx].Present++
		}
	}
	return result, nil
}
