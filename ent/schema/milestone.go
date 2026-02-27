package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Milestone struct {
	ent.Schema
}

func (Milestone) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("milestone_type").Values(
			"baby_dedication",
			"confirmation",
			"membership",
			"marriage",
			"ordination",
			"other",
		),
		field.Time("event_date"),
		field.Text("description").Optional(),
		field.String("officiated_by").Optional(),
		field.Int("contact_id"),
		field.Int("church_id"),
		field.Time("created_at").Default(time.Now),
	}
}

func (Milestone) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("member", Contact.Type).
			Ref("milestones").
			Field("contact_id").
			Unique().
			Required(),
		edge.From("church", Church.Type).
			Ref("milestones").
			Field("church_id").
			Unique().
			Required(),
	}
}
