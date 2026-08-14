package goal

import (
	"context"

	"github.com/google/uuid"
	dtogoal "github.com/sed-evaluacion-desempeno/api/internal/dto/goal"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
)

// ProgressService handles progress updates for goals.
type ProgressService struct {
	goalRepo   GoalRepository
	catRepo    CategoryRepository
	phaseCheck *PhaseCheck
}

// NewProgressService creates a new ProgressService.
func NewProgressService(
	goalRepo GoalRepository,
	catRepo CategoryRepository,
	phaseCheck *PhaseCheck,
) *ProgressService {
	return &ProgressService{
		goalRepo:   goalRepo,
		catRepo:    catRepo,
		phaseCheck: phaseCheck,
	}
}

// UpdateGoalProgress updates the currentValue of a goal.
// Only allowed in the "avance" phase.
func (s *ProgressService) UpdateGoalProgress(ctx context.Context, empID, goalID uuid.UUID, req dtogoal.UpdateProgressRequest) (*repogoal.GoalRow, error) {
	if err := s.phaseCheck.CanUpdateProgress(ctx, empID.String()); err != nil {
		return nil, err
	}

	if req.CurrentValue < 0 {
		return nil, pkgerrors.ErrInvalidRequest
	}

	// Verify ownership
	existing, err := s.goalRepo.GetGoal(ctx, goalID)
	if err != nil {
		return nil, err
	}
	if existing.CategoryID == nil {
		return nil, pkgerrors.ErrGoalNotFound
	}
	cat, err := s.catRepo.GetCategory(ctx, *existing.CategoryID)
	if err != nil {
		return nil, err
	}
	if cat.EmployeeID != empID {
		return nil, pkgerrors.ErrGoalNotFound
	}

	if err := validateProgressValue(existing.Unit, existing.Direction, existing.BaselineValue, req.CurrentValue); err != nil {
		return nil, err
	}

	return s.goalRepo.UpdateGoalCurrentValue(ctx, goalID, req.CurrentValue, &empID)
}

// validateProgressValue validates a current_value update for a goal.
func validateProgressValue(unit, direction string, baselineValue *float64, currentValue float64) error {
	if currentValue < 0 {
		return pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "el progreso no puede ser negativo", nil)
	}
	if unit == "binario" && currentValue != 0 && currentValue != 1 {
		return pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "el progreso de una meta binaria debe ser 0 o 1", nil)
	}
	if direction == "descendente" && baselineValue != nil && currentValue > *baselineValue {
		return pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "el avance no puede ser mayor al valor inicial en objetivos descendentes", nil)
	}
	return nil
}
