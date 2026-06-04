package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// NineBoxMatrix holds the schema definition for a 9x9 matrix per evaluator per cycle.
type NineBoxMatrix struct {
	ent.Schema
}

func (NineBoxMatrix) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (NineBoxMatrix) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.UUID("cycle_id", uuid.UUID{}),
		field.UUID("evaluator_id", uuid.UUID{}),
	}
}

func (NineBoxMatrix) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("cycle", Cycle.Type).
			Ref("nine_box_matrices").
			Unique().
			Required().
			Field("cycle_id"),
		edge.From("evaluator", Employee.Type).
			Ref("nine_box_matrices").
			Unique().
			Required().
			Field("evaluator_id"),
		edge.To("entries", NineBoxEntry.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (NineBoxMatrix) Index() []ent.Index {
	return []ent.Index{
		index.Fields("cycle_id", "evaluator_id").
			Unique(),
		index.Fields("evaluator_id"),
	}
}
