package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Employee holds the schema definition for an employee in the org chart.
type Employee struct {
	ent.Schema
}

func (Employee) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AuditMixin{},
	}
}

func (Employee) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.String("first_name").
			NotEmpty(),
		field.String("last_name").
			NotEmpty(),
		field.String("employee_number").
			NotEmpty(),
		field.String("email").
			Unique().
			NotEmpty(),
		field.Bool("is_active").
			Default(true),
		field.UUID("org_node_id", uuid.UUID{}),
		field.UUID("manager_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("profile_id", uuid.UUID{}),
	}
}

func (Employee) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("org_node", OrgNode.Type).
			Ref("employees").
			Unique().
			Required().
			Field("org_node_id"),
		edge.To("manager", Employee.Type).
			Unique().
			Field("manager_id"),
		edge.From("direct_reports", Employee.Type).
			Ref("manager"),
		edge.From("profile", EvaluationProfile.Type).
			Ref("employees").
			Unique().
			Required().
			Field("profile_id"),
		edge.To("evaluator_scopes", EvaluatorScope.Type),
		edge.To("goal_categories", GoalCategory.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		edge.To("goal_assignments", GoalAssignment.Type),
		edge.To("evaluations", Evaluation.Type),
		edge.To("nine_box_matrices", NineBoxMatrix.Type),
		edge.To("nine_box_entries", NineBoxEntry.Type),
	}
}

func (Employee) Index() []ent.Index {
	return []ent.Index{
		index.Fields("org_node_id", "manager_id"),
		index.Fields("profile_id"),
		index.Fields("manager_id", "is_active"),
	}
}
