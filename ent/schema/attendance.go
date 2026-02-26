package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Attendance records an individual member's presence at a church event.
type Attendance struct {
	ent.Schema
}

func (Attendance) Fields() []ent.Field {
	return []ent.Field{
		field.Int("event_id"),
		field.Int("contact_id"),
		field.Enum("status").
			Values("present", "late").
			Default("present"),
		field.String("notes").Optional().Nillable(),
		field.Time("check_in_time").Default(time.Now),
	}
}

func (Attendance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", Event.Type).
			Ref("attendances").
			Field("event_id").
			Required().
			Unique(),
		edge.From("contact", Contact.Type).
			Ref("attendances").
			Field("contact_id").
			Required().
			Unique(),
	}
}

// Unique composite index prevents a member being checked in twice for the same event.
func (Attendance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("event_id", "contact_id").Unique(),
	}
}
