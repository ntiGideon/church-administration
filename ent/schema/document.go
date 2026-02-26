package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Document struct {
	ent.Schema
}

func (Document) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"),
		field.Text("description").Optional(),
		field.Enum("category").Values(
			"minutes",
			"bulletin",
			"constitution",
			"form",
			"report",
			"financial",
			"other",
		).Default("other"),
		field.String("file_url"),
		field.String("file_name"),
		field.Int64("file_size"),
		field.String("mime_type").Optional(),
		field.Bool("is_public").Default(true),
		field.Int("church_id"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Document) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("church", Church.Type).
			Ref("documents").
			Field("church_id").
			Required().
			Unique(),
	}
}

func (Document) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("church_id", "category"),
	}
}
