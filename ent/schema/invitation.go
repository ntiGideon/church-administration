package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

type Invitation struct {
	ent.Schema
}

func (Invitation) Fields() []ent.Field {
	return []ent.Field{
		field.String("invitee_email"),
		field.String("invitee_name").Optional(),

		field.Enum("role").Values(
			"branch_admin",
			"secretary",
			"records_keeper",
			"finance_officer",
			"pastor",
			"member",
		),

		field.String("token").Unique(),
		field.Enum("status").Values(
			"pending",
			"accepted",
			"expired",
			"revoked",
		).Default("pending"),

		field.Int("custom_role_id").Optional(),

		field.Time("expires_at"),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Invitation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("church", Church.Type).
			Ref("invitations").
			Required().
			Unique(),

		edge.From("inviter", User.Type).
			Ref("sent_invitations").
			Unique(),

		edge.To("accepted_user", User.Type).
			Unique(),

		edge.From("custom_role", CustomRole.Type).
			Ref("invitations").
			Field("custom_role_id").
			Unique(),
	}
}
