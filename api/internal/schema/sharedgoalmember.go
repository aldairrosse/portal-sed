package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// SharedGoalMember holds the schema for a member of a shared goal group.
type SharedGoalMember struct {
	ent.Schema
}

func (SharedGoalMember) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (SharedGoalMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.UUID("group_id", uuid.UUID{}),
		field.UUID("employee_id", uuid.UUID{}),
		field.Float("weight").
			Range(0, 100),
		field.Float("target_value").
			Positive(),
		field.Float("baseline_value").
			Optional().
			Nillable(),
	}
}

func (SharedGoalMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", SharedGoalGroup.Type).
			Ref("members").
			Unique().
			Required().
			Field("group_id"),
		edge.From("employee", Employee.Type).
			Ref("shared_goal_members").
			Unique().
			Required().
			Field("employee_id"),
	}
}

func (SharedGoalMember) Index() []ent.Index {
	return []ent.Index{
		index.Fields("group_id", "employee_id").Unique(),
		index.Fields("employee_id"),
	}
}
