package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// GoalComment holds the schema definition for comments on goals.
// Table is managed via SQL migrations (see 000010 and 000044);
// this schema documents the phase column for future codegen.
type GoalComment struct {
	ent.Schema
}

func (GoalComment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (GoalComment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.UUID("goal_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Enum("phase").
			Values("asignacion", "avance", "cierre").
			Default("cierre"),
	}
}

func (GoalComment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("goal_id", "phase"),
	}
}
