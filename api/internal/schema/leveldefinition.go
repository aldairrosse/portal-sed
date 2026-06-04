package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// LevelDefinition holds the schema definition for global 1–5 scale definitions.
type LevelDefinition struct {
	ent.Schema
}

func (LevelDefinition) Fields() []ent.Field {
	return []ent.Field{
		field.Int("level").
			Unique().
			Range(1, 5).
			StorageKey("level"),
		field.String("label").
			NotEmpty(),
		field.Text("description").
			Optional(),
	}
}

func (LevelDefinition) Index() []ent.Index {
	return []ent.Index{
		index.Fields("level"),
	}
}
