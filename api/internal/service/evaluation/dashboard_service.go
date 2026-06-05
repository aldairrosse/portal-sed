package evaluation

import (
	"context"

	"github.com/google/uuid"
	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/evaluation"
)

// DashboardService provides lightweight dashboard reads using the materialized view.
type DashboardService struct {
	evalRepo EvaluationRepo
}

// NewDashboardService creates a new DashboardService.
func NewDashboardService(evalRepo EvaluationRepo) *DashboardService {
	return &DashboardService{evalRepo: evalRepo}
}

// GetSummary returns evaluation counts by state for a given cycle.
// Reads from the evaluation_summary materialized view.
func (s *DashboardService) GetSummary(ctx context.Context, cycleID uuid.UUID) (*dto.EvaluationSummaryResponse, error) {
	counts, err := s.evalRepo.GetSummaryByCycle(ctx, cycleID)
	if err != nil {
		return nil, err
	}

	return &dto.EvaluationSummaryResponse{
		CycleID: cycleID,
		Counts:  counts,
	}, nil
}
