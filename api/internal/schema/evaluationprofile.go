package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// EvaluationProfile holds the schema definition for an evaluation profile (e.g., colaborador, jefe, rh).
type EvaluationProfile struct {
	ent.Schema
}

func (EvaluationProfile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.String("name").
			Unique().
			NotEmpty(),
		field.Text("description").
			Optional(),
	}
}

func (EvaluationProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("employees", Employee.Type),
		edge.To("acceptance_levels", CompetencyAcceptanceLevel.Type),
	}
}
