package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// NineBoxScale holds the schema definition for the 9 levels per axis in the 9x9 matrix.
type NineBoxScale struct {
	ent.Schema
}

func (NineBoxScale) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.Enum("axis").
			Values("performance", "potential"),
		field.Int("level").
			Range(1, 9),
		field.String("label").
			NotEmpty(),
		field.Text("description").
			Optional(),
	}
}

func (NineBoxScale) Index() []ent.Index {
	return []ent.Index{
		index.Fields("axis", "level").
			Unique(),
	}
}
