package goal

import (
	"context"
	"log"

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
	evalLookup EvaluationLookup
}

// NewProgressService creates a new ProgressService.
// evalLookup may be nil (snapshot write is best-effort and skipped).
func NewProgressService(
	goalRepo GoalRepository,
	catRepo CategoryRepository,
	phaseCheck *PhaseCheck,
	evalLookup EvaluationLookup,
) *ProgressService {
	return &ProgressService{
		goalRepo:   goalRepo,
		catRepo:    catRepo,
		phaseCheck: phaseCheck,
		evalLookup: evalLookup,
	}
}

// UpdateGoalProgress updates the currentValue of a goal and the snapshot of
// the active phase (avance/medio-anio -> avance_progress, cierre -> cierre_progress).
// Only allowed in "avance" (mid-year) and "cierre" (year-end) phases.
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

	// Sync core: goals.current_value always tracks the latest written value.
	row, err := s.goalRepo.UpdateGoalCurrentValue(ctx, goalID, req.CurrentValue, &empID)
	if err != nil {
		return nil, err
	}

	// Snapshot of the active phase (best-effort: never fails the progress write).
	// Phase resolution: explicit request phase wins, then cycles.current_phase,
	// then the evaluation's own phase. Missing evaluation skips with a log.
	if s.evalLookup != nil {
		phase := ""
		if p, perr := s.phaseCheck.CurrentPhase(ctx, empID.String()); perr == nil {
			phase = string(p)
		}
		if req.Phase != nil && *req.Phase != "" {
			phase = *req.Phase
		}
		cycleID, cerr := s.phaseCheck.ActiveCycleID(ctx, empID.String())
		if cerr != nil {
			log.Printf("[progress] skip snapshot goal=%s: no active cycle: %v", goalID, cerr)
		} else if evalRow, ferr := s.evalLookup.FindByEmployeeCycle(ctx, empID, cycleID); ferr != nil || evalRow == nil {
			log.Printf("[progress] skip snapshot goal=%s: no evaluation for cycle=%s: %v", goalID, cycleID, ferr)
		} else {
			if phase == "" {
				phase = evalRow.Phase
			}
			_ = s.goalRepo.UpsertProgressSnapshot(ctx, evalRow.ID, goalID, phase, req.CurrentValue)
		}
	}

	return row, nil
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
