package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// GoalProgressLog holds the schema definition for audit rows recording
// every goal progress change (previous_value -> new_value).
type GoalProgressLog struct {
	ent.Schema
}

func (GoalProgressLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AuditMixin{},
	}
}

func (GoalProgressLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.UUID("goal_id", uuid.UUID{}),
		field.UUID("employee_id", uuid.UUID{}),
		field.Float("previous_value"),
		field.Float("new_value"),
	}
}

func (GoalProgressLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("goal", Goal.Type).
			Ref("progress_logs").
			Unique().
			Required().
			Field("goal_id"),
	}
}

func (GoalProgressLog) Index() []ent.Index {
	return []ent.Index{
		index.Fields("goal_id"),
		index.Fields("employee_id"),
	}
}
