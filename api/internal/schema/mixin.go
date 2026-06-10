package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// TimeMixin adds created_at and updated_at fields to schemas.
type TimeMixin struct {
	mixin.Schema
}

func (TimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// AuditMixin adds created_at, updated_at, created_by, and updated_by fields.
type AuditMixin struct {
	mixin.Schema
}

func (AuditMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.UUID("created_by", uuid.UUID{}),
		field.UUID("updated_by", uuid.UUID{}),
	}
}

// VersionMixin adds an integer version field for optimistic locking.
type VersionMixin struct {
	mixin.Schema
}

func (VersionMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Int("version").
			Default(1).
			Positive(),
	}
}
