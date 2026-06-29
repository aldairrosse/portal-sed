package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ActivityLog holds the schema definition for activity log entries.
// Each entry records one action performed by an employee — append-only.
type ActivityLog struct {
	ent.Schema
}

func (ActivityLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (ActivityLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.UUID("employee_id", uuid.UUID{}),
		field.String("action"),
		field.String("description"),
		field.String("module"),
		field.JSON("metadata", map[string]interface{}{}).
			Optional(),
	}
}

func (ActivityLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("employee", Employee.Type).
			Ref("activity_logs").
			Unique().
			Required().
			Field("employee_id"),
	}
}

func (ActivityLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("employee_id", "created_at"),
	}
}

func (ActivityLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "activity_logs"},
	}
}
