package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RosterEntry is a single volunteer assignment within a Roster.
type RosterEntry struct {
	ent.Schema
}

func (RosterEntry) Fields() []ent.Field {
	return []ent.Field{
		field.Int("roster_id"),
		field.Int("contact_id"),
		field.String("role"),
		field.String("notes").Optional(),
	}
}

func (RosterEntry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("roster", Roster.Type).
			Ref("entries").
			Field("roster_id").
			Required().
			Unique(),
		edge.From("contact", Contact.Type).
			Ref("roster_entries").
			Field("contact_id").
			Required().
			Unique(),
	}
}

// Prevent the same contact being assigned twice in the same roster.
func (RosterEntry) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("roster_id", "contact_id").Unique(),
	}
}
