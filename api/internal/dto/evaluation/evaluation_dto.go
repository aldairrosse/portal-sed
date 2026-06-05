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

// --- Nine-Box DTOs ---

// NineBoxMatrixResponse is the response DTO for a matrix.
type NineBoxMatrixResponse struct {
	ID          uuid.UUID         `json:"id"`
	CycleID     uuid.UUID         `json:"cycleId"`
	EvaluatorID uuid.UUID         `json:"evaluatorId"`
	Entries     []NineBoxEntryDTO `json:"entries"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// NineBoxEntryDTO is the response DTO for a matrix entry.
type NineBoxEntryDTO struct {
	ID               uuid.UUID `json:"id"`
	EvaluateeID      uuid.UUID `json:"evaluateeId"`
	PerformanceScore int       `json:"performanceScore"`
	PotentialScore   int       `json:"potentialScore"`
	Quadrant         int       `json:"quadrant"`
	QuadrantLabel    string    `json:"quadrantLabel"`
	QuadrantColor    string    `json:"quadrantColor"`
	Comments         string    `json:"comments,omitempty"`
	Version          int       `json:"version"`
}

// NineBoxEntryInput is the request DTO for creating/updating a matrix entry.
type NineBoxEntryInput struct {
	EvaluateeID      uuid.UUID `json:"evaluateeId" validate:"required"`
	PerformanceScore int       `json:"performanceScore" validate:"min=1,max=9"`
	PotentialScore   int       `json:"potentialScore" validate:"min=1,max=9"`
	Comments         string    `json:"comments,omitempty"`
}

// NineBoxBatchRequest is the request DTO for batch submission.
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
	Description          string `json:"description"`
	Color                string `json:"color"`
	ActionRecommendation string `json:"actionRecommendation"`
}
