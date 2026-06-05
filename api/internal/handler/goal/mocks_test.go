package goal

import (
	"context"
	"time"

	"github.com/google/uuid"
	dtogoal "github.com/sed-evaluacion-desempeno/api/internal/dto/goal"
	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
	svcgoal "github.com/sed-evaluacion-desempeno/api/internal/service/goal"
)

// ---------------------------------------------------------------------------
// Mock services
// ---------------------------------------------------------------------------

type mockCategoryService struct {
	listFunc   func(ctx context.Context, empID uuid.UUID) ([]*repogoal.CategoryRow, error)
	createFunc func(ctx context.Context, empID uuid.UUID, req dtogoal.CreateCategoryRequest) (*repogoal.CategoryRow, error)
	updateFunc func(ctx context.Context, empID, catID uuid.UUID, req dtogoal.UpdateCategoryRequest) (*repogoal.CategoryRow, error)
	deleteFunc func(ctx context.Context, empID, catID uuid.UUID) error
}

func (m *mockCategoryService) ListCategories(ctx context.Context, empID uuid.UUID) ([]*repogoal.CategoryRow, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, empID)
	}
	return nil, nil
}

func (m *mockCategoryService) CreateCategory(ctx context.Context, empID uuid.UUID, req dtogoal.CreateCategoryRequest) (*repogoal.CategoryRow, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, empID, req)
	}
	return nil, nil
}

func (m *mockCategoryService) UpdateCategory(ctx context.Context, empID, catID uuid.UUID, req dtogoal.UpdateCategoryRequest) (*repogoal.CategoryRow, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, empID, catID, req)
	}
	return nil, nil
}

func (m *mockCategoryService) DeleteCategory(ctx context.Context, empID, catID uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, empID, catID)
	}
	return nil
}

type mockGoalService struct {
	createFunc func(ctx context.Context, empID, catID uuid.UUID, req dtogoal.CreateGoalRequest) (*repogoal.GoalRow, error)
	updateFunc func(ctx context.Context, empID, goalID uuid.UUID, req dtogoal.UpdateGoalRequest) (*repogoal.GoalRow, error)
	deleteFunc func(ctx context.Context, empID, goalID uuid.UUID) error
}

func (m *mockGoalService) CreateGoal(ctx context.Context, empID, catID uuid.UUID, req dtogoal.CreateGoalRequest) (*repogoal.GoalRow, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, empID, catID, req)
	}
	return nil, nil
}

func (m *mockGoalService) UpdateGoal(ctx context.Context, empID, goalID uuid.UUID, req dtogoal.UpdateGoalRequest) (*repogoal.GoalRow, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, empID, goalID, req)
	}
	return nil, nil
}

func (m *mockGoalService) DeleteGoal(ctx context.Context, empID, goalID uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, empID, goalID)
	}
	return nil
}

type mockProgressService struct {
	updateFunc func(ctx context.Context, empID, goalID uuid.UUID, req dtogoal.UpdateProgressRequest) (*repogoal.GoalRow, error)
}

func (m *mockProgressService) UpdateGoalProgress(ctx context.Context, empID, goalID uuid.UUID, req dtogoal.UpdateProgressRequest) (*repogoal.GoalRow, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, empID, goalID, req)
	}
	return nil, nil
}

type mockKPIService struct {
	listFunc   func(ctx context.Context) ([]*repogoal.KpiRow, error)
	createFunc func(ctx context.Context, req dtogoal.CreateKpiRequest) (*repogoal.KpiRow, error)
	updateFunc func(ctx context.Context, kpiID uuid.UUID, req dtogoal.UpdateKpiRequest) (*repogoal.KpiRow, error)
	deleteFunc func(ctx context.Context, kpiID uuid.UUID) error
	linkFunc   func(ctx context.Context, empID, goalID, kpiID uuid.UUID) error
	unlinkFunc func(ctx context.Context, empID, goalID, kpiID uuid.UUID) error
}

func (m *mockKPIService) ListKPIs(ctx context.Context) ([]*repogoal.KpiRow, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return nil, nil
}

func (m *mockKPIService) CreateKPI(ctx context.Context, req dtogoal.CreateKpiRequest) (*repogoal.KpiRow, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockKPIService) UpdateKPI(ctx context.Context, kpiID uuid.UUID, req dtogoal.UpdateKpiRequest) (*repogoal.KpiRow, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, kpiID, req)
	}
	return nil, nil
}

