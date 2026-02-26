package models

import (
	"context"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/contact"
	"github.com/ntiGideon/ent/group"
)

type GroupModel struct {
	Db *ent.Client
}

// Create adds a new group for a church.
func (m *GroupModel) Create(ctx context.Context, dto *GroupDto, churchID int) (*ent.Group, error) {
	b := m.Db.Group.Create().
		SetName(dto.Name).
		SetNillableDescription(nullStr(dto.Description)).
		SetChurchID(churchID).
		SetIsActive(true)

	if dto.GroupType != "" {
		b = b.SetGroupType(group.GroupType(dto.GroupType))
	}
	if dto.MeetingDay != "" {
		b = b.SetMeetingDay(group.MeetingDay(dto.MeetingDay))
	}
	if dto.MeetingTime != "" {
		b = b.SetMeetingTime(dto.MeetingTime)
	}
	if dto.Location != "" {
		b = b.SetLocation(dto.Location)
	}
	if dto.LeaderID > 0 {
		b = b.SetLeaderID(dto.LeaderID)
	}

	g, err := b.Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return g, nil
}

// GetByID returns a group with leader and members edges loaded.
func (m *GroupModel) GetByID(ctx context.Context, id int) (*ent.Group, error) {
	g, err := m.Db.Group.Query().
		Where(group.IDEQ(id)).
		WithLeader().
		WithMembers().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return g, nil
}

// ListByChurch returns all groups for a church, with member count loaded.
func (m *GroupModel) ListByChurch(ctx context.Context, churchID int) ([]*ent.Group, error) {
	q := m.Db.Group.Query().
		WithLeader().
		WithMembers()
	if churchID > 0 {
		q = q.Where(group.HasChurchWith(church.IDEQ(churchID)))
	}
	return q.Order(ent.Asc(group.FieldName)).All(ctx)
}

// Update saves editable fields on a group.
func (m *GroupModel) Update(ctx context.Context, id int, dto *GroupDto) error {
	upd := m.Db.Group.UpdateOneID(id).
		SetName(dto.Name).
		SetNillableDescription(nullStr(dto.Description)).
		SetNillableMeetingTime(nullStr(dto.MeetingTime)).
		SetNillableLocation(nullStr(dto.Location)).
		SetIsActive(dto.IsActive)

	if dto.GroupType != "" {
		upd = upd.SetGroupType(group.GroupType(dto.GroupType))
	}
	if dto.MeetingDay != "" {
		upd = upd.SetMeetingDay(group.MeetingDay(dto.MeetingDay))
	} else {
		upd = upd.ClearMeetingDay()
	}
	if dto.LeaderID > 0 {
		upd = upd.SetLeaderID(dto.LeaderID)
	} else {
		upd = upd.ClearLeader()
	}

	_, err := upd.Save(ctx)
	return err
}

// Delete removes a group.
func (m *GroupModel) Delete(ctx context.Context, id int) error {
	return m.Db.Group.DeleteOneID(id).Exec(ctx)
}

// AddMember adds a contact to a group.
func (m *GroupModel) AddMember(ctx context.Context, groupID, contactID int) error {
	_, err := m.Db.Group.UpdateOneID(groupID).
		AddMemberIDs(contactID).
		Save(ctx)
	return err
}

// RemoveMember removes a contact from a group.
func (m *GroupModel) RemoveMember(ctx context.Context, groupID, contactID int) error {
	_, err := m.Db.Group.UpdateOneID(groupID).
		RemoveMemberIDs(contactID).
		Save(ctx)
	return err
}

// IsMember reports whether a contact is already in a group.
func (m *GroupModel) IsMember(ctx context.Context, groupID, contactID int) (bool, error) {
	return m.Db.Group.Query().
		Where(
			group.IDEQ(groupID),
			group.HasMembersWith(contact.IDEQ(contactID)),
		).
		Exist(ctx)
}
