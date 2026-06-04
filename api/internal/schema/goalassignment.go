package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// GoalAssignment holds the schema definition for an employee's goal assignment in a cycle.
type GoalAssignment struct {
	ent.Schema
}

func (GoalAssignment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (GoalAssignment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.UUID("employee_id", uuid.UUID{}),
		field.UUID("cycle_id", uuid.UUID{}),
	}
}

func (GoalAssignment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("employee", Employee.Type).
			Ref("goal_assignments").
			Unique().
			Required().
			Field("employee_id"),
		edge.From("cycle", Cycle.Type).
			Ref("goal_assignments").
			Unique().
			Required().
			Field("cycle_id"),
	}
}

func (GoalAssignment) Index() []ent.Index {
	return []ent.Index{
		index.Fields("employee_id", "cycle_id").
			Unique(),
	}
}
