package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// SharedGoalGroup holds the schema for a group of shared goals.
type SharedGoalGroup struct {
	ent.Schema
}

func (SharedGoalGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AuditMixin{},
	}
}

func (SharedGoalGroup) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.UUID("goal_id", uuid.UUID{}),
		field.String("name").
			NotEmpty(),
		field.Text("description").
			Optional(),
	}
}

func (SharedGoalGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("goal", Goal.Type).
			Ref("shared_group").
			Unique().
			Required().
			Field("goal_id"),
		edge.From("creator", Employee.Type).
			Ref("shared_goal_groups").
			Unique().
			Required().
			Field("created_by"),
		edge.To("members", SharedGoalMember.Type).
		Annotations(entsql.Annotation{
			OnDelete: entsql.Cascade,
		}),
	}
}

func (SharedGoalGroup) Index() []ent.Index {
	return []ent.Index{
		index.Fields("goal_id").Unique(),
		index.Fields("created_by"),
	}
}
