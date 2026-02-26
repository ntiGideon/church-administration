package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Pledge records a financial commitment made by a member.
type Pledge struct {
	ent.Schema
}

func (Pledge) Fields() []ent.Field {
	return []ent.Field{
		field.Int("contact_id"),
		field.Int("church_id"),
		field.String("title"),
		field.String("category").Default("General Fund"),
		field.Float("amount"),
		field.String("currency").Default("GHS"),
		field.Time("start_date"),
		field.Time("end_date").Optional(),
		field.Enum("frequency").Values(
			"one_time",
			"weekly",
			"monthly",
			"quarterly",
			"yearly",
		).Default("one_time"),
		field.String("notes").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (Pledge) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("contact", Contact.Type).
			Ref("pledges").
			Field("contact_id").
			Required().
			Unique(),
		edge.From("church", Church.Type).
			Ref("pledges").
			Field("church_id").
			Required().
			Unique(),
	}
}
