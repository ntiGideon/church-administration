package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

type CustomRole struct {
	ent.Schema
}

func (CustomRole) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("description").Optional(),
		field.Text("permissions").Default("[]"),
		field.Bool("is_active").Default(true),
		field.Int("church_id"),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (CustomRole) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("church", Church.Type).
			Ref("custom_roles").
			Field("church_id").
			Required().
			Unique(),

		edge.To("users", User.Type).
			Annotations(entsql.Annotation{OnDelete: entsql.SetNull}),

		edge.To("invitations", Invitation.Type),
	}
}
