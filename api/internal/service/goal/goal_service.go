package goal

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	dtogoal "github.com/sed-evaluacion-desempeno/api/internal/dto/goal"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/validation"
	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
)

// validUnits is the set of allowed goal units.
var validUnits = map[string]bool{
	"porcentaje": true,
	"moneda":     true,
	"numero":     true,
}

// GoalService handles business logic for goals.
type GoalService struct {
	goalRepo      GoalRepository
	catRepo       CategoryRepository
	kpiRepo       KPIRepository
	linkRepo      LinkKPIRepository
	weightQueries WeightQuerier
	phaseCheck    *PhaseCheck
}

// NewGoalService creates a new GoalService.
func NewGoalService(
	goalRepo GoalRepository,
	catRepo CategoryRepository,
	kpiRepo KPIRepository,
	linkRepo LinkKPIRepository,
	weightQueries WeightQuerier,
	phaseCheck *PhaseCheck,
) *GoalService {
	return &GoalService{
		goalRepo:      goalRepo,
		catRepo:       catRepo,
		kpiRepo:       kpiRepo,
		linkRepo:      linkRepo,
		weightQueries: weightQueries,
		phaseCheck:    phaseCheck,
	}
}

// CreateGoal creates a new goal with weight overflow prevention.
func (s *GoalService) CreateGoal(ctx context.Context, empID, catID uuid.UUID, req dtogoal.CreateGoalRequest) (*repogoal.GoalRow, error) {
	if err := s.phaseCheck.CanCreateGoal(ctx, empID.String()); err != nil {
		return nil, err
	}

	// Validate basic fields
	if err := validateGoalRequest(req); err != nil {
		return nil, err
	}

	// Verify category belongs to employee
	cat, err := s.catRepo.LockCategory(ctx, catID)
	if err != nil {
		return nil, err
	}
	if cat.EmployeeID != empID {
		return nil, pkgerrors.ErrCategoryNotFound
	}

	// Validate weight doesn't overflow category
	existingSum, err := s.weightQueries.SumGoalWeightsByCategoryID(ctx, catID)
	if err != nil {
		return nil, err
	}
	if existingSum+req.Weight > 100.0+validation.Epsilon {
		return nil, pkgerrors.ErrGoalWeightOverflow
	}

	// Validate KPI IDs if provided
	if len(req.KpiIDs) > 5 {
		return nil, pkgerrors.ErrKpiLinkLimitExceeded
	}
	for _, kpiID := range req.KpiIDs {
		kpiUUID, err := uuid.Parse(kpiID)
		if err != nil {
			return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid kpi_id format", err)
		}
		if _, err := s.kpiRepo.GetKPI(ctx, kpiUUID); err != nil {
			return nil, err
		}
	}

	// Create goal via raw SQL (to set version=1)
	goal, err := s.goalRepo.CreateGoal(ctx, catID, req.Name, req.Description, req.Unit, req.Weight, req.TargetValue)
	if err != nil {
		return nil, fmt.Errorf("create goal: %w", err)
	}

	// Link KPIs if requested
	for _, kpiID := range req.KpiIDs {
		kpiUUID, _ := uuid.Parse(kpiID)
		if err := s.linkRepo.LinkKPI(ctx, goal.ID, kpiUUID); err != nil {
			return nil, fmt.Errorf("link kpi: %w", err)
		}
	}

	return goal, nil
}

// UpdateGoal updates a goal with optimistic locking.
func (s *GoalService) UpdateGoal(ctx context.Context, empID, goalID uuid.UUID, req dtogoal.UpdateGoalRequest) (*repogoal.GoalRow, error) {
	if err := s.phaseCheck.CanUpdateGoal(ctx, empID.String()); err != nil {
		return nil, err
	}

	// Verify the goal exists and belongs to the employee
	existing, err := s.goalRepo.GetGoal(ctx, goalID)
	if err != nil {
		return nil, err
	}

	// Get the category to verify ownership
	cat, err := s.catRepo.GetCategory(ctx, existing.CategoryID)
	if err != nil {
		return nil, err
	}
	if cat.EmployeeID != empID {
		return nil, pkgerrors.ErrGoalNotFound
	}

	// Validate basic fields
	if err := validateGoalRequest(dtogoal.CreateGoalRequest{
		Name:        req.Name,
		Description: req.Description,
		Unit:        req.Unit,
		Weight:      req.Weight,
		TargetValue: req.TargetValue,
	}); err != nil {
		return nil, err
	}

	// In avance phase, reject weight and targetValue changes
	currentPhase, _ := s.phaseCheck.CurrentPhase(ctx, empID.String())
	if currentPhase == PhaseAvance {
		if req.Weight != existing.Weight || req.TargetValue != existing.TargetValue {
			return nil, ErrPhaseRestricted
		}
	}

	// Update goal with optimistic lock
	updated, err := s.goalRepo.UpdateGoal(ctx, goalID, req.Name, req.Description, req.Unit, req.Weight, req.TargetValue, req.Version)
	if err != nil {
		return nil, err
	}

	// Replace KPI links if provided
	if req.KpiIDs != nil {
		kpiUUIDs := make([]uuid.UUID, len(req.KpiIDs))
		for i, id := range req.KpiIDs {
			kpiUUIDs[i], err = uuid.Parse(id)
			if err != nil {
				return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid kpi_id format", err)
			}
		}
		if err := s.linkRepo.ReplaceGoalKpiLinks(ctx, goalID, kpiUUIDs); err != nil {
			return nil, fmt.Errorf("replace kpi links: %w", err)
		}
	}

	return updated, nil
}

// DeleteGoal deletes a goal with phase enforcement.
func (s *GoalService) DeleteGoal(ctx context.Context, empID, goalID uuid.UUID) error {
	if err := s.phaseCheck.CanDeleteGoal(ctx, empID.String()); err != nil {
		return err
	}

	// Verify ownership
	existing, err := s.goalRepo.GetGoal(ctx, goalID)
	if err != nil {
		return err
	}
	cat, err := s.catRepo.GetCategory(ctx, existing.CategoryID)
	if err != nil {
		return err
	}
	if cat.EmployeeID != empID {
		return pkgerrors.ErrGoalNotFound
	}

	return s.goalRepo.DeleteGoal(ctx, goalID)
}

// validateGoalRequest validates a goal create/update request.
func validateGoalRequest(req dtogoal.CreateGoalRequest) error {
	if req.Name == "" {
		return pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "goal name is required", nil)
	}
	if !validUnits[req.Unit] {
		return pkgerrors.ErrInvalidUnit
	}
	if req.Weight <= 0 || req.Weight > 100 {
		return pkgerrors.ErrInvalidWeightRange
	}
	if req.TargetValue <= 0 {
		return pkgerrors.ErrInvalidTargetValue
	}
	return nil
}
