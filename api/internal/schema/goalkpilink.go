package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// GoalKpiLink holds the schema definition for the N:M relationship between goals and KPIs.
type GoalKpiLink struct {
	ent.Schema
}

func (GoalKpiLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("goal_id", uuid.UUID{}),
		field.UUID("kpi_id", uuid.UUID{}),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (GoalKpiLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("goal", Goal.Type).
			Ref("kpi_links").
			Unique().
			Required().
			Field("goal_id"),
		edge.From("kpi", KPI.Type).
			Ref("goal_links").
			Unique().
			Required().
			Field("kpi_id"),
	}
}

func (GoalKpiLink) Index() []ent.Index {
	return []ent.Index{
		index.Fields("goal_id", "kpi_id").
			Unique(),
		index.Fields("kpi_id"),
	}
}
