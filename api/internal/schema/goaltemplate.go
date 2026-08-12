package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// GoalTemplate holds the schema definition for a reusable goal template.
type GoalTemplate struct {
	ent.Schema
}

func (GoalTemplate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AuditMixin{},
	}
}

func (GoalTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.String("name").
			NotEmpty(),
		field.Text("description").
			Optional(),
		field.Enum("unit").
			Values("porcentaje", "moneda", "numero"),
		field.Enum("direction").
			Values("ascendente", "descendente").
			Default("ascendente"),
		field.Float("target_value").
			Positive(),
		field.Enum("goal_kind").
			Values("qualitative", "quantitative"),
		field.Bool("is_public").
			Default(false),
	}
}

func (GoalTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("creator", Employee.Type).
			Ref("goal_templates").
			Unique().
			Required().
			Field("created_by"),
		edge.To("kpi_links", GoalTemplateKpiLink.Type),
	}
}

func (GoalTemplate) Index() []ent.Index {
	return []ent.Index{
		index.Fields("created_by"),
		index.Fields("is_public"),
	}
}
