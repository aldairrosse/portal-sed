package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Goal holds the schema definition for a goal within a category.
type Goal struct {
	ent.Schema
}

func (Goal) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AuditMixin{},
	}
}

func (Goal) Fields() []ent.Field {
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
		field.Float("weight").
			Range(0, 100),
		field.Float("target_value").
			Positive(),
		field.Float("current_value").
			Default(0),
		field.Enum("state").
			Values("borrador", "fijada", "en_seguimiento", "evaluada", "cerrada"),
		field.UUID("category_id", uuid.UUID{}),
	}
}

func (Goal) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("category", GoalCategory.Type).
			Ref("goals").
			Unique().
			Required().
			Field("category_id"),
		edge.To("kpi_links", GoalKpiLink.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		edge.To("evaluation_goals", EvaluationGoal.Type),
	}
}

func (Goal) Index() []ent.Index {
	return []ent.Index{
		index.Fields("category_id"),
		index.Fields("state"),
		index.Fields("created_by"),
	}
}
