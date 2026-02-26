package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Finance holds the schema definition for the Finance entity.
type Finance struct {
	ent.Schema
}

// Fields of the Finance.
func (Finance) Fields() []ent.Field {
	return []ent.Field{
		field.String("description"),
		field.Enum("transaction_type").Values(
			"donation",
			"tithe",
			"offering",
			"expense",
			"salary",
			"other",
		),
		field.Float("amount"),
		field.String("currency").Default("USD"),
		field.Time("transaction_date"),
		field.String("category"),
		field.String("payment_method").Optional(),
		field.String("reference_number").Optional(),
		field.String("notes").Optional(),
		field.Int("contact_id").Optional(),
	}
}

// Edges of the Finance.
func (Finance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("recorded_by", User.Type).
			Ref("finance_records").
			Unique(),

		edge.From("church", Church.Type).
			Ref("finances").
			Unique(),

		edge.To("donor", Contact.Type).
			Field("contact_id").
			Unique(),
	}
}
