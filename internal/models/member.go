package models

import (
	"context"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/contact"
	"github.com/ntiGideon/ent/user"
)

type MemberModel struct {
	Db *ent.Client
}

// ListByChurch returns all users (members/staff) belonging to a church.
// If churchID is 0, returns all users across all churches (super_admin view).
func (m *MemberModel) ListByChurch(ctx context.Context, churchID int) ([]*ent.User, error) {
	q := m.Db.User.Query().WithContact().WithChurch()
	if churchID > 0 {
		q = q.Where(user.HasChurchWith(church.IDEQ(churchID)))
	}
	return q.Order(ent.Asc(user.FieldCreatedAt)).All(ctx)
}

// CountByChurch returns the number of members in a church.
// If churchID is 0, counts all users.
func (m *MemberModel) CountByChurch(ctx context.Context, churchID int) (int, error) {
	q := m.Db.User.Query()
	if churchID > 0 {
		q = q.Where(user.HasChurchWith(church.IDEQ(churchID)))
	}
	return q.Count(ctx)
}

// GetByID returns a user (member) with all edges loaded, including spouse.
func (m *MemberModel) GetByID(ctx context.Context, id int) (*ent.User, error) {
	u, err := m.Db.User.Query().
		Where(user.IDEQ(id)).
		WithContact(func(q *ent.ContactQuery) {
			q.WithSpouseContact()
		}).
		WithChurch().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

// UpdateContact saves all editable contact fields for a given contact ID.
// The spouse link is handled separately by the caller to allow bidirectional sync.
func (m *MemberModel) UpdateContact(ctx context.Context, contactID int, dto *MemberDto) error {
	upd := m.Db.Contact.UpdateOneID(contactID).
		SetFirstName(dto.FirstName).
		SetLastName(dto.LastName).
		SetNillableMiddleName(nullStr(dto.MiddleName)).
		SetNillableEmail(nullStr(dto.Email)).
		SetNillablePhone(nullStr(dto.Phone)).
		SetNillableOccupation(nullStr(dto.Occupation)).
		SetNillableAddressLine1(nullStr(dto.AddressLine1)).
		SetNillableCity(nullStr(dto.City)).
		SetNillableCountry(nullStr(dto.Country)).
		SetNillableIDNumber(nullStr(dto.IDNumber)).
		SetNillableHometown(nullStr(dto.Hometown)).
		SetNillableRegion(nullStr(dto.Region)).
		SetNillableSundaySchoolClass(nullStr(dto.SundaySchoolClass)).
		SetHasSpouse(dto.HasSpouse).
		SetIsBaptized(dto.IsBaptized).
		SetNillableBaptizedBy(nullStr(dto.BaptizedBy)).
		SetNillableBaptismChurch(nullStr(dto.BaptismChurch)).
		SetNillableBaptismCertNumber(nullStr(dto.BaptismCertNumber))

	if dto.Gender != "" {
		upd = upd.SetGender(contact.Gender(dto.Gender))
	} else {
		upd = upd.ClearGender()
	}
	if dto.MaritalStatus != "" {
		upd = upd.SetMaritalStatus(contact.MaritalStatus(dto.MaritalStatus))
	} else {
		upd = upd.ClearMaritalStatus()
	}
	if dto.DayBorn != "" {
		upd = upd.SetDayBorn(contact.DayBorn(dto.DayBorn))
	} else {
		upd = upd.ClearDayBorn()
	}
	if dto.DateOfBirth != "" {
		if dob, err := time.Parse("2006-01-02", dto.DateOfBirth); err == nil {
			upd = upd.SetDateOfBirth(dob)
		}
	}
	if dto.BaptismDate != "" {
		if bd, err := time.Parse("2006-01-02", dto.BaptismDate); err == nil {
			upd = upd.SetBaptismDate(bd)
		}
	}
	if dto.MembershipYear > 0 {
		upd = upd.SetMembershipYear(dto.MembershipYear)
	}

	_, err := upd.Save(ctx)
	return err
}

// Deactivate sets a user as inactive.
func (m *MemberModel) Deactivate(ctx context.Context, id int) error {
	_, err := m.Db.User.UpdateOneID(id).SetIsActive(false).Save(ctx)
	return err
}
