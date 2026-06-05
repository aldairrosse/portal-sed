package goal

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
)

// mockCategoryRepoForValidation is a minimal CategoryRepository mock.
type mockCategoryRepoForValidation struct {
	cats []*repogoal.CategoryRow
	err  error
}

func (m *mockCategoryRepoForValidation) ListCategoriesByEmployee(ctx context.Context, empID uuid.UUID) ([]*repogoal.CategoryRow, error) {
	return m.cats, m.err
}

func (m *mockCategoryRepoForValidation) CreateCategory(ctx context.Context, empID uuid.UUID, name, description string, weight float64) (*repogoal.CategoryRow, error) {
	return nil, nil
}

func (m *mockCategoryRepoForValidation) UpdateCategory(ctx context.Context, catID uuid.UUID, name, description string, weight float64) (*repogoal.CategoryRow, error) {
	return nil, nil
}

func (m *mockCategoryRepoForValidation) DeleteCategory(ctx context.Context, catID uuid.UUID) error {
	return nil
}

func (m *mockCategoryRepoForValidation) GetCategory(ctx context.Context, catID uuid.UUID) (*repogoal.CategoryRow, error) {
	return nil, nil
}

func (m *mockCategoryRepoForValidation) LockCategory(ctx context.Context, catID uuid.UUID) (*repogoal.CategoryRow, error) {
	return nil, nil
}

// mockGoalRepoForValidation is a minimal GoalRepository mock.
type mockGoalRepoForValidation struct {
	goals map[uuid.UUID][]*repogoal.GoalRow
	err   error
}

func (m *mockGoalRepoForValidation) CreateGoal(ctx context.Context, catID uuid.UUID, name, description, unit string, weight, targetValue float64) (*repogoal.GoalRow, error) {
	return nil, nil
}

func (m *mockGoalRepoForValidation) GetGoal(ctx context.Context, goalID uuid.UUID) (*repogoal.GoalRow, error) {
	return nil, nil
}

func (m *mockGoalRepoForValidation) UpdateGoal(ctx context.Context, goalID uuid.UUID, name, description, unit string, weight, targetValue float64, expectedVersion int) (*repogoal.GoalRow, error) {
	return nil, nil
}

func (m *mockGoalRepoForValidation) DeleteGoal(ctx context.Context, goalID uuid.UUID) error {
	return nil
}

func (m *mockGoalRepoForValidation) UpdateGoalCurrentValue(ctx context.Context, goalID uuid.UUID, currentValue float64) (*repogoal.GoalRow, error) {
	return nil, nil
}

func (m *mockGoalRepoForValidation) ListGoalsByCategory(ctx context.Context, catID uuid.UUID) ([]*repogoal.GoalRow, error) {
	if m.goals == nil {
		return nil, m.err
	}
	return m.goals[catID], m.err
}

func newValidationService(cats []*repogoal.CategoryRow, goals map[uuid.UUID][]*repogoal.GoalRow) *WeightValidationService {
	catRepo := &mockCategoryRepoForValidation{cats: cats}
	goalRepo := &mockGoalRepoForValidation{goals: goals}
	return NewWeightValidationService(catRepo, goalRepo)
}

func TestValidateDoubleWeighting_Perfect100(t *testing.T) {
	empID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	catID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	cats := []*repogoal.CategoryRow{
		{ID: catID, EmployeeID: empID, Name: "Q1", Weight: 100.0},
	}
	goals := map[uuid.UUID][]*repogoal.GoalRow{
		catID: {
			{ID: uuid.MustParse("11111111-1111-1111-1111-111111111112"), CategoryID: catID, Name: "G1", Weight: 40.0},
			{ID: uuid.MustParse("11111111-1111-1111-1111-111111111113"), CategoryID: catID, Name: "G2", Weight: 60.0},
		},
	}

	svc := newValidationService(cats, goals)
	resp, err := svc.ValidateDoubleWeighting(context.Background(), empID)
	require.NoError(t, err)
	assert.True(t, resp.Valid, "expected valid when both category and goals sum to 100")
	assert.InDelta(t, 100.0, resp.CategorySum, 0.001)
	assert.Len(t, resp.GoalSums, 1)
	assert.InDelta(t, 100.0, resp.GoalSums[0].Sum, 0.001)
}