func (m *mockKPIService) DeleteKPI(ctx context.Context, kpiID uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, kpiID)
	}
	return nil
}

func (m *mockKPIService) LinkKPI(ctx context.Context, empID, goalID, kpiID uuid.UUID) error {
	if m.linkFunc != nil {
		return m.linkFunc(ctx, empID, goalID, kpiID)
	}
	return nil
}

func (m *mockKPIService) UnlinkKPI(ctx context.Context, empID, goalID, kpiID uuid.UUID) error {
	if m.unlinkFunc != nil {
		return m.unlinkFunc(ctx, empID, goalID, kpiID)
	}
	return nil
}

type mockWeightValidationService struct {
	validateFunc func(ctx context.Context, empID uuid.UUID) (*dtogoal.WeightValidationResponse, error)
}

func (m *mockWeightValidationService) ValidateDoubleWeighting(ctx context.Context, empID uuid.UUID) (*dtogoal.WeightValidationResponse, error) {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, empID)
	}
	return nil, nil
}

type mockBatchService struct {
	batchFunc func(ctx context.Context, empID uuid.UUID, req dtogoal.BatchGoalRequest) ([]*repogoal.GoalRow, error)
}

func (m *mockBatchService) BatchCreateUpdateGoals(ctx context.Context, empID uuid.UUID, req dtogoal.BatchGoalRequest) ([]*repogoal.GoalRow, error) {
	if m.batchFunc != nil {
		return m.batchFunc(ctx, empID, req)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// Mock repos
// ---------------------------------------------------------------------------

type mockCategoryRepo struct {
	listByEmployeeFunc func(ctx context.Context, empID uuid.UUID) ([]*repogoal.CategoryRow, error)
	getFunc            func(ctx context.Context, catID uuid.UUID) (*repogoal.CategoryRow, error)
}

func (m *mockCategoryRepo) ListCategoriesByEmployee(ctx context.Context, empID uuid.UUID) ([]*repogoal.CategoryRow, error) {
	if m.listByEmployeeFunc != nil {
		return m.listByEmployeeFunc(ctx, empID)
	}
	return nil, nil
}

func (m *mockCategoryRepo) CreateCategory(ctx context.Context, empID uuid.UUID, name, description string, weight float64) (*repogoal.CategoryRow, error) {
	return nil, nil
}

func (m *mockCategoryRepo) UpdateCategory(ctx context.Context, catID uuid.UUID, name, description string, weight float64) (*repogoal.CategoryRow, error) {
	return nil, nil
}

func (m *mockCategoryRepo) DeleteCategory(ctx context.Context, catID uuid.UUID) error {
	return nil
}

func (m *mockCategoryRepo) GetCategory(ctx context.Context, catID uuid.UUID) (*repogoal.CategoryRow, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, catID)
	}
	return nil, nil
}

func (m *mockCategoryRepo) LockCategory(ctx context.Context, catID uuid.UUID) (*repogoal.CategoryRow, error) {
	return nil, nil
}

type mockGoalRepo struct {
	getFunc            func(ctx context.Context, goalID uuid.UUID) (*repogoal.GoalRow, error)
	listByCategoryFunc func(ctx context.Context, catID uuid.UUID) ([]*repogoal.GoalRow, error)
}

func (m *mockGoalRepo) CreateGoal(ctx context.Context, catID uuid.UUID, name, description, unit string, weight, targetValue float64) (*repogoal.GoalRow, error) {
	return nil, nil
}

func (m *mockGoalRepo) GetGoal(ctx context.Context, goalID uuid.UUID) (*repogoal.GoalRow, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, goalID)
	}
	return nil, nil
}

func (m *mockGoalRepo) UpdateGoal(ctx context.Context, goalID uuid.UUID, name, description, unit string, weight, targetValue float64, expectedVersion int) (*repogoal.GoalRow, error) {
	return nil, nil
}

func (m *mockGoalRepo) DeleteGoal(ctx context.Context, goalID uuid.UUID) error {
	return nil
}

func (m *mockGoalRepo) UpdateGoalCurrentValue(ctx context.Context, goalID uuid.UUID, currentValue float64) (*repogoal.GoalRow, error) {
	return nil, nil
}

