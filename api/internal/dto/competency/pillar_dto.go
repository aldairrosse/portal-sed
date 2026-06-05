// Package competency provides data transfer objects for the competency framework API.
package competency

import "time"

// PillarListItem is the light projection used in pillar list responses.
type PillarListItem struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	CompetencyCount int    `json:"competency_count"`
}

// PillarDetail is the full projection of a pillar, optionally including competencies.
type PillarDetail struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Description  string           `json:"description,omitempty"`
	Competencies []CompetencyLite `json:"competencies,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// CreatePillarRequest is the request body for POST /api/v1/pillars.
type CreatePillarRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty" validate:"max=2000"`
}

// UpdatePillarRequest is the request body for PUT /api/v1/pillars/:id.
type UpdatePillarRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty" validate:"max=2000"`
}

// CompetencyLite is the light projection used in nested competency arrays.
type CompetencyLite struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CompetencyDetail is the full projection of a competency with scale criteria.
type CompetencyDetail struct {
	ID            string              `json:"id"`
	PillarID      string              `json:"pillar_id"`
	Name          string              `json:"name"`
	Description   string              `json:"description,omitempty"`
	ScaleCriteria map[int][]string    `json:"scale_criteria,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// CreateCompetencyRequest is the request body for POST /api/v1/pillars/:pillarId/competencies.
type CreateCompetencyRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty" validate:"max=2000"`
}

// UpdateCompetencyRequest is the request body for PUT /api/v1/competencies/:id.
type UpdateCompetencyRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty" validate:"max=2000"`
	PillarID    string `json:"pillar_id,omitempty" validate:"omitempty,uuid"`
}

// ScaleCriterionItem is a single criterion in a bulk request.
type ScaleCriterionItem struct {
	Level       int    `json:"level" validate:"required,min=1,max=5"`
	Description string `json:"description" validate:"required,max=2000"`
}

// ScaleCriteriaBulkRequest is the request body for POST /api/v1/competencies/:id/scale-criteria.
type ScaleCriteriaBulkRequest struct {
	Criteria []ScaleCriterionItem `json:"criteria" validate:"required,dive"`
}

// ScaleCriteriaResponse is the response for scale criteria endpoints.
type ScaleCriteriaResponse struct {
	CompetencyID string           `json:"competency_id"`
	Criteria     map[int][]string `json:"criteria"`
	Version      int              `json:"version"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// LevelDefinitionItem is a single level definition from the static catalog.
type LevelDefinitionItem struct {
	Level       int    `json:"level"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// EvaluationProfileItem is a single evaluation profile from the static catalog.
type EvaluationProfileItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// AcceptanceLevelItem represents a competency acceptance level per profile.
type AcceptanceLevelItem struct {
	ID           string    `json:"id"`
	CompetencyID string    `json:"competency_id"`
	ProfileID    string    `json:"profile_id"`
	Level        int       `json:"level"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UpsertAcceptanceRequest is the request body for POST /api/v1/acceptance-levels.
type UpsertAcceptanceRequest struct {
	CompetencyID string `json:"competency_id" validate:"required,uuid"`
	ProfileID    string `json:"profile_id" validate:"required,uuid"`
	Level        int    `json:"level" validate:"required,min=1,max=5"`
}

// PaginatedResponse wraps paginated list results.
type PaginatedResponse[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Pagination holds cursor-based pagination metadata.
type Pagination struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}
