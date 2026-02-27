package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"time"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").Unique(),
		field.String("password_hash"),

		field.Enum("role").Values(
			"super_admin",
			"branch_admin",
			"secretary",
			"records_keeper",
			"finance_officer",
			"pastor",
			"member",
		).Default("member"),

		field.Bool("is_active").Default(true),
		field.Time("last_login").Optional(),

		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		// A user belongs to exactly one church
		edge.From("church", Church.Type).
			Ref("users").
			Unique(),

		// New: A user has one contact
		edge.From("contact", Contact.Type).
			Ref("user").
			Required().
			Unique(),
		edge.To("finance_records", Finance.Type),

		edge.To("sent_invitations", Invitation.Type),
		edge.To("announcements", Announcement.Type),

		edge.From("accepted_invitation", Invitation.Type).
			Ref("accepted_user").
			Unique(),
		edge.To("pastoral_notes_recorded", PastoralNote.Type),
		edge.To("sent_communications", Communication.Type),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email").Unique(),
		index.Fields("role"),
	}
}
