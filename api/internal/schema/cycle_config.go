package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CycleConfig stores hierarchical Level1 weights G%/P% per cycle.
// G+P=100 enforced by DB check, fallback P=100 when no row.
type CycleConfig struct {
	ent.Schema
}

func (CycleConfig) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).StorageKey("id"),
		field.UUID("cycle_id", uuid.UUID{}),
		field.Float("g_weight").Min(0).Max(100).Default(0),
		field.Float("p_weight").Min(0).Max(100).Default(100),
	}
}

func (CycleConfig) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("cycle", Cycle.Type).Ref("cycle_config").Unique().Required().Field("cycle_id"),
	}
}

func (CycleConfig) Index() []ent.Index {
	return []ent.Index{
		index.Fields("cycle_id").Unique(),
	}
}
