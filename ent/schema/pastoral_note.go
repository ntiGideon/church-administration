package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// PastoralNote holds the schema definition for pastoral care records.
type PastoralNote struct {
	ent.Schema
}

func (PastoralNote) Fields() []ent.Field {
	return []ent.Field{
		field.Time("visit_date"),
		field.Enum("care_type").Values(
			"visit",
			"counseling",
			"phone_call",
			"prayer_session",
			"hospital_visit",
			"bereavement",
			"other",
		).Default("visit"),
		field.Text("notes"),
		field.Bool("needs_follow_up").Default(false),
		field.Time("follow_up_date").Optional(),
		field.Bool("follow_up_done").Default(false),
		field.Int("contact_id"),
		field.Int("church_id"),
		field.Int("recorded_by_id").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (PastoralNote) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("church", Church.Type).
			Ref("pastoral_notes").
			Field("church_id").
			Required().
			Unique(),
		edge.From("member", Contact.Type).
			Ref("pastoral_notes").
			Field("contact_id").
			Required().
			Unique(),
		edge.From("recorder", User.Type).
			Ref("pastoral_notes_recorded").
			Field("recorded_by_id").
			Unique(),
	}
}
