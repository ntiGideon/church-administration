package models

import (
	"context"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/invitation"
)

type InvitationModel struct {
	Db *ent.Client
}

// Create saves a new pending invitation record.
func (m *InvitationModel) Create(ctx context.Context, churchID, inviterID int, dto *MemberInviteDto, token string, expiresAt time.Time) (*ent.Invitation, error) {
	// When a custom role is chosen use "member" as the base enum role.
	roleVal := dto.Role
	if dto.CustomRoleID > 0 {
		roleVal = "member"
	}

	b := m.Db.Invitation.Create().
		SetInviteeEmail(dto.Email).
		SetRole(invitation.Role(roleVal)).
		SetToken(token).
		SetChurchID(churchID).
		SetExpiresAt(expiresAt)

	if dto.Name != "" {
		b = b.SetInviteeName(dto.Name)
	}
	if inviterID > 0 {
		b = b.SetInviterID(inviterID)
	}
	if dto.CustomRoleID > 0 {
		b = b.SetCustomRoleID(dto.CustomRoleID)
	}

	inv, err := b.Save(ctx)
	if err != nil {
		return nil, CreationError
	}
	return inv, nil
}

// GetByToken looks up a pending, non-expired invitation by its token.
func (m *InvitationModel) GetByToken(ctx context.Context, token string) (*ent.Invitation, error) {
	inv, err := m.Db.Invitation.Query().
		Where(
			invitation.TokenEQ(token),
			invitation.StatusEQ(invitation.StatusPending),
		).
		WithChurch().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}
	if time.Now().After(inv.ExpiresAt) {
		return nil, ErrInvitationExpired
	}
	return inv, nil
}

// MarkAccepted sets the invitation status to accepted and links the created user.
func (m *InvitationModel) MarkAccepted(ctx context.Context, invID, userID int) error {
	_, err := m.Db.Invitation.UpdateOneID(invID).
		SetStatus(invitation.StatusAccepted).
		SetAcceptedUserID(userID).
		Save(ctx)
	return err
}

// ListByChurch returns all invitations for a church, newest first.
func (m *InvitationModel) ListByChurch(ctx context.Context, churchID int) ([]*ent.Invitation, error) {
	return m.Db.Invitation.Query().
		Where(invitation.HasChurchWith()).
		Order(ent.Desc(invitation.FieldCreatedAt)).
		All(ctx)
}
