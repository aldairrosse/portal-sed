package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// EvaluationGoal holds the schema definition for a goal rating within an evaluation.
type EvaluationGoal struct {
	ent.Schema
}

func (EvaluationGoal) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (EvaluationGoal) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.Int("final_rating").
			Range(1, 5).
			Optional().
			Nillable(),
		field.Text("final_comments").
			Optional(),
		field.UUID("evaluation_id", uuid.UUID{}),
		field.UUID("goal_id", uuid.UUID{}),
	}
}

func (EvaluationGoal) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("evaluation", Evaluation.Type).
			Ref("goal_ratings").
			Unique().
			Required().
			Field("evaluation_id"),
		edge.From("goal", Goal.Type).
			Ref("evaluation_goals").
			Unique().
			Required().
			Field("goal_id"),
	}
}

func (EvaluationGoal) Index() []ent.Index {
	return []ent.Index{
		index.Fields("evaluation_id", "goal_id").
			Unique(),
	}
}
