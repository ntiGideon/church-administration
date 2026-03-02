package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// BudgetLine holds the schema definition for a single line item within a budget.
type BudgetLine struct {
	ent.Schema
}

func (BudgetLine) Fields() []ent.Field {
	return []ent.Field{
		field.String("category"),
		field.Enum("line_type").
			Values("income", "expense").
			Default("income"),
		field.Float("allocated_amount"),
		field.String("currency").Default("GHS"),
		field.String("notes").Optional(),
		field.Int("budget_id"),
	}
}

func (BudgetLine) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("budget", Budget.Type).
			Ref("lines").
			Field("budget_id").
			Required().
			Unique(),
	}
}
