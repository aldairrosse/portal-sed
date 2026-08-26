package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// GlobalGoalRule holds the schema for mass assignment rules.
type GlobalGoalRule struct {
	ent.Schema
}

func (GlobalGoalRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (GlobalGoalRule) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.UUID("goal_id", uuid.UUID{}),
		field.Enum("rule_type").
			Values("department", "min_direct_reports", "role"),
		field.UUID("department_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("profile_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Int("min_direct_reports").
			Optional().
			Nillable(),
		field.Float("default_weight").
			Min(0).
			Max(100),
		field.Float("default_target").
			Default(100).
			Min(0),
	}
}

func (GlobalGoalRule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("goal", Goal.Type).
			Ref("global_rules").
			Unique().
			Required().
			Field("goal_id"),
		edge.From("department", OrgNode.Type).
			Ref("global_goal_rules").
			Unique().
			Field("department_id"),
		edge.From("profile", EvaluationProfile.Type).
			Ref("global_goal_rules").
			Unique().
			Field("profile_id"),
	}
}

// Index returns no unique index on (goal_id, rule_type): a goal may hold
// multiple rules of the same type (e.g. one per department).
func (GlobalGoalRule) Index() []ent.Index {
	return nil
}
