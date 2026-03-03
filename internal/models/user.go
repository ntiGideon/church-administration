package models

import (
	"context"
	"errors"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/contact"
	"github.com/ntiGideon/ent/user"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	Db *ent.Client
}

// Authenticate verifies email + password. Returns the user on success.
func (m *UserModel) Authenticate(ctx context.Context, email, password string) (*ent.User, error) {
	u, err := m.Db.User.Query().
		Where(user.EmailEQ(email), user.IsActiveEQ(true)).
		WithContact().
		WithChurch().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	now := time.Now()
	_, _ = m.Db.User.UpdateOne(u).SetLastLogin(now).Save(ctx)

	return u, nil
}

// GetByID fetches a user with their contact, church, and custom_role edges loaded.
func (m *UserModel) GetByID(ctx context.Context, id int) (*ent.User, error) {
	u, err := m.Db.User.Query().
		Where(user.IDEQ(id), user.IsActiveEQ(true)).
		WithContact().
		WithChurch().
		WithCustomRole().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

// Create registers a new user from an invitation token (church admin registration).
// It creates a Contact + User in a transaction and links the user to the church.
func (m *UserModel) Create(ctx context.Context, dto *RegisterDto, churchID int, role user.Role) (int, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), 12)
	if err != nil {
		return 0, BcryptError
	}

	tx, err := m.Db.Tx(ctx)
	if err != nil {
		return 0, err
	}

	c, err := tx.Contact.Create().
		SetFirstName(dto.FirstName).
		SetLastName(dto.LastName).
		SetNillablePhone(nullStr(dto.Phone)).
		SetEmail(dto.Email).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return 0, CreationError
	}

	ub := tx.User.Create().
		SetEmail(dto.Email).
		SetPasswordHash(string(hash)).
		SetRole(role).
		SetContactID(c.ID)

	if churchID > 0 {
		ub = ub.SetChurchID(churchID)
	}

	u, err := ub.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return 0, CreationError
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return u.ID, nil
}

// CreateMember creates a member profile (contact) without a user account, or
// creates a minimal member user when called from an invitation accept flow.
func (m *UserModel) CreateMemberProfile(ctx context.Context, dto *MemberDto, churchID int) (*ent.Contact, error) {
	cb := m.Db.Contact.Create().
		SetFirstName(dto.FirstName).
		SetLastName(dto.LastName).
		SetNillableMiddleName(nullStr(dto.MiddleName)).
		SetNillableEmail(nullStr(dto.Email)).
		SetNillablePhone(nullStr(dto.Phone)).
		SetNillableOccupation(nullStr(dto.Occupation)).
		SetNillableAddressLine1(nullStr(dto.AddressLine1)).
		SetNillableCity(nullStr(dto.City)).
		SetNillableCountry(nullStr(dto.Country))

	if dto.Gender != "" {
		cb = cb.SetGender(contact.Gender(dto.Gender))
	}
	if dto.MaritalStatus != "" {
		cb = cb.SetMaritalStatus(contact.MaritalStatus(dto.MaritalStatus))
	}

	c, err := cb.Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return c, nil
}

// ListByChurch returns all users that belong to a given church.
func (m *UserModel) ListByChurch(ctx context.Context, churchID int) ([]*ent.User, error) {
	return m.Db.User.Query().
		Where(user.HasChurchWith()).
		WithContact().
		All(ctx)
}

// nullStr returns a pointer to s if non-empty, else nil.
func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
