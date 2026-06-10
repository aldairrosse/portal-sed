package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ScaleCriterion holds the schema definition for a descriptive criterion
// for a specific competency × level.
type ScaleCriterion struct {
	ent.Schema
}

func (ScaleCriterion) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
		VersionMixin{},
	}
}

func (ScaleCriterion) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.Int("level").
			Range(1, 5),
		field.Text("description").
			NotEmpty(),
		field.UUID("competency_id", uuid.UUID{}),
		field.UUID("pillar_id", uuid.UUID{}),
	}
}

func (ScaleCriterion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("competency", Competency.Type).
			Ref("scale_criteria").
			Unique().
			Required().
			Field("competency_id"),
		edge.From("pillar", Pillar.Type).
			Ref("scale_criteria").
			Unique().
			Required().
			Field("pillar_id"),
	}
}

func (ScaleCriterion) Index() []ent.Index {
	return []ent.Index{
		index.Fields("competency_id", "pillar_id", "level"),
	}
}
