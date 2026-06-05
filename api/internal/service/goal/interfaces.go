package goal

import (
	"context"

	"github.com/google/uuid"
	dtogoal "github.com/sed-evaluacion-desempeno/api/internal/dto/goal"
	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
)

// CategoryRepository defines the storage contract for categories.
type CategoryRepository interface {
	ListCategoriesByEmployee(ctx context.Context, empID uuid.UUID) ([]*repogoal.CategoryRow, error)
	CreateCategory(ctx context.Context, empID uuid.UUID, name, description string, weight float64) (*repogoal.CategoryRow, error)
	UpdateCategory(ctx context.Context, catID uuid.UUID, name, description string, weight float64) (*repogoal.CategoryRow, error)
	DeleteCategory(ctx context.Context, catID uuid.UUID) error
	GetCategory(ctx context.Context, catID uuid.UUID) (*repogoal.CategoryRow, error)
	LockCategory(ctx context.Context, catID uuid.UUID) (*repogoal.CategoryRow, error)
}

// GoalRepository defines the storage contract for goals.
type GoalRepository interface {
	CreateGoal(ctx context.Context, catID uuid.UUID, name, description, unit string, weight, targetValue float64) (*repogoal.GoalRow, error)
	GetGoal(ctx context.Context, goalID uuid.UUID) (*repogoal.GoalRow, error)
	UpdateGoal(ctx context.Context, goalID uuid.UUID, name, description, unit string, weight, targetValue float64, expectedVersion int) (*repogoal.GoalRow, error)
	DeleteGoal(ctx context.Context, goalID uuid.UUID) error
	UpdateGoalCurrentValue(ctx context.Context, goalID uuid.UUID, currentValue float64) (*repogoal.GoalRow, error)
	ListGoalsByCategory(ctx context.Context, catID uuid.UUID) ([]*repogoal.GoalRow, error)
}

// KPIRepository defines the storage contract for KPIs.
type KPIRepository interface {
	ListKPIs(ctx context.Context) ([]*repogoal.KpiRow, error)
	GetKPI(ctx context.Context, kpiID uuid.UUID) (*repogoal.KpiRow, error)
	CreateKPI(ctx context.Context, name, unit, description string) (*repogoal.KpiRow, error)
	UpdateKPI(ctx context.Context, kpiID uuid.UUID, name, unit, description string) (*repogoal.KpiRow, error)
	DeleteKPI(ctx context.Context, kpiID uuid.UUID) error
	CountGoalLinksByKPI(ctx context.Context, kpiID uuid.UUID) (int, error)
}

// LinkKPIRepository defines the storage contract for KPI links.
type LinkKPIRepository interface {
	LinkKPI(ctx context.Context, goalID, kpiID uuid.UUID) error
	UnlinkKPI(ctx context.Context, goalID, kpiID uuid.UUID) error
	CountGoalKPILinks(ctx context.Context, goalID uuid.UUID) (int, error)
	ReplaceGoalKpiLinks(ctx context.Context, goalID uuid.UUID, kpiIDs []uuid.UUID) error
	ListKpiIDsByGoal(ctx context.Context, goalID uuid.UUID) ([]uuid.UUID, error)
}

// AssignmentRepository defines the storage contract for assignments.
type AssignmentRepository interface {
	GetAssignment(ctx context.Context, empID uuid.UUID) (*repogoal.AssignmentRow, error)
	CreateAssignment(ctx context.Context, empID, cycleID uuid.UUID) (*repogoal.AssignmentRow, error)
}

// WeightQuerier defines aggregate weight queries.
type WeightQuerier interface {
	SumGoalWeightsByCategoryID(ctx context.Context, catID uuid.UUID) (float64, error)
	SumCategoryWeightsByEmployee(ctx context.Context, empID uuid.UUID) (float64, error)
}

// ---------------------------------------------------------------------------
// Service interfaces consumed by handlers.
// ---------------------------------------------------------------------------

// CategoryServicer handles category business logic.
type CategoryServicer interface {
	ListCategories(ctx context.Context, empID uuid.UUID) ([]*repogoal.CategoryRow, error)
	CreateCategory(ctx context.Context, empID uuid.UUID, req dtogoal.CreateCategoryRequest) (*repogoal.CategoryRow, error)
	UpdateCategory(ctx context.Context, empID, catID uuid.UUID, req dtogoal.UpdateCategoryRequest) (*repogoal.CategoryRow, error)
	DeleteCategory(ctx context.Context, empID, catID uuid.UUID) error
}

// GoalServicer handles goal business logic.
type GoalServicer interface {
	CreateGoal(ctx context.Context, empID, catID uuid.UUID, req dtogoal.CreateGoalRequest) (*repogoal.GoalRow, error)
	UpdateGoal(ctx context.Context, empID, goalID uuid.UUID, req dtogoal.UpdateGoalRequest) (*repogoal.GoalRow, error)
	DeleteGoal(ctx context.Context, empID, goalID uuid.UUID) error
}

// ProgressServicer handles progress update business logic.
type ProgressServicer interface {
	UpdateGoalProgress(ctx context.Context, empID, goalID uuid.UUID, req dtogoal.UpdateProgressRequest) (*repogoal.GoalRow, error)
}

// KpiServicer handles KPI business logic.
type KpiServicer interface {
	ListKPIs(ctx context.Context) ([]*repogoal.KpiRow, error)
	CreateKPI(ctx context.Context, req dtogoal.CreateKpiRequest) (*repogoal.KpiRow, error)
	UpdateKPI(ctx context.Context, kpiID uuid.UUID, req dtogoal.UpdateKpiRequest) (*repogoal.KpiRow, error)
	DeleteKPI(ctx context.Context, kpiID uuid.UUID) error
	LinkKPI(ctx context.Context, empID, goalID, kpiID uuid.UUID) error
	UnlinkKPI(ctx context.Context, empID, goalID, kpiID uuid.UUID) error
}

// WeightValidationServicer handles weight validation business logic.
type WeightValidationServicer interface {
	ValidateDoubleWeighting(ctx context.Context, empID uuid.UUID) (*dtogoal.WeightValidationResponse, error)
}

// BatchServicer handles batch operations.
type BatchServicer interface {
	BatchCreateUpdateGoals(ctx context.Context, empID uuid.UUID, req dtogoal.BatchGoalRequest) ([]*repogoal.GoalRow, error)
}
