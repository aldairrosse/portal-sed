package goal

import (
	"context"

	"github.com/google/uuid"
	dtogoal "github.com/sed-evaluacion-desempeno/api/internal/dto/goal"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
)

// KPIService handles business logic for KPIs and their linking.
type KPIService struct {
	kpiRepo    KPIRepository
	linkRepo   LinkKPIRepository
	goalRepo   GoalRepository
	catRepo    CategoryRepository
	phaseCheck *PhaseCheck
}

// NewKPIService creates a new KPIService.
func NewKPIService(
	kpiRepo KPIRepository,
	linkRepo LinkKPIRepository,
	goalRepo GoalRepository,
	catRepo CategoryRepository,
	phaseCheck *PhaseCheck,
) *KPIService {
	return &KPIService{
		kpiRepo:    kpiRepo,
		linkRepo:   linkRepo,
		goalRepo:   goalRepo,
		catRepo:    catRepo,
		phaseCheck: phaseCheck,
	}
}

// ListKPIs returns all KPIs.
func (s *KPIService) ListKPIs(ctx context.Context) ([]*repogoal.KpiRow, error) {
	return s.kpiRepo.ListKPIs(ctx)
}

// CreateKPI creates a new KPI.
func (s *KPIService) CreateKPI(ctx context.Context, req dtogoal.CreateKpiRequest) (*repogoal.KpiRow, error) {
	if req.Name == "" {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "KPI name is required", nil)
	}
	if !validUnits[req.Unit] {
		return nil, pkgerrors.ErrInvalidUnit
	}
	return s.kpiRepo.CreateKPI(ctx, req.Name, req.Unit, req.Description)
}

// UpdateKPI updates an existing KPI.
func (s *KPIService) UpdateKPI(ctx context.Context, kpiID uuid.UUID, req dtogoal.UpdateKpiRequest) (*repogoal.KpiRow, error) {
	if req.Name == "" {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "KPI name is required", nil)
	}
	if !validUnits[req.Unit] {
		return nil, pkgerrors.ErrInvalidUnit
	}
	return s.kpiRepo.UpdateKPI(ctx, kpiID, req.Name, req.Unit, req.Description)
}

// DeleteKPI deletes a KPI. Rejects if linked to any goals.
func (s *KPIService) DeleteKPI(ctx context.Context, kpiID uuid.UUID) error {
	return s.kpiRepo.DeleteKPI(ctx, kpiID)
}

// LinkKPI links a KPI to a goal. Gated to asignacion phase.
func (s *KPIService) LinkKPI(ctx context.Context, empID, goalID, kpiID uuid.UUID) error {
	if err := s.phaseCheck.CanLinkKPI(ctx, empID.String()); err != nil {
		return err
	}

	// Verify goal exists and belongs to employee
	goal, err := s.goalRepo.GetGoal(ctx, goalID)
	if err != nil {
		return err
	}
	cat, err := s.catRepo.GetCategory(ctx, goal.CategoryID)
	if err != nil {
		return err
	}
	if cat.EmployeeID != empID {
		return pkgerrors.ErrGoalNotFound
	}

	// Verify KPI exists
	if _, err := s.kpiRepo.GetKPI(ctx, kpiID); err != nil {
		return err
	}

	// Check link limit
	count, err := s.linkRepo.CountGoalKPILinks(ctx, goalID)
	if err != nil {
		return err
	}
	if count >= 5 {
		return pkgerrors.ErrKpiLinkLimitExceeded
	}

	return s.linkRepo.LinkKPI(ctx, goalID, kpiID)
}

// UnlinkKPI removes a KPI link from a goal. Gated to asignacion phase.
func (s *KPIService) UnlinkKPI(ctx context.Context, empID, goalID, kpiID uuid.UUID) error {
	if err := s.phaseCheck.CanLinkKPI(ctx, empID.String()); err != nil {
		return err
	}

	// Verify goal exists and belongs to employee
	goal, err := s.goalRepo.GetGoal(ctx, goalID)
	if err != nil {
		return err
	}
	cat, err := s.catRepo.GetCategory(ctx, goal.CategoryID)
	if err != nil {
		return err
	}
	if cat.EmployeeID != empID {
		return pkgerrors.ErrGoalNotFound
	}

	return s.linkRepo.UnlinkKPI(ctx, goalID, kpiID)
}

// ListKpiIDsByGoal returns the KPI IDs linked to a goal.
func (s *KPIService) ListKpiIDsByGoal(ctx context.Context, goalID uuid.UUID) ([]uuid.UUID, error) {
	return s.linkRepo.ListKpiIDsByGoal(ctx, goalID)
}
