package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// PhaseDefinition holds the schema definition for a cycle phase configuration.
type PhaseDefinition struct {
	ent.Schema
}

func (PhaseDefinition) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (PhaseDefinition) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.Enum("phase").
			Values("asignacion", "avance", "cierre"),
		field.String("label").
			NotEmpty(),
		field.Int("order").
			Range(1, 3),
		field.JSON("allowed_actors", []string{}).
			Optional(),
		field.JSON("allowed_actions", []string{}).
			Optional(),
		field.JSON("blocked_actions", []string{}).
			Optional(),
		field.UUID("cycle_id", uuid.UUID{}),
	}
}

func (PhaseDefinition) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("cycle", Cycle.Type).
			Ref("phase_definitions").
			Unique().
			Required().
			Field("cycle_id"),
		edge.To("outgoing_transitions", PhaseTransition.Type),
		edge.To("incoming_transitions", PhaseTransition.Type),
	}
}
