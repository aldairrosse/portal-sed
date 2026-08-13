package evaluation

import (
	"context"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/evaluation"
)

// EvalService defines the evaluation lifecycle operations used by the handler.
type EvalService interface {
	ListEvaluations(ctx context.Context, cycleID uuid.UUID, stateFilter string, cursor string, limit int) (*dto.EvaluationListResponse, error)
	GetEvaluation(ctx context.Context, id uuid.UUID) (*dto.EvaluationDetailResponse, error)
	GetEmployeeCompetencyRatings(ctx context.Context, employeeID, cycleID uuid.UUID) (*dto.EmployeeCompetencyRatingsResponse, error)
	ResolveActiveCycleID(ctx context.Context, employeeID uuid.UUID) (uuid.UUID, error)
	SubmitSelfEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.SelfEvaluationRequest, idempotencyKey string) (*dto.EvaluationDetailResponse, error)
	UpdateSelfEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.SelfEvaluationRequest, ifMatch int) (*dto.EvaluationDetailResponse, error)
	SubmitRHEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.RHEvaluationRequest, idempotencyKey string) (*dto.EvaluationDetailResponse, error)
	UpdateRHEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.RHEvaluationRequest, ifMatch int) (*dto.EvaluationDetailResponse, error)
	FinalizeEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.FinalizeEvaluationRequest) (*dto.EvaluationDetailResponse, error)
	GetCompetencyResults(ctx context.Context, cycleID uuid.UUID, query string, scope string, currentUserID uuid.UUID, offset, limit int) (*dto.CompetencyResultsResponse, error)
	UpdateGoalState(ctx context.Context, evaluationID uuid.UUID, input dto.GoalStateUpdateInput, ifMatch int) (*dto.EvaluationDetailResponse, error)
	UpdateGoalComments(ctx context.Context, evaluationID uuid.UUID, input dto.GoalCommentUpdateInput, ifMatch int) (*dto.EvaluationDetailResponse, error)
}

// BoxService defines the 9×9 matrix operations used by the handler.
type BoxService interface {
	ListMatrices(ctx context.Context, cycleID, evaluatorID, phaseID uuid.UUID) ([]dto.NineBoxMatrixResponse, error)
	ComputeMatrixView(ctx context.Context, cycleID uuid.UUID, phaseID *uuid.UUID, viewerID uuid.UUID, viewerRole auth.Role) ([]dto.NineBoxMatrixResponse, error)
	CreateMatrix(ctx context.Context, cycleID, evaluatorID uuid.UUID) (*dto.NineBoxMatrixResponse, error)
	GetMatrix(ctx context.Context, matrixID uuid.UUID) (*dto.NineBoxMatrixResponse, error)
	GetMatrixEntriesFiltered(ctx context.Context, matrixID uuid.UUID, quadrant *int) ([]dto.NineBoxEntryDTO, error)
	CanViewMatrix(ctx context.Context, viewerID uuid.UUID, viewerRole auth.Role, matrixID uuid.UUID) (bool, error)
	RecomputeMatrix(ctx context.Context, cycleID, phaseID uuid.UUID) error
	UpdateQuadrantByNumber(ctx context.Context, quadrantNumber int, input dto.NineBoxQuadrantUpdateInput) (*dto.NineBoxQuadrantDTO, error)
	GetScales(ctx context.Context) ([]dto.NineBoxScaleDTO, error)
	GetQuadrants(ctx context.Context) ([]dto.NineBoxQuadrantDTO, error)
}

// DashService defines the dashboard operations used by the handler.
type DashService interface {
	GetSummary(ctx context.Context, cycleID uuid.UUID) (*dto.EvaluationSummaryResponse, error)
}
