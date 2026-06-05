// Package competency provides repository contracts and implementations
// for the competency framework catalog (pillars, competencies, scale criteria,
// level definitions, evaluation profiles, and acceptance levels).
package competency

import (
	"context"
	"time"

	"github.com/sed-evaluacion-desempeno/api/internal"
)

// TxFunc is a function that runs inside a transaction.
type TxFunc func(tx *internal.Tx) error

// PillarRepo defines CRUD operations for pillars.
type PillarRepo interface {
	// WithTx runs a function inside a database transaction.
	WithTx(ctx context.Context, fn TxFunc) error
	// List returns paginated pillars ordered by name. Cursor is a base64-encoded
	// JSON with the last item's name. includeCompetencies eager-loads competencies.
	List(ctx context.Context, cursor string, limit int, includeCompetencies bool) ([]*internal.Pillar, string, error)
	// Get returns a single pillar by ID, optionally with competencies eager-loaded.
	Get(ctx context.Context, id string, includeCompetencies bool) (*internal.Pillar, error)
	// Create inserts a new pillar.
	Create(ctx context.Context, name, description string) (*internal.Pillar, error)
	// Update updates a pillar with optimistic locking via ifMatch (updated_at).
	Update(ctx context.Context, id string, name, description string, ifMatch time.Time) (*internal.Pillar, error)
	// Delete removes a pillar by ID.
	Delete(ctx context.Context, id string) error
	// CountCompetencies returns the number of competencies in a pillar.
	CountCompetencies(ctx context.Context, pillarID string) (int, error)
}

// CompetencyRepo defines CRUD operations for competencies.
type CompetencyRepo interface {
	// WithTx runs a function inside a database transaction.
	WithTx(ctx context.Context, fn TxFunc) error
	// ListByPillar returns paginated competencies for a pillar, ordered by name.
	ListByPillar(ctx context.Context, pillarID string, cursor string, limit int) ([]*internal.Competency, string, error)
	// Get returns a single competency by ID.
	Get(ctx context.Context, id string) (*internal.Competency, error)
	// Create inserts a new competency in a pillar.
	Create(ctx context.Context, pillarID, name, description string) (*internal.Competency, error)
	// Update updates a competency with optional pillar move and optimistic locking.
	Update(ctx context.Context, id string, name, description, pillarID string, ifMatch time.Time) (*internal.Competency, error)
	// Delete removes a competency by ID.
	Delete(ctx context.Context, id string) error
	// CountScaleCriteria returns the number of scale criteria for a competency.
	CountScaleCriteria(ctx context.Context, competencyID string) (int, error)
}

// ScaleCriterionInput is a single criterion for bulk replace.
type ScaleCriterionInput struct {
	Level       int
	Description string
}

// ScaleRepo defines operations for scale criteria.
type ScaleRepo interface {
	// WithTx runs a function inside a database transaction.
	WithTx(ctx context.Context, fn TxFunc) error
	// GetByCompetency returns all scale criteria for a competency, ordered by level.
	GetByCompetency(ctx context.Context, competencyID string) ([]*internal.ScaleCriterion, error)
	// ReplaceAll deletes all existing criteria for a competency and inserts new ones.
	ReplaceAll(ctx context.Context, competencyID, pillarID string, criteria []ScaleCriterionInput) error
}

// CatalogRepo defines read-only operations for static catalogs.
type CatalogRepo interface {
	// ListLevels returns all level definitions ordered by level ASC.
	ListLevels(ctx context.Context) ([]*internal.LevelDefinition, error)
	// ListProfiles returns all evaluation profiles ordered by name ASC.
	ListProfiles(ctx context.Context) ([]*internal.EvaluationProfile, error)
}

// AcceptanceRepo defines operations for competency acceptance levels.
type AcceptanceRepo interface {
	// WithTx runs a function inside a database transaction.
	WithTx(ctx context.Context, fn TxFunc) error
	// List returns acceptance levels, optionally filtered by competency_id and/or profile_id.
	List(ctx context.Context, competencyID, profileID *string) ([]*internal.CompetencyAcceptanceLevel, error)
	// Upsert creates or updates an acceptance level on (competency_id, profile_id) unique key.
	Upsert(ctx context.Context, competencyID, profileID string, level int) (*internal.CompetencyAcceptanceLevel, error)
}
