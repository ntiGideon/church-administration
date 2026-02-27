package models

import (
	"context"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/contact"
	"github.com/ntiGideon/ent/department"
)

type DepartmentModel struct {
	Db *ent.Client
}

func (m *DepartmentModel) Create(ctx context.Context, dto *DepartmentDto, churchID int) (*ent.Department, error) {
	b := m.Db.Department.Create().
		SetName(dto.Name).
		SetDescription(dto.Description).
		SetDepartmentType(department.DepartmentType(dto.DepartmentType)).
		SetIsActive(true).
		SetChurchID(churchID)

	if dto.LeaderID > 0 {
		b = b.SetLeaderID(dto.LeaderID)
	}

	d, err := b.Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return d, nil
}

func (m *DepartmentModel) GetByID(ctx context.Context, id int) (*ent.Department, error) {
	d, err := m.Db.Department.Query().
		Where(department.IDEQ(id)).
		WithLeader().
		WithMembers().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return d, nil
}

func (m *DepartmentModel) ListByChurch(ctx context.Context, churchID int) ([]*ent.Department, error) {
	return m.Db.Department.Query().
		Where(department.HasChurchWith(church.IDEQ(churchID))).
		WithLeader().
		WithMembers().
		Order(ent.Asc(department.FieldName)).
		All(ctx)
}

func (m *DepartmentModel) Update(ctx context.Context, id int, dto *DepartmentDto) error {
	upd := m.Db.Department.UpdateOneID(id).
		SetName(dto.Name).
		SetDescription(dto.Description).
		SetIsActive(dto.IsActive)

	if dto.DepartmentType != "" {
		upd = upd.SetDepartmentType(department.DepartmentType(dto.DepartmentType))
	}
	if dto.LeaderID > 0 {
		upd = upd.SetLeaderID(dto.LeaderID)
	} else {
		upd = upd.ClearLeader()
	}

	_, err := upd.Save(ctx)
	return err
}

func (m *DepartmentModel) Delete(ctx context.Context, id int) error {
	return m.Db.Department.DeleteOneID(id).Exec(ctx)
}

func (m *DepartmentModel) AddMember(ctx context.Context, deptID, contactID int) error {
	_, err := m.Db.Department.UpdateOneID(deptID).
		AddMemberIDs(contactID).
		Save(ctx)
	return err
}

func (m *DepartmentModel) RemoveMember(ctx context.Context, deptID, contactID int) error {
	_, err := m.Db.Department.UpdateOneID(deptID).
		RemoveMemberIDs(contactID).
		Save(ctx)
	return err
}

func (m *DepartmentModel) IsMember(ctx context.Context, deptID, contactID int) (bool, error) {
	return m.Db.Department.Query().
		Where(
			department.IDEQ(deptID),
			department.HasMembersWith(contact.IDEQ(contactID)),
		).
		Exist(ctx)
}
