package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// GlobalGoalAssignment holds the schema for assigning a global goal to an employee.
type GlobalGoalAssignment struct {
	ent.Schema
}

func (GlobalGoalAssignment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (GlobalGoalAssignment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.UUID("goal_id", uuid.UUID{}),
		field.UUID("employee_id", uuid.UUID{}),
		field.Float("weight").
			Min(0).
			Max(100),
		field.Float("target_value").
			Min(0),
		field.Float("baseline_value").
			Optional().
			Nillable(),
	}
}

func (GlobalGoalAssignment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("goal", Goal.Type).
			Ref("global_assignments").
			Unique().
			Required().
			Field("goal_id"),
		edge.From("employee", Employee.Type).
			Ref("global_goal_assignments").
			Unique().
			Required().
			Field("employee_id"),
	}
}

func (GlobalGoalAssignment) Index() []ent.Index {
	return []ent.Index{
		index.Fields("goal_id", "employee_id").Unique(),
		index.Fields("employee_id"),
	}
}