func (m *mockGoalRepo) ListGoalsByCategory(ctx context.Context, catID uuid.UUID) ([]*repogoal.GoalRow, error) {
	if m.listByCategoryFunc != nil {
		return m.listByCategoryFunc(ctx, catID)
	}
	return nil, nil
}

type mockKpiRepo struct{}

func (m *mockKpiRepo) ListKPIs(ctx context.Context) ([]*repogoal.KpiRow, error)                  { return nil, nil }
func (m *mockKpiRepo) GetKPI(ctx context.Context, kpiID uuid.UUID) (*repogoal.KpiRow, error)       { return nil, nil }
func (m *mockKpiRepo) CreateKPI(ctx context.Context, name, unit, description string) (*repogoal.KpiRow, error) {
	return nil, nil
}
func (m *mockKpiRepo) UpdateKPI(ctx context.Context, kpiID uuid.UUID, name, unit, description string) (*repogoal.KpiRow, error) {
	return nil, nil
}
func (m *mockKpiRepo) DeleteKPI(ctx context.Context, kpiID uuid.UUID) error                     { return nil }
func (m *mockKpiRepo) CountGoalLinksByKPI(ctx context.Context, kpiID uuid.UUID) (int, error)    { return 0, nil }

type mockLinkRepo struct{}

func (m *mockLinkRepo) LinkKPI(ctx context.Context, goalID, kpiID uuid.UUID) error                    { return nil }
func (m *mockLinkRepo) UnlinkKPI(ctx context.Context, goalID, kpiID uuid.UUID) error                  { return nil }
func (m *mockLinkRepo) CountGoalKPILinks(ctx context.Context, goalID uuid.UUID) (int, error)          { return 0, nil }
func (m *mockLinkRepo) ReplaceGoalKpiLinks(ctx context.Context, goalID uuid.UUID, kpiIDs []uuid.UUID) error {
	return nil
}
func (m *mockLinkRepo) ListKpiIDsByGoal(ctx context.Context, goalID uuid.UUID) ([]uuid.UUID, error) { return nil, nil }

type mockAssignmentRepo struct {
	getFunc    func(ctx context.Context, empID uuid.UUID) (*repogoal.AssignmentRow, error)
	createFunc func(ctx context.Context, empID, cycleID uuid.UUID) (*repogoal.AssignmentRow, error)
}

func (m *mockAssignmentRepo) GetAssignment(ctx context.Context, empID uuid.UUID) (*repogoal.AssignmentRow, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, empID)
	}
	return nil, nil
}

func (m *mockAssignmentRepo) CreateAssignment(ctx context.Context, empID, cycleID uuid.UUID) (*repogoal.AssignmentRow, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, empID, cycleID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestHandler(
	catSvc svcgoal.CategoryServicer,
	goalSvc svcgoal.GoalServicer,
	progSvc svcgoal.ProgressServicer,
	kpiSvc svcgoal.KpiServicer,
	weightSvc svcgoal.WeightValidationServicer,
	batchSvc svcgoal.BatchServicer,
	catRepo svcgoal.CategoryRepository,
	goalRepo svcgoal.GoalRepository,
	kpiRepo svcgoal.KPIRepository,
	linkRepo svcgoal.LinkKPIRepository,
	assignRepo svcgoal.AssignmentRepository,
) *GoalHandler {
	if catSvc == nil {
		catSvc = &mockCategoryService{}
	}
	if goalSvc == nil {
		goalSvc = &mockGoalService{}
	}
	if progSvc == nil {
		progSvc = &mockProgressService{}
	}
	if kpiSvc == nil {
		kpiSvc = &mockKPIService{}
	}
	if weightSvc == nil {
		weightSvc = &mockWeightValidationService{}
	}
	if batchSvc == nil {
		batchSvc = &mockBatchService{}
	}
	if catRepo == nil {
		catRepo = &mockCategoryRepo{}
	}
	if goalRepo == nil {
		goalRepo = &mockGoalRepo{}
	}
	if kpiRepo == nil {
		kpiRepo = &mockKpiRepo{}
	}
	if linkRepo == nil {
		linkRepo = &mockLinkRepo{}
	}
	if assignRepo == nil {
		assignRepo = &mockAssignmentRepo{}
	}
	return NewGoalHandler(catSvc, goalSvc, progSvc, kpiSvc, weightSvc, batchSvc, catRepo, goalRepo, kpiRepo, linkRepo, assignRepo)
}

func mustParseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}

func fixedTime() time.Time {
	return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
}
