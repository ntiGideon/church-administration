package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Event holds the schema definition for the Event entity.
type Event struct {
	ent.Schema
}

// Fields of the Event.
func (Event) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"),
		field.String("description"),
		field.Time("start_time"),
		field.Time("end_time"),
		field.String("location"),
		field.Enum("event_type").Values(
			"service",
			"meeting",
			"conference",
			"outreach",
			"training",
		),
		field.Int("attendance_count").Default(0),
		field.String("image_url").Optional(),
		field.Bool("is_published").Default(false),
	}
}

// Edges of the Event.
func (Event) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("church", Church.Type).
			Ref("events").
			Required().
			Unique(),
	}
}
