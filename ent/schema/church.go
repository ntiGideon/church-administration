package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"time"
)

type Church struct {
	ent.Schema
}

func (Church) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("address"),
		field.String("city").Default("Kumasi"),
		field.String("state").Optional(),
		field.String("zip_code").Optional(),
		field.String("country").Default("GHANA"),
		field.String("website").Optional(),
		field.String("phone").Optional(),
		field.String("registration_token").
			Unique().
			Optional().
			Nillable(),
		field.String("email"),
		field.Time("established_date").Optional(),

		field.Enum("type").Values(
			"headquarters",
			"branch",
			"mission",
		).Default("branch"),

		field.Int("parent_id").Optional(),

		field.Float("latitude").Optional(),
		field.Float("longitude").Optional(),
		field.String("logo_url").Optional(),
		field.String("banner_url").Optional(),

		field.Int("congregation_size").Default(0),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Church) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("parent", Church.Type).
			Ref("children").
			Field("parent_id").
			Unique(),

		edge.To("children", Church.Type),

		edge.To("users", User.Type),
		edge.To("departments", Department.Type),
		edge.To("events", Event.Type),
		edge.To("finances", Finance.Type),
		edge.To("invitations", Invitation.Type),
		edge.To("announcements", Announcement.Type),
	}
}

func (Church) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parent_id"),
		index.Fields("type"),
		index.Fields("is_active"),
	}
}
