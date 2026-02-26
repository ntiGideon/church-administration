package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Visitor struct {
	ent.Schema
}

func (Visitor) Fields() []ent.Field {
	return []ent.Field{
		field.String("first_name"),
		field.String("last_name"),
		field.String("email").Optional(),
		field.String("phone").Optional(),
		field.String("address").Optional(),
		field.Time("visit_date"),
		field.Enum("how_heard").Values(
			"walk_in",
			"invited_by_member",
			"social_media",
			"website",
			"flyer",
			"other",
		).Optional(),
		field.String("invited_by").Optional(),
		field.Text("notes").Optional(),
		field.Enum("follow_up_status").Values(
			"new",
			"contacted",
			"follow_up_scheduled",
			"follow_up_done",
			"converted",
			"no_response",
		).Default("new"),
		field.Int("church_id"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Visitor) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("church", Church.Type).
			Ref("visitors").
			Field("church_id").
			Required().
			Unique(),
	}
}

func (Visitor) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("church_id", "visit_date"),
		index.Fields("follow_up_status"),
	}
}
