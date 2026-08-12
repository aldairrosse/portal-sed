package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// GoalTemplateKpiLink holds the schema for linking templates to KPIs.
type GoalTemplateKpiLink struct {
	ent.Schema
}

func (GoalTemplateKpiLink) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (GoalTemplateKpiLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.UUID("template_id", uuid.UUID{}),
		field.UUID("kpi_id", uuid.UUID{}),
	}
}

func (GoalTemplateKpiLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("template", GoalTemplate.Type).
			Ref("kpi_links").
			Unique().
			Required().
			Field("template_id"),
		edge.From("kpi", KPI.Type).
			Ref("template_links").
			Unique().
			Required().
			Field("kpi_id"),
	}
}

func (GoalTemplateKpiLink) Index() []ent.Index {
	return []ent.Index{
		index.Fields("template_id", "kpi_id").Unique(),
	}
}
