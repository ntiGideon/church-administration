package models

import (
	"context"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/contact"
	"github.com/ntiGideon/ent/relationship"
)

type RelationshipModel struct {
	Db *ent.Client
}

// Create adds a relationship from one contact to another.
// Returns ErrDuplicateRelationship if the pair already exists.
func (m *RelationshipModel) Create(ctx context.Context, fromID, toID int, relType, notes string) (*ent.Relationship, error) {
	b := m.Db.Relationship.Create().
		SetFromContactID(fromID).
		SetToContactID(toID).
		SetRelationType(relationship.RelationType(relType))

	if notes != "" {
		b = b.SetNotes(notes)
	}

	r, err := b.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, ErrDuplicateRelationship
		}
		return nil, CreationError
	}
	return r, nil
}

// ListByContact returns all relationships involving this contact (either as from or to),
// with both contact edges loaded.
func (m *RelationshipModel) ListByContact(ctx context.Context, contactID int) ([]*ent.Relationship, error) {
	return m.Db.Relationship.Query().
		Where(
			relationship.Or(
				relationship.HasFromContactWith(contact.IDEQ(contactID)),
				relationship.HasToContactWith(contact.IDEQ(contactID)),
			),
		).
		WithFromContact().
		WithToContact().
		Order(ent.Asc(relationship.FieldCreatedAt)).
		All(ctx)
}

// Delete removes a relationship record by ID.
func (m *RelationshipModel) Delete(ctx context.Context, id int) error {
	return m.Db.Relationship.DeleteOneID(id).Exec(ctx)
}
