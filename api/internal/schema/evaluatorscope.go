package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// EvaluatorScope holds the schema definition for defining what a manager can see/edit.
type EvaluatorScope struct {
	ent.Schema
}

func (EvaluatorScope) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (EvaluatorScope) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.Enum("scope_type").
			Values("department", "team", "individual"),
		field.JSON("scope_data", map[string]interface{}{}).
			Optional(),
		field.UUID("evaluator_id", uuid.UUID{}),
		field.UUID("cycle_id", uuid.UUID{}).
			Optional().
			Nillable(),
	}
}

func (EvaluatorScope) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("evaluator", Employee.Type).
			Ref("evaluator_scopes").
			Unique().
			Required().
			Field("evaluator_id"),
		edge.From("cycle", Cycle.Type).
			Ref("evaluator_scopes").
			Unique().
			Field("cycle_id"),
	}
}

func (EvaluatorScope) Index() []ent.Index {
	return []ent.Index{
		index.Fields("evaluator_id", "cycle_id"),
	}
}
