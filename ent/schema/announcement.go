package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

type Announcement struct {
	ent.Schema
}

func (Announcement) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"),
		field.Text("content"),
		field.Enum("category").Values(
			"general",
			"prayer",
			"event",
			"urgent",
			"financial",
		).Default("general"),
		field.Bool("is_published").Default(false),
		field.Time("published_at").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Announcement) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("church", Church.Type).
			Ref("announcements").
			Required().
			Unique(),
		edge.From("author", User.Type).
			Ref("announcements").
			Unique(),
	}
}
