package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// PhaseTransition holds the schema definition for a phase transition rule.
type PhaseTransition struct {
	ent.Schema
}

func (PhaseTransition) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.Enum("from_phase").
			Values("asignacion", "avance", "cierre"),
		field.Enum("to_phase").
			Values("asignacion", "avance", "cierre"),
		field.Enum("trigger").
			Values("auto", "manual_rh"),
		field.JSON("conditions", map[string]interface{}{}).
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.UUID("cycle_id", uuid.UUID{}),
		field.UUID("from_phase_id", uuid.UUID{}),
		field.UUID("to_phase_id", uuid.UUID{}),
	}
}

func (PhaseTransition) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("cycle", Cycle.Type).
			Ref("phase_transitions").
			Unique().
			Required().
			Field("cycle_id"),
		edge.From("from_phase_def", PhaseDefinition.Type).
			Ref("outgoing_transitions").
			Unique().
			Required().
			Field("from_phase_id"),
		edge.From("to_phase_def", PhaseDefinition.Type).
			Ref("incoming_transitions").
			Unique().
			Required().
			Field("to_phase_id"),
	}
}

func (PhaseTransition) Index() []ent.Index {
	return []ent.Index{
		index.Fields("from_phase", "to_phase").
			Unique(),
	}
}
