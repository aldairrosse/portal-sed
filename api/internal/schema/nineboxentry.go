package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// NineBoxEntry holds the schema definition for a single evaluatee rating in a 9x9 matrix.
type NineBoxEntry struct {
	ent.Schema
}

func (NineBoxEntry) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AuditMixin{},
		VersionMixin{},
	}
}

func (NineBoxEntry) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.Int("performance_score").
			Range(1, 9),
		field.Int("potential_score").
			Range(1, 9),
		field.Int("quadrant").
			Range(1, 9),
		field.Text("comments").
			Optional(),
		field.UUID("matrix_id", uuid.UUID{}),
		field.UUID("evaluatee_id", uuid.UUID{}),
	}
}

func (NineBoxEntry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("matrix", NineBoxMatrix.Type).
			Ref("entries").
			Unique().
			Required().
			Field("matrix_id"),
		edge.From("evaluatee", Employee.Type).
			Ref("nine_box_entries").
			Unique().
			Required().
			Field("evaluatee_id"),
	}
}

func (NineBoxEntry) Index() []ent.Index {
	return []ent.Index{
		index.Fields("matrix_id", "evaluatee_id").
			Unique(),
		index.Fields("evaluatee_id"),
	}
}
