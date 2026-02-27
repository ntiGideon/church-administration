package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Relationship struct {
	ent.Schema
}

func (Relationship) Fields() []ent.Field {
	return []ent.Field{
		field.Int("from_contact_id"),
		field.Int("to_contact_id"),
		field.Enum("relation_type").Values(
			"wife", "husband",
			"mother", "father",
			"daughter", "son",
			"sister", "brother",
			"grandmother", "grandfather",
			"granddaughter", "grandson",
			"aunt", "uncle",
			"niece", "nephew",
			"cousin",
			"friend",
			"godmother", "godfather",
			"goddaughter", "godson",
			"other",
		),
		field.Text("notes").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (Relationship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("from_contact", Contact.Type).
			Ref("relationships_from").
			Field("from_contact_id").
			Unique().
			Required(),
		edge.From("to_contact", Contact.Type).
			Ref("relationships_to").
			Field("to_contact_id").
			Unique().
			Required(),
	}
}

func (Relationship) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("from_contact_id", "to_contact_id").Unique(),
	}
}
