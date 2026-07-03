// Package evaluation provides shared DTOs for the evaluation and 9×9 API.
// These types are used by the service and handler layers for request/response
// serialization.
package evaluation

import (
	"time"

	"github.com/google/uuid"
)

// --- Evaluation DTOs ---

// CompetencyRatingInput is the request body for rating a single competency.
type CompetencyRatingInput struct {
	CompetencyID uuid.UUID `json:"competencyId" validate:"required"`
	Rating       int       `json:"rating" validate:"min=1,max=5"`
	Comments     string    `json:"comments,omitempty"`
}

// GoalCommentInput is the request body for adding comments to a goal.
type GoalCommentInput struct {
	GoalID  uuid.UUID `json:"goalId" validate:"required"`
	Comment string    `json:"comment,omitempty"`
}

// GoalStateUpdateInput is the request body for PUT /evaluations/{id}/goal-state.
// All fields except GoalID are optional; only the provided ones are updated.
type GoalStateUpdateInput struct {
	GoalID         uuid.UUID `json:"goalId" validate:"required"`
	FinalProgress  *float64  `json:"finalProgress,omitempty"`
	SelfAssessment *string   `json:"selfAssessment,omitempty"`
	RhAssessment   *string   `json:"rhAssessment,omitempty"`
}

// GoalCommentUpdateInput is the request body for PUT /evaluations/{id}/goal-comments.
type GoalCommentUpdateInput struct {
	GoalID  uuid.UUID `json:"goalId" validate:"required"`
	Role    string    `json:"role" validate:"required,oneof=manager"`
	Comment string    `json:"comment,omitempty"`
}

// SelfEvaluationRequest is the request body for submitting a self-evaluation.
type SelfEvaluationRequest struct {
	Competencies []CompetencyRatingInput `json:"competencies" validate:"required,min=1,dive"`
	GoalComments []GoalCommentInput      `json:"goalComments,omitempty"`
}

// RHEvaluationRequest is the request body for submitting an RH evaluation.
type RHEvaluationRequest struct {
	Competencies  []CompetencyRatingInput `json:"competencies" validate:"required,min=1,dive"`
	FinalComments string                  `json:"finalComments,omitempty"`
}

// FinalizeEvaluationRequest is the optional request body for finalizing.
type FinalizeEvaluationRequest struct {
	Reason string `json:"reason,omitempty"`
}

// CompetencyRatingDTO is the response DTO for a competency rating.
type CompetencyRatingDTO struct {
	CompetencyID uuid.UUID `json:"competencyId"`
	Rating       int       `json:"rating"`
	Comments     string    `json:"comments,omitempty"`
}

// GoalRatingDTO is the response DTO for a goal rating.
type GoalRatingDTO struct {
	GoalID        uuid.UUID `json:"goalId"`
	FinalRating   *int      `json:"finalRating,omitempty"`
	FinalComments string    `json:"finalComments,omitempty"`
}