func TestValidateDoubleWeighting_Under100(t *testing.T) {
	empID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	catID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	cats := []*repogoal.CategoryRow{
		{ID: catID, EmployeeID: empID, Name: "Q1", Weight: 80.0},
	}
	goals := map[uuid.UUID][]*repogoal.GoalRow{
		catID: {
			{ID: uuid.MustParse("11111111-1111-1111-1111-111111111112"), CategoryID: catID, Name: "G1", Weight: 40.0},
			{ID: uuid.MustParse("11111111-1111-1111-1111-111111111113"), CategoryID: catID, Name: "G2", Weight: 50.0},
		},
	}

	svc := newValidationService(cats, goals)
	resp, err := svc.ValidateDoubleWeighting(context.Background(), empID)
	require.NoError(t, err)
	assert.False(t, resp.Valid)
	assert.InDelta(t, 80.0, resp.CategorySum, 0.001)
	assert.InDelta(t, 90.0, resp.GoalSums[0].Sum, 0.001)
}

func TestValidateDoubleWeighting_Over100(t *testing.T) {
	empID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	catID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	cats := []*repogoal.CategoryRow{
		{ID: catID, EmployeeID: empID, Name: "Q1", Weight: 110.0},
	}
	goals := map[uuid.UUID][]*repogoal.GoalRow{
		catID: {
			{ID: uuid.MustParse("11111111-1111-1111-1111-111111111112"), CategoryID: catID, Name: "G1", Weight: 60.0},
			{ID: uuid.MustParse("11111111-1111-1111-1111-111111111113"), CategoryID: catID, Name: "G2", Weight: 50.0},
		},
	}

	svc := newValidationService(cats, goals)
	resp, err := svc.ValidateDoubleWeighting(context.Background(), empID)
	require.NoError(t, err)
	assert.False(t, resp.Valid)
	assert.InDelta(t, 110.0, resp.CategorySum, 0.001)
	assert.InDelta(t, 110.0, resp.GoalSums[0].Sum, 0.001)
}

func TestValidateDoubleWeighting_Tolerance9999(t *testing.T) {
	// 99.99 should be within Epsilon (0.01) of 100.0
	empID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	catID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	cats := []*repogoal.CategoryRow{
		{ID: catID, EmployeeID: empID, Name: "Q1", Weight: 99.99},
	}
	goals := map[uuid.UUID][]*repogoal.GoalRow{
		catID: {
			{ID: uuid.MustParse("11111111-1111-1111-1111-111111111112"), CategoryID: catID, Name: "G1", Weight: 99.99},
		},
	}

	svc := newValidationService(cats, goals)
	resp, err := svc.ValidateDoubleWeighting(context.Background(), empID)
	require.NoError(t, err)
	assert.True(t, resp.Valid, "99.99 should be within Epsilon of 100")
}

func TestValidateDoubleWeighting_Tolerance10001(t *testing.T) {
	// 100.01 should be within Epsilon (0.01) of 100.0
	empID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	catID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	cats := []*repogoal.CategoryRow{
		{ID: catID, EmployeeID: empID, Name: "Q1", Weight: 100.01},
	}
	goals := map[uuid.UUID][]*repogoal.GoalRow{
		catID: {
			{ID: uuid.MustParse("11111111-1111-1111-1111-111111111112"), CategoryID: catID, Name: "G1", Weight: 100.01},
		},
	}

	svc := newValidationService(cats, goals)
	resp, err := svc.ValidateDoubleWeighting(context.Background(), empID)
	require.NoError(t, err)
	assert.True(t, resp.Valid, "100.01 should be within Epsilon of 100")
}

