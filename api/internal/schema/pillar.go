package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Pillar holds the schema definition for a competency pillar (corporate catalog).
type Pillar struct {
	ent.Schema
}

func (Pillar) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (Pillar) Fields() []ent.Field {
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

func (Pillar) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("competencies", Competency.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		edge.To("scale_criteria", ScaleCriterion.Type),
	}
}
