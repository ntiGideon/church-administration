package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"time"
)

// Session holds the schema definition for the Session entity.
type Session struct {
	ent.Schema
}

// Fields of the Session.
func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.String("token").
			Unique().
			MaxLen(43).
			NotEmpty(),
		field.Bytes("data").
			NotEmpty(),
		field.Time("expiry").
			Default(time.Now).
			SchemaType(map[string]string{
				"postgres": "TIMESTAMP(6)",
			}),
	}
}

// Edges of the Session.
func (Session) Edges() []ent.Edge {
	return nil
}

// Indexes defines the indexes for the Session entity.
func (Session) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("expiry"),
	}
}