func TestValidateDoubleWeighting_EmptyCategories(t *testing.T) {
	empID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	svc := newValidationService([]*repogoal.CategoryRow{}, nil)
	resp, err := svc.ValidateDoubleWeighting(context.Background(), empID)
	require.NoError(t, err)
	assert.False(t, resp.Valid)
	assert.InDelta(t, 0.0, resp.CategorySum, 0.001)
	assert.InDelta(t, 100.0, resp.Deficit, 0.001)
	assert.Empty(t, resp.GoalSums)
}

func TestValidateDoubleWeighting_SingleGoalPerCategory(t *testing.T) {
	empID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	catID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	cats := []*repogoal.CategoryRow{
		{ID: catID, EmployeeID: empID, Name: "Q1", Weight: 100.0},
	}
	goals := map[uuid.UUID][]*repogoal.GoalRow{
		catID: {
			{ID: uuid.MustParse("11111111-1111-1111-1111-111111111112"), CategoryID: catID, Name: "G1", Weight: 100.0},
		},
	}

	svc := newValidationService(cats, goals)
	resp, err := svc.ValidateDoubleWeighting(context.Background(), empID)
	require.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.Len(t, resp.GoalSums, 1)
	assert.InDelta(t, 100.0, resp.GoalSums[0].Sum, 0.001)
}

func TestValidateDoubleWeighting_MultipleGoalsPerCategory(t *testing.T) {
	empID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	catID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	cats := []*repogoal.CategoryRow{
		{ID: catID, EmployeeID: empID, Name: "Q1", Weight: 100.0},
	}
	goals := map[uuid.UUID][]*repogoal.GoalRow{
		catID: {
			{ID: uuid.MustParse("11111111-1111-1111-1111-111111111112"), CategoryID: catID, Name: "G1", Weight: 25.0},
			{ID: uuid.MustParse("11111111-1111-1111-1111-111111111113"), CategoryID: catID, Name: "G2", Weight: 25.0},
			{ID: uuid.MustParse("11111111-1111-1111-1111-111111111114"), CategoryID: catID, Name: "G3", Weight: 25.0},
			{ID: uuid.MustParse("11111111-1111-1111-1111-111111111115"), CategoryID: catID, Name: "G4", Weight: 25.0},
		},
	}

	svc := newValidationService(cats, goals)
	resp, err := svc.ValidateDoubleWeighting(context.Background(), empID)
	require.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.Len(t, resp.GoalSums, 1)
	assert.InDelta(t, 100.0, resp.GoalSums[0].Sum, 0.001)
}

func TestValidateDoubleWeighting_FloatingPointPrecision(t *testing.T) {
	// 33.333333 + 33.333333 + 33.333334 = 100.0 exactly in float64? Let's check.
	empID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	catID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	cats := []*repogoal.CategoryRow{
		{ID: catID, EmployeeID: empID, Name: "Q1", Weight: 100.0},
	}
	goals := map[uuid.UUID][]*repogoal.GoalRow{
		catID: {
			{ID: uuid.MustParse("11111111-1111-1111-1111-111111111112"), CategoryID: catID, Name: "G1", Weight: 33.333333},
			{ID: uuid.MustParse("11111111-1111-1111-1111-111111111113"), CategoryID: catID, Name: "G2", Weight: 33.333333},
			{ID: uuid.MustParse("11111111-1111-1111-1111-111111111114"), CategoryID: catID, Name: "G3", Weight: 33.333334},
		},
	}

	svc := newValidationService(cats, goals)
	resp, err := svc.ValidateDoubleWeighting(context.Background(), empID)
	require.NoError(t, err)
	assert.True(t, resp.Valid, "sum of 33.333333+33.333333+33.333334 should be within Epsilon of 100")
	assert.InDelta(t, 100.0, resp.GoalSums[0].Sum, 0.001)
}
