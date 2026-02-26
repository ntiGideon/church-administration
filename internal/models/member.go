package models

import (
	"context"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/contact"
	"github.com/ntiGideon/ent/user"
)

// ─── Congregation contact methods ────────────────────────────────────────────

// ListContactsByChurch returns congregation Contact records for a church.
// If churchID is 0, returns all contacts linked to any church (super_admin view).
func (m *MemberModel) ListContactsByChurch(ctx context.Context, churchID int) ([]*ent.Contact, error) {
	q := m.Db.Contact.Query().WithSpouseContact()
	if churchID > 0 {
		q = q.Where(contact.HasChurchWith(church.IDEQ(churchID)))
	} else {
		q = q.Where(contact.HasChurchWith())
	}
	return q.Order(ent.Asc(contact.FieldFirstName)).All(ctx)
}

// CountContactsByChurch counts congregation contacts for a church.
func (m *MemberModel) CountContactsByChurch(ctx context.Context, churchID int) (int, error) {
	q := m.Db.Contact.Query()
	if churchID > 0 {
		q = q.Where(contact.HasChurchWith(church.IDEQ(churchID)))
	} else {
		q = q.Where(contact.HasChurchWith())
	}
	return q.Count(ctx)
}

// GetContactByID returns a single Contact with spouse and church edges loaded.
func (m *MemberModel) GetContactByID(ctx context.Context, id int) (*ent.Contact, error) {
	c, err := m.Db.Contact.Query().
		Where(contact.IDEQ(id)).
		WithSpouseContact().
		WithChurch().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return c, nil
}

// CreateContact adds a new congregation member record directly (no system account).
func (m *MemberModel) CreateContact(ctx context.Context, dto *MemberDto, churchID int) (*ent.Contact, error) {
	b := m.Db.Contact.Create().
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
		SetNillableBaptismCertNumber(nullStr(dto.BaptismCertNumber)).
		SetChurchID(churchID)
	if dto.Gender != "" {
		b = b.SetGender(contact.Gender(dto.Gender))
	}
	if dto.MaritalStatus != "" {
		b = b.SetMaritalStatus(contact.MaritalStatus(dto.MaritalStatus))
	}
	if dto.DayBorn != "" {
		b = b.SetDayBorn(contact.DayBorn(dto.DayBorn))
	}
	if dto.DateOfBirth != "" {
		if dob, err := time.Parse("2006-01-02", dto.DateOfBirth); err == nil {
			b = b.SetDateOfBirth(dob)
		}
	}
	if dto.BaptismDate != "" {
		if bd, err := time.Parse("2006-01-02", dto.BaptismDate); err == nil {
			b = b.SetBaptismDate(bd)
		}
	}
	if dto.MembershipYear > 0 {
		b = b.SetMembershipYear(dto.MembershipYear)
	}
	c, err := b.Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return c, nil
}

// DeleteContact permanently removes a congregation contact record.
func (m *MemberModel) DeleteContact(ctx context.Context, id int) error {
	return m.Db.Contact.DeleteOneID(id).Exec(ctx)
}

// ListWorkersByChurch returns all users (workers/staff) belonging to a church.
func (m *MemberModel) ListWorkersByChurch(ctx context.Context, churchID int) ([]*ent.User, error) {
	return m.ListByChurch(ctx, churchID)
}

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