// EvaluationListItem is the lightweight DTO for evaluation list responses.
type EvaluationListItem struct {
	ID         uuid.UUID `json:"id"`
	EmployeeID uuid.UUID `json:"employeeId"`
	CycleID    uuid.UUID `json:"cycleId"`
	State      string    `json:"state"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// EvaluationDetailResponse is the full DTO for a single evaluation.
type EvaluationDetailResponse struct {
	ID                      uuid.UUID             `json:"id"`
	EmployeeID              uuid.UUID             `json:"employeeId"`
	CycleID                 uuid.UUID             `json:"cycleId"`
	State                   string                `json:"state"`
	SelfEvalCompletedAt     *time.Time            `json:"selfEvaluationCompletedAt,omitempty"`
	RHEvalCompletedAt       *time.Time            `json:"rhEvaluationCompletedAt,omitempty"`
	CompetencyRatings       []CompetencyRatingDTO `json:"competencies"`
	GoalRatings             []GoalRatingDTO       `json:"goals"`
	Version                 int                   `json:"version"`
	CreatedAt               time.Time             `json:"createdAt"`
	UpdatedAt               time.Time             `json:"updatedAt"`
}

// EvaluationListResponse is the paginated list response.
type EvaluationListResponse struct {
	Data       []EvaluationListItem `json:"data"`
	NextCursor string               `json:"nextCursor,omitempty"`
}

// EvaluationSummaryResponse is the dashboard summary response.
type EvaluationSummaryResponse struct {
	CycleID uuid.UUID        `json:"cycleId"`
	Counts  map[string]int64 `json:"counts"`
}

// CompetencyResultItem is a single employee's competency average in the paginated response.
type CompetencyResultItem struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	ProfileName   string   `json:"profileName"`
	SelfRatingAvg *float64 `json:"selfRatingAvg"`
	RHRatingAvg   *float64 `json:"rhRatingAvg"`
	Status        string   `json:"status"`
}

// CompetencyResultsResponse is the paginated response for competency results.
type CompetencyResultsResponse struct {
	Data []CompetencyResultItem `json:"data"`
	Meta PaginationMeta         `json:"meta"`
}

// PaginationMeta carries pagination metadata in list responses.
type PaginationMeta struct {
	HasMore bool `json:"hasMore"`
	Total   int  `json:"total"`
	Offset  int  `json:"offset"`
	Limit   int  `json:"limit"`
}

// --- Nine-Box DTOs ---

// --- Employee Competency Ratings DTOs ---

// EmployeeCompetencyRatingDTO is a single competency rating in the employee response.
type EmployeeCompetencyRatingDTO struct {
	CompetencyID    uuid.UUID `json:"competencyId"`
	SelfRating      *int      `json:"selfRating,omitempty"`
	RhRating        *int      `json:"rhRating,omitempty"`
	Comments        *string   `json:"comments,omitempty"`
	AcceptanceLevel *int      `json:"acceptanceLevel,omitempty"`
}

// EmployeeCompetencyRatingsResponse is the response for GET /evaluations/employee/{employeeId}.
type EmployeeCompetencyRatingsResponse struct {
	EmployeeID uuid.UUID                     `json:"employeeId"`
	CycleID    uuid.UUID                     `json:"cycleId"`
	Ratings    []EmployeeCompetencyRatingDTO `json:"ratings"`
}

// NineBoxMatrixResponse is the response DTO for a matrix.
type NineBoxMatrixResponse struct {
	ID          uuid.UUID         `json:"id"`
	CycleID     uuid.UUID         `json:"cycleId"`
	EvaluatorID uuid.UUID         `json:"evaluatorId"`
	PhaseID     uuid.UUID         `json:"phaseId"`
	PhaseLabel  string            `json:"phaseLabel,omitempty"`
	Entries     []NineBoxEntryDTO `json:"entries"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// NineBoxEntryDTO is the response DTO for a matrix entry (tier-based).
type NineBoxEntryDTO struct {
	ID              uuid.UUID `json:"id"`
	EvaluateeID     uuid.UUID `json:"evaluateeId"`
	EmployeeName    string    `json:"employeeName"`
	ProfileID       uuid.UUID `json:"profileId"`
	PerformanceTier int       `json:"performanceTier"` // was performanceScore
	PotentialTier   int       `json:"potentialTier"`   // was potentialScore
	Quadrant        int       `json:"quadrant"`
	QuadrantLabel   string    `json:"quadrantLabel"`
	QuadrantColor   string    `json:"quadrantColor"` // now uses colorHex from quadrant
	Comments        string    `json:"comments,omitempty"`
	Version         int       `json:"version"`
}

// NineBoxEntryInput is the request DTO for creating/updating a matrix entry (legacy, preserved for migration).
// Deprecated: Use RecomputeMatrix for automatic tier computation.
type NineBoxEntryInput struct {
	EvaluateeID      uuid.UUID `json:"evaluateeId" validate:"required"`
	PerformanceScore int       `json:"performanceScore" validate:"min=1,max=9"`
	PotentialScore   int       `json:"potentialScore" validate:"min=1,max=9"`
	Comments        *string   `json:"comments,omitempty"`
}

// NineBoxBatchRequest is the request DTO for batch submission.
// Deprecated: Use RecomputeMatrix for automatic tier computation.
type NineBoxBatchRequest struct {
	Entries []NineBoxEntryInput `json:"entries" validate:"required,min=1,max=20,dive"`
}

// NineBoxScaleDTO is the response DTO for a scale definition.
type NineBoxScaleDTO struct {
	Axis        string `json:"axis"`
	Level       int    `json:"level"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// NineBoxQuadrantDTO is the response DTO for a quadrant definition.
type NineBoxQuadrantDTO struct {
	Quadrant             int    `json:"quadrant"`
	Label                string `json:"label"`
	Title                string `json:"title,omitempty"`
	Description          string `json:"description"`
	Color                string `json:"color"`
	ColorHex             string `json:"colorHex,omitempty"`
	ActionRecommendation string `json:"actionRecommendation"`
}

// NineBoxQuadrantUpdateInput is the request DTO for updating a quadrant (RH edit).
type NineBoxQuadrantUpdateInput struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	ColorHex    string `json:"colorHex" validate:"omitempty,hexcolor"`
}
