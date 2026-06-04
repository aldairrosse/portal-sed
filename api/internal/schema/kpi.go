package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// KPI holds the schema definition for a reusable key performance indicator.
type KPI struct {
	ent.Schema
}

func (KPI) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (KPI) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.String("name").
			Unique().
			NotEmpty(),
		field.Enum("unit").
			Values("porcentaje", "moneda", "numero"),
		field.Text("description").
			Optional(),
	}
}

func (KPI) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("goal_links", GoalKpiLink.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (KPI) Index() []ent.Index {
	return []ent.Index{
		index.Fields("name").
			Unique(),
	}
}
