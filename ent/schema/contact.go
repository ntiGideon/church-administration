package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

type Contact struct {
	ent.Schema
}

func (Contact) Fields() []ent.Field {
	return []ent.Field{
		field.String("first_name"),
		field.String("last_name"),
		field.String("middle_name").Optional(),

		field.String("phone").Optional(),
		field.String("secondary_phone").Optional(),

		field.String("email").Optional(),

		field.Enum("gender").Values("male", "female", "other").Optional(),

		field.Time("date_of_birth").Optional(),

		field.Enum("marital_status").Values(
			"single",
			"married",
			"divorced",
			"widowed",
		).Optional(),

		field.String("occupation").Optional(),

		field.String("address_line1").Optional(),
		field.String("address_line2").Optional(),
		field.String("city").Optional(),
		field.String("state").Optional(),
		field.String("country").Optional(),
		field.String("postal_code").Optional(),

		field.String("emergency_contact_name").Optional(),
		field.String("emergency_contact_phone").Optional(),
		field.String("emergency_contact_relationship").Optional(),

		field.String("profile_picture_url").Optional(),

		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Contact) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique(),
	}
}
