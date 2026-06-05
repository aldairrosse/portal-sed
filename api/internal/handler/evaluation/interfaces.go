package evaluation

import (
	"context"

	"github.com/google/uuid"
	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/evaluation"
)

// EvalService defines the evaluation lifecycle operations used by the handler.
type EvalService interface {
	ListEvaluations(ctx context.Context, cycleID uuid.UUID, stateFilter string, cursor string, limit int) (*dto.EvaluationListResponse, error)
	GetEvaluation(ctx context.Context, id uuid.UUID) (*dto.EvaluationDetailResponse, error)
	SubmitSelfEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.SelfEvaluationRequest, idempotencyKey string) (*dto.EvaluationDetailResponse, error)
	UpdateSelfEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.SelfEvaluationRequest, ifMatch int) (*dto.EvaluationDetailResponse, error)
	SubmitRHEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.RHEvaluationRequest, idempotencyKey string) (*dto.EvaluationDetailResponse, error)
	UpdateRHEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.RHEvaluationRequest, ifMatch int) (*dto.EvaluationDetailResponse, error)
	FinalizeEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.FinalizeEvaluationRequest) (*dto.EvaluationDetailResponse, error)
}

// BoxService defines the 9×9 matrix operations used by the handler.
type BoxService interface {
	ListMatrices(ctx context.Context, cycleID, evaluatorID uuid.UUID) ([]dto.NineBoxMatrixResponse, error)
	CreateMatrix(ctx context.Context, cycleID, evaluatorID uuid.UUID) (*dto.NineBoxMatrixResponse, error)
	GetMatrix(ctx context.Context, matrixID uuid.UUID) (*dto.NineBoxMatrixResponse, error)
	UpsertEntry(ctx context.Context, matrixID uuid.UUID, req dto.NineBoxEntryInput) (*dto.NineBoxEntryDTO, error)
	UpdateEntry(ctx context.Context, entryID uuid.UUID, req dto.NineBoxEntryInput, ifMatch int) (*dto.NineBoxEntryDTO, error)
	BatchSubmitEntries(ctx context.Context, matrixID uuid.UUID, req dto.NineBoxBatchRequest) ([]dto.NineBoxEntryDTO, error)
	GetScales(ctx context.Context) ([]dto.NineBoxScaleDTO, error)
	GetQuadrants(ctx context.Context) ([]dto.NineBoxQuadrantDTO, error)
}

// DashService defines the dashboard operations used by the handler.
type DashService interface {
	GetSummary(ctx context.Context, cycleID uuid.UUID) (*dto.EvaluationSummaryResponse, error)
}
