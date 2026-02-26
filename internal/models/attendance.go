package models

import (
	"context"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/attendance"
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

// IsCheckedIn reports whether a contact already has an attendance record for an event.
func (m *AttendanceModel) IsCheckedIn(ctx context.Context, eventID, contactID int) (bool, error) {
	return m.Db.Attendance.Query().
		Where(
			attendance.HasEventWith(event.IDEQ(eventID)),
			attendance.HasContactWith(contact.IDEQ(contactID)),
		).
		Exist(ctx)
}
