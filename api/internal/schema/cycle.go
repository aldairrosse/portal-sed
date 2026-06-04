package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Cycle holds the schema definition for an annual evaluation cycle.
type Cycle struct {
	ent.Schema
}

func (Cycle) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (Cycle) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.Int("year"),
		field.Enum("current_phase").
			Values("asignacion", "avance", "cierre"),
		field.Time("started_at").
			Optional().
			Nillable(),
		field.Time("finished_at").
			Optional().
			Nillable(),
		field.UUID("organization_id", uuid.UUID{}),
	}
}

func (Cycle) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("cycles").
			Unique().
			Required().
			Field("organization_id"),
		edge.To("phase_transitions", PhaseTransition.Type),
		edge.To("phase_definitions", PhaseDefinition.Type),
		edge.To("evaluator_scopes", EvaluatorScope.Type),
	}
}

func (Cycle) Index() []ent.Index {
	return []ent.Index{
		index.Fields("organization_id", "year").
			Unique(),
		index.Fields("current_phase"),
	}
}
