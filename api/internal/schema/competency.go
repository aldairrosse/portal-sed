package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Competency holds the schema definition for a competency within a pillar.
type Competency struct {
	ent.Schema
}

func (Competency) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (Competency) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.String("name").
			NotEmpty(),
		field.Text("description").
			Optional(),
		field.UUID("pillar_id", uuid.UUID{}),
	}
}

func (Competency) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("pillar", Pillar.Type).
			Ref("competencies").
			Unique().
			Required().
			Field("pillar_id"),
		edge.To("scale_criteria", ScaleCriterion.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		edge.To("acceptance_levels", CompetencyAcceptanceLevel.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		edge.To("evaluation_competencies", EvaluationCompetency.Type),
	}
}

func (Competency) Index() []ent.Index {
	return []ent.Index{
		index.Fields("pillar_id"),
	}
}
