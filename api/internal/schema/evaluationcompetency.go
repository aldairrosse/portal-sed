package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// EvaluationCompetency holds the schema definition for a competency rating within an evaluation.
type EvaluationCompetency struct {
	ent.Schema
}

func (EvaluationCompetency) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (EvaluationCompetency) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.Int("rating").
			Range(1, 5),
		field.Text("comments").
			Optional(),
		field.UUID("evaluation_id", uuid.UUID{}),
		field.UUID("competency_id", uuid.UUID{}),
		field.UUID("profile_id", uuid.UUID{}),
	}
}

func (EvaluationCompetency) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("evaluation", Evaluation.Type).
			Ref("competency_ratings").
			Unique().
			Required().
			Field("evaluation_id"),
		edge.From("competency", Competency.Type).
			Ref("evaluation_competencies").
			Unique().
			Required().
			Field("competency_id"),
		edge.From("profile", EvaluationProfile.Type).
			Ref("evaluation_competencies").
			Unique().
			Required().
			Field("profile_id"),
	}
}

func (EvaluationCompetency) Index() []ent.Index {
	return []ent.Index{
		index.Fields("evaluation_id", "competency_id").
			Unique(),
	}
}
