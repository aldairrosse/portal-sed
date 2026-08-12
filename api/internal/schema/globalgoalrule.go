package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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
			Values("department", "min_direct_reports"),
		field.UUID("department_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Int("min_direct_reports").
			Optional().
			Nillable(),
		field.Float("default_weight").
			Range(0, 100),
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
	}
}

func (GlobalGoalRule) Index() []ent.Index {
	return []ent.Index{
		index.Fields("goal_id", "rule_type").Unique(),
	}
}
