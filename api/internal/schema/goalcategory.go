package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// GoalCategory holds the schema definition for a goal category defined by an employee.
type GoalCategory struct {
	ent.Schema
}

func (GoalCategory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AuditMixin{},
	}
}

func (GoalCategory) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.String("name").
			NotEmpty(),
		field.Text("description").
			Optional(),
		field.Float("weight").
			Range(0, 100),
		field.UUID("employee_id", uuid.UUID{}),
	}
}

func (GoalCategory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("employee", Employee.Type).
			Ref("goal_categories").
			Unique().
			Required().
			Field("employee_id"),
		edge.To("goals", Goal.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (GoalCategory) Index() []ent.Index {
	return []ent.Index{
		index.Fields("employee_id"),
		index.Fields("employee_id", "name").
			Unique(),
	}
}
