package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Group holds the schema definition for a church small group or cell group.
type Group struct {
	ent.Schema
}

func (Group) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("description").Optional(),
		field.Enum("group_type").Values(
			"cell_group",
			"bible_study",
			"youth",
			"women",
			"men",
			"choir",
			"committee",
			"prayer_team",
			"other",
		).Default("cell_group"),
		field.Enum("meeting_day").Values(
			"monday", "tuesday", "wednesday", "thursday",
			"friday", "saturday", "sunday",
		).Optional(),
		field.String("meeting_time").Optional(),
		field.String("location").Optional(),
		field.Bool("is_active").Default(true),
		field.Int("church_id"),
		field.Int("leader_id").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Group) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("church", Church.Type).
			Ref("groups").
			Field("church_id").
			Required().
			Unique(),
		edge.To("leader", Contact.Type).
			Field("leader_id").
			Unique(),
		edge.To("members", Contact.Type),
	}
}
