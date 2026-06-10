package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Evaluation holds the schema definition for an employee evaluation in a cycle.
type Evaluation struct {
	ent.Schema
}

func (Evaluation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AuditMixin{},
		VersionMixin{},
	}
}

func (Evaluation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.Enum("phase").
			Values("asignacion", "avance", "cierre"),
		field.Enum("state").
			Values("pendiente_asignacion", "pendiente_avance", "pendiente_evaluacion_final", "completada"),
		field.Time("self_evaluation_completed_at").
			Optional().
			Nillable(),
		field.Time("rh_evaluation_completed_at").
			Optional().
			Nillable(),
		field.UUID("employee_id", uuid.UUID{}),
		field.UUID("cycle_id", uuid.UUID{}),
	}
}

func (Evaluation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("employee", Employee.Type).
			Ref("evaluations").
			Unique().
			Required().
			Field("employee_id"),
		edge.From("cycle", Cycle.Type).
			Ref("evaluations").
			Unique().
			Required().
			Field("cycle_id"),
		edge.To("competency_ratings", EvaluationCompetency.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		edge.To("goal_ratings", EvaluationGoal.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (Evaluation) Index() []ent.Index {
	return []ent.Index{
		index.Fields("employee_id", "cycle_id").
			Unique(),
		index.Fields("cycle_id", "state"),
		index.Fields("cycle_id", "phase"),
	}
}
