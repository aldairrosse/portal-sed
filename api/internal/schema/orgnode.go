package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// OrgNode holds the schema definition for a node in the organizational tree.
type OrgNode struct {
	ent.Schema
}

func (OrgNode) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (OrgNode) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.String("name").
			NotEmpty(),
		field.Enum("type").
			Values("corporate", "retail"),
		field.String("code").
			NotEmpty(),
		field.JSON("metadata", map[string]interface{}{}).
			Optional(),
		field.UUID("organization_id", uuid.UUID{}),
		field.UUID("parent_id", uuid.UUID{}).
			Optional().
			Nillable(),
	}
}

func (OrgNode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("org_nodes").
			Unique().
			Required().
			Field("organization_id"),
		edge.To("parent", OrgNode.Type).
			Unique().
			Field("parent_id"),
		edge.From("children", OrgNode.Type).
			Ref("parent"),
		edge.To("employees", Employee.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (OrgNode) Index() []ent.Index {
	return []ent.Index{
		index.Fields("organization_id", "parent_id"),
		index.Fields("organization_id", "type"),
	}
}
