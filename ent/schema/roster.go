package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Roster is a scheduled duty list for a specific service or event date.
type Roster struct {
	ent.Schema
}

func (Roster) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"),
		field.Time("service_date"),
		field.String("department").Optional(),
		field.String("notes").Optional(),
		field.Int("church_id"),
		field.Time("created_at").Default(time.Now),
	}
}

func (Roster) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("church", Church.Type).
			Ref("rosters").
			Field("church_id").
			Required().
			Unique(),
		edge.To("entries", RosterEntry.Type),
	}
}
