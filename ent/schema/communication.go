package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Communication struct {
	ent.Schema
}

func (Communication) Fields() []ent.Field {
	return []ent.Field{
		field.String("subject"),
		field.Text("body_html"),
		// "all" | "with_email" | "gender:male" | "gender:female" | "group:{id}"
		field.String("recipient_filter").Default("all"),
		// Human-readable description, e.g. "All Members", "Male Members", "Cell Group: Youth"
		field.String("recipient_filter_label").Default("All Members"),
		field.Int("recipient_count").Default(0),
		field.Int("sent_count").Default(0),
		field.Int("fail_count").Default(0),
		field.Time("sent_at").Default(time.Now),
	}
}

func (Communication) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("church", Church.Type).
			Ref("communications").
			Unique().
			Required(),
		edge.From("sender", User.Type).
			Ref("sent_communications").
			Unique(),
	}
}
