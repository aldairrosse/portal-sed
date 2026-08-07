package goal

import (
	"context"

	"github.com/google/uuid"
	dtogoal "github.com/sed-evaluacion-desempeno/api/internal/dto/goal"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
	repoorganization "github.com/sed-evaluacion-desempeno/api/internal/repository/org"
)

// KPIService handles business logic for KPIs and their linking.
type KPIService struct {
	kpiRepo      KPIRepository
	linkRepo     LinkKPIRepository
	goalRepo     GoalRepository
	catRepo      CategoryRepository
	phaseCheck   *PhaseCheck
	orgNodeRepo  *repoorganization.OrgNodeRepo
	orgTreeRepo  *repoorganization.OrgTreeRepo
	employeeRepo *repoorganization.EmployeeRepo
}

// NewKPIService creates a new KPIService.
func NewKPIService(
	kpiRepo KPIRepository,
	linkRepo LinkKPIRepository,
	goalRepo GoalRepository,
	catRepo CategoryRepository,
	phaseCheck *PhaseCheck,
	orgNodeRepo *repoorganization.OrgNodeRepo,
	orgTreeRepo *repoorganization.OrgTreeRepo,
	employeeRepo *repoorganization.EmployeeRepo,
) *KPIService {
	return &KPIService{
		kpiRepo:      kpiRepo,
		linkRepo:     linkRepo,
		goalRepo:     goalRepo,
		catRepo:      catRepo,
		phaseCheck:   phaseCheck,
		orgNodeRepo:  orgNodeRepo,
		orgTreeRepo:  orgTreeRepo,
		employeeRepo: employeeRepo,
	}
}

// ListKPIs returns all KPIs.
func (s *KPIService) ListKPIs(ctx context.Context, employeeID uuid.UUID) ([]*repogoal.KpiRow, error) {
	dept, err := s.resolveDepartmentOrgNode(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	return s.kpiRepo.ListKPIs(ctx, dept)
}

// CreateKPI creates a new KPI.
func (s *KPIService) CreateKPI(ctx context.Context, req dtogoal.CreateKpiRequest, employeeID uuid.UUID) (*repogoal.KpiRow, error) {
	if req.Name == "" {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "KPI name is required", nil)
	}
	if !validUnits[req.Unit] {
		return nil, pkgerrors.ErrInvalidUnit
	}
	// Validate target_value: allow 0 for descendente or binario, otherwise > 0
	if req.TargetValue != nil && *req.TargetValue <= 0 && req.Unit != "binario" {
		return nil, pkgerrors.ErrInvalidTargetValue
	}
	// Normalize binary target
	var normalizedTarget *float64
	if req.TargetValue != nil {
		v := normalizeBinaryValue(req.Unit, *req.TargetValue)
		normalizedTarget = &v
	}
	dept, err := s.resolveDepartmentOrgNode(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	return s.kpiRepo.CreateKPI(ctx, req.Name, req.Unit, req.Description, normalizedTarget, dept)
}

// UpdateKPI updates an existing KPI.
func (s *KPIService) UpdateKPI(ctx context.Context, kpiID uuid.UUID, req dtogoal.UpdateKpiRequest) (*repogoal.KpiRow, error) {
	if req.Name == "" {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "KPI name is required", nil)
	}
	if !validUnits[req.Unit] {
		return nil, pkgerrors.ErrInvalidUnit
	}
	// Validate target_value: allow 0 for descendente or binario, otherwise > 0
	if req.TargetValue != nil && *req.TargetValue <= 0 && req.Unit != "binario" {
		return nil, pkgerrors.ErrInvalidTargetValue
	}
	// Normalize binary target
	var normalizedTarget *float64
	if req.TargetValue != nil {
		v := normalizeBinaryValue(req.Unit, *req.TargetValue)
		normalizedTarget = &v
	}
	return s.kpiRepo.UpdateKPI(ctx, kpiID, req.Name, req.Unit, req.Description, normalizedTarget)
}

// UpdateKPIValue updates only the current_value of a KPI.
func (s *KPIService) UpdateKPIValue(ctx context.Context, kpiID uuid.UUID, currentValue float64) (*repogoal.KpiRow, error) {
	if currentValue < 0 {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "current_value must be >= 0", nil)
	}
	// Get the KPI to check unit for binary normalization
	kpi, err := s.kpiRepo.GetKPI(ctx, kpiID)
	if err != nil {
		return nil, err
	}
	// Normalize binary value: any non-zero becomes 1
	normalizedValue := normalizeBinaryValue(kpi.Unit, currentValue)
	return s.kpiRepo.UpdateKPIValue(ctx, kpiID, normalizedValue)
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

func (s *KPIService) resolveDepartmentOrgNode(ctx context.Context, employeeID uuid.UUID) (*uuid.UUID, error) {
	emp, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	if emp.OrgNodeID == uuid.Nil {
		return nil, nil
	}

	path, err := s.orgNodeRepo.GetPathToRoot(ctx, emp.OrgNodeID)
	if err != nil {
		return nil, err
	}
	if len(path) == 0 {
		return nil, nil
	}

	t := path[len(path)-1]
	if t.ParentID != nil {
		for i := len(path) - 1; i >= 0; i-- {
			if path[i].ParentID == nil {
				t = path[i]
				break
			}
		}
	}

	tree, err := s.orgTreeRepo.GetByID(ctx, t.OrganizationID)
	if err != nil || tree == nil {
		return &t.ID, nil
	}

	rootID := tree.RootNodeID
	if rootID != nil && *rootID == t.ID {
		for _, n := range path {
			if n.ID != t.ID && n.ParentID != nil && *n.ParentID == *rootID {
				return &n.ID, nil
			}
		}
		return &t.ID, nil
	}

	return &t.ID, nil
}
