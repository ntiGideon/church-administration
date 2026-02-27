package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Department struct {
	ent.Schema
}

func (Department) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("description").Optional(),
		field.Enum("department_type").Values(
			"worship",
			"youth",
			"children",
			"outreach",
			"administration",
			"finance",
			"media",
			"other",
		),
		field.Bool("is_active").Default(true),
		field.Int("church_id"),
		field.Int("leader_id").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Department) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("church", Church.Type).
			Ref("departments").
			Field("church_id").
			Unique().
			Required(),
		edge.To("leader", Contact.Type).
			Field("leader_id").
			Unique(),
		edge.To("members", Contact.Type),
	}
}
