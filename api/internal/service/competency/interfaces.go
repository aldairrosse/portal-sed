// Package competency provides service-layer business rules for the competency
// framework catalog — pillars, competencies, scale criteria, level definitions,
// evaluation profiles, and acceptance levels.
package competency

import (
	"context"
	"time"

	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/competency"
)

// ListOptions holds pagination and filter options for list operations.
type ListOptions struct {
	Cursor  string
	Limit   int
	Include []string
}

// ListResult holds a paginated result set.
type ListResult[T any] struct {
	Data       []T
	NextCursor string
	HasMore    bool
}

// AcceptanceFilter holds optional filters for listing acceptance levels.
type AcceptanceFilter struct {
	ProfileID    *string
	CompetencyID *string
}

// PillarService defines business rules for pillars.
type PillarService interface {
	List(ctx context.Context, opts ListOptions) (*ListResult[dto.PillarListItem], error)
	Get(ctx context.Context, id string, includeCompetencies bool) (*dto.PillarDetail, error)
	Create(ctx context.Context, req dto.CreatePillarRequest) (*dto.PillarDetail, error)
	Update(ctx context.Context, id string, req dto.UpdatePillarRequest, ifMatch time.Time) (*dto.PillarDetail, error)
	Delete(ctx context.Context, id string, force bool) error
}

// CompetencyService defines business rules for competencies.
type CompetencyService interface {
	ListByPillar(ctx context.Context, pillarID string, opts ListOptions) (*ListResult[dto.CompetencyLite], error)
	Get(ctx context.Context, id string) (*dto.CompetencyDetail, error)
	Create(ctx context.Context, pillarID string, req dto.CreateCompetencyRequest) (*dto.CompetencyDetail, error)
	Update(ctx context.Context, id string, req dto.UpdateCompetencyRequest, ifMatch time.Time) (*dto.CompetencyDetail, error)
	Delete(ctx context.Context, id string, force bool) error
}

// ScaleService defines business rules for scale criteria.
type ScaleService interface {
	GetByCompetency(ctx context.Context, competencyID string) (*dto.ScaleCriteriaResponse, error)
	Upsert(ctx context.Context, competencyID string, req dto.ScaleCriteriaBulkRequest) (*dto.ScaleCriteriaResponse, error)
}

// CatalogService defines read-only access to static catalogs.
type CatalogService interface {
	ListLevels(ctx context.Context) ([]dto.LevelDefinitionItem, error)
	ListProfiles(ctx context.Context) ([]dto.EvaluationProfileItem, error)
}

// AcceptanceService defines business rules for competency acceptance levels.
type AcceptanceService interface {
	List(ctx context.Context, filter AcceptanceFilter) ([]dto.AcceptanceLevelItem, error)
	Upsert(ctx context.Context, req dto.UpsertAcceptanceRequest) (*dto.AcceptanceLevelItem, error)
}
