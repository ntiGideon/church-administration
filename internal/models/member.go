package models

import (
	"context"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
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

// GetByID returns a user (member) with all edges loaded.
func (m *MemberModel) GetByID(ctx context.Context, id int) (*ent.User, error) {
	u, err := m.Db.User.Query().
		Where(user.IDEQ(id)).
		WithContact().
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

// Deactivate sets a user as inactive.
func (m *MemberModel) Deactivate(ctx context.Context, id int) error {
	_, err := m.Db.User.UpdateOneID(id).SetIsActive(false).Save(ctx)
	return err
}
