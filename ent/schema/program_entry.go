package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProgramEntry holds the schema definition for a church calendar entry.
type ProgramEntry struct {
	ent.Schema
}

func (ProgramEntry) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"),
		field.Enum("program_type").Values(
			"service",
			"prayer_meeting",
			"bible_study",
			"conference",
			"outreach",
			"youth_service",
			"special_event",
			"other",
		),
		field.Time("date"),

		// Service content
		field.String("theme").Optional(),
		field.String("sermon_topic").Optional(),
		field.String("vision_goals").Optional(),

		// Personnel
		field.String("preacher").Optional(),
		field.String("opening_prayer_by").Optional(),
		field.String("closing_prayer_by").Optional(),
		field.String("worship_leader").Optional(),
		field.String("responsible_person").Optional(),

		// Extra
		field.String("notes").Optional(),
		field.Bool("is_published").Default(false),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (ProgramEntry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("church", Church.Type).
			Ref("programs").
			Unique(),
	}
}

func (ProgramEntry) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("date"),
	}
}
