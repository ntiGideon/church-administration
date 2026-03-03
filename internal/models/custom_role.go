package models

import (
	"context"
	"encoding/json"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/customrole"
)

type CustomRoleModel struct {
	Db *ent.Client
}

// Create saves a new custom role for a church.
func (m *CustomRoleModel) Create(ctx context.Context, churchID int, dto *CustomRoleDto) (*ent.CustomRole, error) {
	permsJSON, err := json.Marshal(dto.Permissions)
	if err != nil {
		return nil, CreationError
	}
	r, err := m.Db.CustomRole.Create().
		SetName(dto.Name).
		SetDescription(dto.Description).
		SetPermissions(string(permsJSON)).
		SetIsActive(dto.IsActive).
		SetChurchID(churchID).
		Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return r, nil
}

// ListByChurch returns all custom roles for a church, ordered by name.
func (m *CustomRoleModel) ListByChurch(ctx context.Context, churchID int) ([]*ent.CustomRole, error) {
	return m.Db.CustomRole.Query().
		Where(customrole.ChurchIDEQ(churchID)).
		WithUsers().
		Order(ent.Asc(customrole.FieldName)).
		All(ctx)
}

// GetByID returns a single custom role by ID.
func (m *CustomRoleModel) GetByID(ctx context.Context, id int) (*ent.CustomRole, error) {
	r, err := m.Db.CustomRole.Query().
		Where(customrole.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrCustomRoleNotFound
		}
		return nil, err
	}
	return r, nil
}

// Update modifies an existing custom role.
func (m *CustomRoleModel) Update(ctx context.Context, id int, dto *CustomRoleDto) (*ent.CustomRole, error) {
	permsJSON, err := json.Marshal(dto.Permissions)
	if err != nil {
		return nil, CreationError
	}
	r, err := m.Db.CustomRole.UpdateOneID(id).
		SetName(dto.Name).
		SetDescription(dto.Description).
		SetPermissions(string(permsJSON)).
		SetIsActive(dto.IsActive).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrCustomRoleNotFound
		}
		return nil, err
	}
	return r, nil
}

// Delete removes a custom role by ID.
func (m *CustomRoleModel) Delete(ctx context.Context, id int) error {
	err := m.Db.CustomRole.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrCustomRoleNotFound
		}
		return err
	}
	return nil
}

// ParsePermissions converts the JSON permissions string into a []string slice.
func ParsePermissions(raw string) []string {
	var perms []string
	_ = json.Unmarshal([]byte(raw), &perms)
	return perms
}
