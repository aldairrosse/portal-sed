package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// TeamWeightConfig stores hierarchical Level2 weights J%/PJ% per team/cycle.
// J+PJ=100 enforced by DB check, fallback PJ=100 when no row.
type TeamWeightConfig struct {
	ent.Schema
}

func (TeamWeightConfig) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).StorageKey("id"),
		field.UUID("cycle_id", uuid.UUID{}),
		field.UUID("team_id", uuid.UUID{}),
		field.Float("j_weight").Min(0).Max(100).Default(0),
		field.Float("pj_weight").Min(0).Max(100).Default(100),
	}
}

func (TeamWeightConfig) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("cycle", Cycle.Type).Ref("team_weight_configs").Unique().Required().Field("cycle_id"),
		edge.From("team_node", OrgNode.Type).Ref("team_weight_configs").Unique().Required().Field("team_id"),
	}
}

func (TeamWeightConfig) Index() []ent.Index {
	return []ent.Index{
		index.Fields("cycle_id", "team_id").Unique(),
	}
}
