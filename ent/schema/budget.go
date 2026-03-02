package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Budget holds the schema definition for a church budget plan.
type Budget struct {
	ent.Schema
}

func (Budget) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.Int("fiscal_year"),
		field.Enum("period").
			Values("annual", "quarterly", "monthly").
			Default("annual"),
		field.Time("start_date"),
		field.Time("end_date"),
		field.Enum("status").
			Values("draft", "active", "closed").
			Default("draft"),
		field.String("notes").Optional(),
		field.Int("church_id"),
		field.Time("created_at").Default(time.Now),
	}
}

func (Budget) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("church", Church.Type).
			Ref("budgets").
			Field("church_id").
			Required().
			Unique(),
		edge.To("lines", BudgetLine.Type),
	}
}
