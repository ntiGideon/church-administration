package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type PrayerRequest struct {
	ent.Schema
}

func (PrayerRequest) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"),
		field.Text("body"),
		field.String("requester_name").Optional(),
		field.Bool("is_anonymous").Default(false),
		field.Bool("is_private").Default(false),
		field.Enum("status").Values(
			"active",
			"answered",
			"closed",
		).Default("active"),
		field.Int("church_id"),
		field.Int("contact_id").Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (PrayerRequest) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("church", Church.Type).
			Ref("prayer_requests").
			Field("church_id").
			Required().
			Unique(),
		edge.From("contact", Contact.Type).
			Ref("prayer_requests").
			Field("contact_id").
			Unique(),
	}
}

func (PrayerRequest) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("church_id", "status"),
		index.Fields("is_private"),
	}
}
