package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CompetencyAcceptanceLevel holds the schema definition for the minimum
// acceptable level per competency × profile.
type CompetencyAcceptanceLevel struct {
	ent.Schema
}

func (CompetencyAcceptanceLevel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (CompetencyAcceptanceLevel) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.Int("level").
			Range(1, 5),
		field.UUID("competency_id", uuid.UUID{}),
		field.UUID("profile_id", uuid.UUID{}),
	}
}

func (CompetencyAcceptanceLevel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("competency", Competency.Type).
			Ref("acceptance_levels").
			Unique().
			Required().
			Field("competency_id"),
		edge.From("profile", EvaluationProfile.Type).
			Ref("acceptance_levels").
			Unique().
			Required().
			Field("profile_id"),
	}
}

func (CompetencyAcceptanceLevel) Index() []ent.Index {
	return []ent.Index{
		index.Fields("competency_id", "profile_id").
			Unique(),
	}
}
