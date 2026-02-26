package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Sermon struct {
	ent.Schema
}

func (Sermon) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"),
		field.String("speaker"),
		field.String("series").Optional(),
		field.String("scripture").Optional(),
		field.Text("description").Optional(),
		field.String("media_url").Optional(),
		field.Time("service_date"),
		field.Bool("is_published").Default(false),
		field.Int("church_id"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Sermon) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("church", Church.Type).
			Ref("sermons").
			Field("church_id").
			Required().
			Unique(),
	}
}

func (Sermon) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("church_id", "service_date"),
		index.Fields("is_published"),
	}
}
