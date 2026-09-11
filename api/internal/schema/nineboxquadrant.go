package schema

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// NineBoxQuadrant holds the schema definition for a quadrant in the 9x9 matrix catalog.
type NineBoxQuadrant struct {
	ent.Schema
}

func (NineBoxQuadrant) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.Int("quadrant").
			Unique().
			Range(1, 9),
		field.String("label").
			NotEmpty(),
		field.Text("description").
			Optional(),
		field.String("color").
			Optional(), // was NotEmpty — changed to Optional for backward compat (now replaced by color_hex)
		field.Text("action_recommendation").
			Optional(),
		field.String("title").
			Optional().
			MaxLen(100),
		field.String("color_hex").
			Optional().
			Match(regexp.MustCompile("^#[0-9A-Fa-f]{6}$")),
	}
}

func (NineBoxQuadrant) Index() []ent.Index {
	return []ent.Index{
		index.Fields("quadrant"),
	}
}
