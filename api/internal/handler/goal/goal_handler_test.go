package goal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dtogoal "github.com/sed-evaluacion-desempeno/api/internal/dto/goal"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
)

// ---------------------------------------------------------------------------
// Happy path (17 endpoints)
// ---------------------------------------------------------------------------

func TestListCategories_Success(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	catSvc := &mockCategoryService{
		listFunc: func(ctx context.Context, id uuid.UUID) ([]*repogoal.CategoryRow, error) {
			require.Equal(t, empID, id)
			return []*repogoal.CategoryRow{
				{ID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), EmployeeID: empID, Name: "Q1", Weight: 50},
				{ID: uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"), EmployeeID: empID, Name: "Q2", Weight: 50},
			}, nil
		},
	}

	h := newTestHandler(catSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Get("/employees/{empId}/categories", h.ListCategories)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/employees/%s/categories", empID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp dtogoal.CategoryListResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Items, 2)
	assert.Equal(t, "Q1", resp.Items[0].Name)
	assert.Equal(t, "Q2", resp.Items[1].Name)
}

func TestCreateCategory_Success(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	catID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	catSvc := &mockCategoryService{
		createFunc: func(ctx context.Context, id uuid.UUID, req dtogoal.CreateCategoryRequest) (*repogoal.CategoryRow, error) {
			require.Equal(t, empID, id)
			require.Equal(t, "Nueva categoría", req.Name)
			return &repogoal.CategoryRow{ID: catID, EmployeeID: empID, Name: req.Name, Weight: req.Weight, CreatedAt: fixedTime(), UpdatedAt: fixedTime()}, nil
		},
	}

	h := newTestHandler(catSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Post("/employees/{empId}/categories", h.CreateCategory)

	body := `{"name":"Nueva categoría","weight":25}`
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/employees/%s/categories", empID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp dtogoal.CategoryResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, catID.String(), resp.ID)
	assert.Equal(t, "Nueva categoría", resp.Name)
}

func TestUpdateCategory_Success(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	catID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	catSvc := &mockCategoryService{
		updateFunc: func(ctx context.Context, eid, cid uuid.UUID, req dtogoal.UpdateCategoryRequest) (*repogoal.CategoryRow, error) {
			require.Equal(t, empID, eid)
			require.Equal(t, catID, cid)
			return &repogoal.CategoryRow{ID: catID, EmployeeID: empID, Name: req.Name, Weight: req.Weight, CreatedAt: fixedTime(), UpdatedAt: fixedTime()}, nil
		},
	}

	h := newTestHandler(catSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Put("/employees/{empId}/categories/{catId}", h.UpdateCategory)

	body := `{"name":"Updated","weight":30}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/employees/%s/categories/%s", empID, catID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp dtogoal.CategoryResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Updated", resp.Name)
}

func TestDeleteCategory_Success(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	catID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	catSvc := &mockCategoryService{
		deleteFunc: func(ctx context.Context, eid, cid uuid.UUID) error {
			require.Equal(t, empID, eid)
			require.Equal(t, catID, cid)
			return nil
		},
	}

	h := newTestHandler(catSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Delete("/employees/{empId}/categories/{catId}", h.DeleteCategory)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/employees/%s/categories/%s", empID, catID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestCreateGoal_Success(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	catID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	goalID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	goalSvc := &mockGoalService{
		createFunc: func(ctx context.Context, eid, cid uuid.UUID, req dtogoal.CreateGoalRequest) (*repogoal.GoalRow, error) {
			require.Equal(t, empID, eid)
			require.Equal(t, catID, cid)
			return &repogoal.GoalRow{ID: goalID, CategoryID: catID, Name: req.Name, Weight: req.Weight, Unit: req.Unit, TargetValue: req.TargetValue, State: "borrador", Version: 1, CreatedAt: fixedTime(), UpdatedAt: fixedTime()}, nil
		},
	}

	h := newTestHandler(nil, goalSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Post("/employees/{empId}/categories/{catId}/goals", h.CreateGoal)

	body := `{"name":"Aumentar ventas","unit":"porcentaje","weight":50,"target_value":100}`
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/employees/%s/categories/%s/goals", empID, catID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp dtogoal.GoalResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, goalID.String(), resp.ID)
	assert.Equal(t, "Aumentar ventas", resp.Name)
}

func TestUpdateGoal_Success(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	goalID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	goalSvc := &mockGoalService{
		updateFunc: func(ctx context.Context, eid, gid uuid.UUID, req dtogoal.UpdateGoalRequest) (*repogoal.GoalRow, error) {
			require.Equal(t, empID, eid)
			require.Equal(t, goalID, gid)
			return &repogoal.GoalRow{ID: goalID, CategoryID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: req.Name, Weight: req.Weight, Unit: req.Unit, TargetValue: req.TargetValue, Version: req.Version + 1, CreatedAt: fixedTime(), UpdatedAt: fixedTime()}, nil
		},
	}

	h := newTestHandler(nil, goalSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Put("/goals/{goalId}", h.UpdateGoal)

	body := `{"name":"Updated goal","unit":"porcentaje","weight":40,"target_value":90,"version":1}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/goals/%s", goalID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp dtogoal.GoalResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Updated goal", resp.Name)
}

func TestDeleteGoal_Success(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	goalID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	goalSvc := &mockGoalService{
		deleteFunc: func(ctx context.Context, eid, gid uuid.UUID) error {
			require.Equal(t, empID, eid)
			require.Equal(t, goalID, gid)
			return nil
		},
	}

	h := newTestHandler(nil, goalSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Delete("/goals/{goalId}", h.DeleteGoal)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/goals/%s", goalID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestUpdateGoalProgress_Success(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	goalID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	progSvc := &mockProgressService{
		updateFunc: func(ctx context.Context, eid, gid uuid.UUID, req dtogoal.UpdateProgressRequest) (*repogoal.GoalRow, error) {
			require.Equal(t, empID, eid)
			require.Equal(t, goalID, gid)
			return &repogoal.GoalRow{ID: goalID, CategoryID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "Ventas", CurrentValue: req.CurrentValue, Version: 2, CreatedAt: fixedTime(), UpdatedAt: fixedTime()}, nil
		},
	}

	h := newTestHandler(nil, nil, progSvc, nil, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Patch("/goals/{goalId}/progress", h.UpdateGoalProgress)

	body := `{"current_value":75.5}`
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/goals/%s/progress", goalID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp dtogoal.GoalResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.InDelta(t, 75.5, resp.CurrentValue, 0.001)
}

func TestValidateWeights_Success(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	weightSvc := &mockWeightValidationService{
		validateFunc: func(ctx context.Context, id uuid.UUID) (*dtogoal.WeightValidationResponse, error) {
			require.Equal(t, empID, id)
			return &dtogoal.WeightValidationResponse{
				Valid:       true,
				CategorySum: 100.0,
				ExpectedSum: 100.0,
				Deficit:     0,
				GoalSums: []dtogoal.CategoryGoalSum{
					{CategoryID: "cat-1", CategoryName: "Q1", Sum: 100.0, ExpectedSum: 100.0, Deficit: 0},
				},
			}, nil
		},
	}

	h := newTestHandler(nil, nil, nil, nil, weightSvc, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Post("/employees/{empId}/validate-weights", h.ValidateWeights)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/employees/%s/validate-weights", empID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp dtogoal.WeightValidationResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Valid)
	assert.InDelta(t, 100.0, resp.CategorySum, 0.001)
}

func TestListKPIs_Success(t *testing.T) {
	kpiSvc := &mockKPIService{
		listFunc: func(ctx context.Context) ([]*repogoal.KpiRow, error) {
			return []*repogoal.KpiRow{
				{ID: uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"), Name: "KPI-1", Unit: "numero", CreatedAt: fixedTime(), UpdatedAt: fixedTime()},
				{ID: uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"), Name: "KPI-2", Unit: "porcentaje", CreatedAt: fixedTime(), UpdatedAt: fixedTime()},
			}, nil
		},
	}

	h := newTestHandler(nil, nil, nil, kpiSvc, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Get("/kpis", h.ListKPIs)

	req := httptest.NewRequest(http.MethodGet, "/kpis", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp dtogoal.KpiListResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Items, 2)
	assert.Equal(t, "KPI-1", resp.Items[0].Name)
}

func TestCreateKPI_Success(t *testing.T) {
	kpiID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	kpiSvc := &mockKPIService{
		createFunc: func(ctx context.Context, req dtogoal.CreateKpiRequest) (*repogoal.KpiRow, error) {
			return &repogoal.KpiRow{ID: kpiID, Name: req.Name, Unit: req.Unit, Description: req.Description, CreatedAt: fixedTime(), UpdatedAt: fixedTime()}, nil
		},
	}

	h := newTestHandler(nil, nil, nil, kpiSvc, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Post("/kpis", h.CreateKPI)

	body := `{"name":"NPS","unit":"numero","description":"Net Promoter Score"}`
	req := httptest.NewRequest(http.MethodPost, "/kpis", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp dtogoal.KpiResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "NPS", resp.Name)
}

func TestUpdateKPI_Success(t *testing.T) {
	kpiID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	kpiSvc := &mockKPIService{
		updateFunc: func(ctx context.Context, id uuid.UUID, req dtogoal.UpdateKpiRequest) (*repogoal.KpiRow, error) {
			require.Equal(t, kpiID, id)
			return &repogoal.KpiRow{ID: kpiID, Name: req.Name, Unit: req.Unit, CreatedAt: fixedTime(), UpdatedAt: fixedTime()}, nil
		},
	}

	h := newTestHandler(nil, nil, nil, kpiSvc, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Put("/kpis/{kpiId}", h.UpdateKPI)

	body := `{"name":"NPS v2","unit":"porcentaje"}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/kpis/%s", kpiID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp dtogoal.KpiResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "NPS v2", resp.Name)
}

func TestLinkKPI_Success(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	goalID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	kpiID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")

	kpiSvc := &mockKPIService{
		linkFunc: func(ctx context.Context, eid, gid, kid uuid.UUID) error {
			require.Equal(t, empID, eid)
			require.Equal(t, goalID, gid)
			require.Equal(t, kpiID, kid)
			return nil
		},
	}

	goalRepo := &mockGoalRepo{
		getFunc: func(ctx context.Context, id uuid.UUID) (*repogoal.GoalRow, error) {
			return &repogoal.GoalRow{ID: goalID, CategoryID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "Ventas", Version: 1, CreatedAt: fixedTime(), UpdatedAt: fixedTime()}, nil
		},
	}

	h := newTestHandler(nil, nil, nil, kpiSvc, nil, nil, nil, goalRepo, nil, nil, nil)
	r := chi.NewRouter()
	r.Post("/goals/{goalId}/kpis", h.LinkKPI)

	body := fmt.Sprintf(`{"kpi_id":"%s"}`, kpiID)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/goals/%s/kpis", goalID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp dtogoal.GoalResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, goalID.String(), resp.ID)
}

func TestUnlinkKPI_Success(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	goalID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	kpiID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")

	kpiSvc := &mockKPIService{
		unlinkFunc: func(ctx context.Context, eid, gid, kid uuid.UUID) error {
			require.Equal(t, empID, eid)
			require.Equal(t, goalID, gid)
			require.Equal(t, kpiID, kid)
			return nil
		},
	}

	h := newTestHandler(nil, nil, nil, kpiSvc, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Delete("/goals/{goalId}/kpis/{kpiId}", h.UnlinkKPI)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/goals/%s/kpis/%s", goalID, kpiID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestGetAssignment_Success(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	assignID := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
	cycleID := uuid.MustParse("99999999-9999-9999-9999-999999999999")

	assignRepo := &mockAssignmentRepo{
		getFunc: func(ctx context.Context, id uuid.UUID) (*repogoal.AssignmentRow, error) {
			require.Equal(t, empID, id)
			return &repogoal.AssignmentRow{ID: assignID, EmployeeID: empID, CycleID: cycleID, CreatedAt: fixedTime()}, nil
		},
	}

	catRepo := &mockCategoryRepo{
		listByEmployeeFunc: func(ctx context.Context, id uuid.UUID) ([]*repogoal.CategoryRow, error) {
			return []*repogoal.CategoryRow{
				{ID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), EmployeeID: empID, Name: "Q1", Weight: 100, CreatedAt: fixedTime(), UpdatedAt: fixedTime()},
			}, nil
		},
	}

	goalRepo := &mockGoalRepo{
		listByCategoryFunc: func(ctx context.Context, catID uuid.UUID) ([]*repogoal.GoalRow, error) {
			return []*repogoal.GoalRow{
				{ID: uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"), CategoryID: catID, Name: "Meta 1", Weight: 100, Version: 1, CreatedAt: fixedTime(), UpdatedAt: fixedTime()},
			}, nil
		},
	}

	h := newTestHandler(nil, nil, nil, nil, nil, nil, catRepo, goalRepo, nil, nil, assignRepo)
	r := chi.NewRouter()
	r.Get("/employees/{empId}/assignments", h.GetAssignment)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/employees/%s/assignments", empID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp dtogoal.AssignmentResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, assignID.String(), resp.ID)
	require.Len(t, resp.Categories, 1)
	require.Len(t, resp.Categories[0].Goals, 1)
	assert.Equal(t, "Meta 1", resp.Categories[0].Goals[0].Name)
}

func TestCreateAssignment_Success(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	cycleID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	assignID := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")

	assignRepo := &mockAssignmentRepo{
		createFunc: func(ctx context.Context, eid, cid uuid.UUID) (*repogoal.AssignmentRow, error) {
			require.Equal(t, empID, eid)
			require.Equal(t, cycleID, cid)
			return &repogoal.AssignmentRow{ID: assignID, EmployeeID: empID, CycleID: cycleID, CreatedAt: fixedTime()}, nil
		},
	}

	catRepo := &mockCategoryRepo{
		listByEmployeeFunc: func(ctx context.Context, id uuid.UUID) ([]*repogoal.CategoryRow, error) {
			return nil, nil
		},
	}

	h := newTestHandler(nil, nil, nil, nil, nil, nil, catRepo, nil, nil, nil, assignRepo)
	r := chi.NewRouter()
	r.Post("/employees/{empId}/assignments", h.CreateAssignment)

	body := fmt.Sprintf(`{"cycle_id":"%s"}`, cycleID)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/employees/%s/assignments", empID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp dtogoal.AssignmentResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, assignID.String(), resp.ID)
}

func TestBatchGoals_Success(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	goalID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")

	batchSvc := &mockBatchService{
		batchFunc: func(ctx context.Context, id uuid.UUID, req dtogoal.BatchGoalRequest) ([]*repogoal.GoalRow, error) {
			require.Equal(t, empID, id)
			return []*repogoal.GoalRow{
				{ID: goalID, CategoryID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "Batch goal", Weight: 50, Version: 1, CreatedAt: fixedTime(), UpdatedAt: fixedTime()},
			}, nil
		},
	}

	h := newTestHandler(nil, nil, nil, nil, nil, batchSvc, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Post("/goals/batch", h.BatchGoals)

	body := `{"items":[{"operation":"create","category_id":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","goal":{"name":"Batch goal","unit":"porcentaje","weight":50,"target_value":100}}]}`
	req := httptest.NewRequest(http.MethodPost, "/goals/batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp dtogoal.BatchGoalResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "Batch goal", resp.Items[0].Name)
}

// ---------------------------------------------------------------------------
// Error cases
// ---------------------------------------------------------------------------

func TestCreateCategory_DuplicateName(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	catSvc := &mockCategoryService{
		createFunc: func(ctx context.Context, id uuid.UUID, req dtogoal.CreateCategoryRequest) (*repogoal.CategoryRow, error) {
			return nil, pkgerrors.ErrDuplicateCategoryName
		},
	}

	h := newTestHandler(catSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Post("/employees/{empId}/categories", h.CreateCategory)

	body := `{"name":"Duplicada","weight":30}`
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/employees/%s/categories", empID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "DUPLICATE_CATEGORY_NAME")
}

func TestCreateGoal_WeightOverflow(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	catID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	goalSvc := &mockGoalService{
		createFunc: func(ctx context.Context, eid, cid uuid.UUID, req dtogoal.CreateGoalRequest) (*repogoal.GoalRow, error) {
			return nil, pkgerrors.ErrGoalWeightOverflow
		},
	}

	h := newTestHandler(nil, goalSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Post("/employees/{empId}/categories/{catId}/goals", h.CreateGoal)

	body := `{"name":"Overflow","unit":"porcentaje","weight":200,"target_value":100}`
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/employees/%s/categories/%s/goals", empID, catID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "GOAL_WEIGHT_OVERFLOW")
}

func TestDeleteGoal_PhaseRestricted(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	_ = empID
	goalID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	goalSvc := &mockGoalService{
		deleteFunc: func(ctx context.Context, eid, gid uuid.UUID) error {
			return pkgerrors.ErrPhaseRestricted
		},
	}

	h := newTestHandler(nil, goalSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Delete("/goals/{goalId}", h.DeleteGoal)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/goals/%s", goalID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "PHASE_RESTRICTED")
}

func TestUpdateProgress_WrongPhase(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	_ = empID
	goalID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	progSvc := &mockProgressService{
		updateFunc: func(ctx context.Context, eid, gid uuid.UUID, req dtogoal.UpdateProgressRequest) (*repogoal.GoalRow, error) {
			return nil, pkgerrors.ErrPhaseRestricted
		},
	}

	h := newTestHandler(nil, nil, progSvc, nil, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Patch("/goals/{goalId}/progress", h.UpdateGoalProgress)

	body := `{"current_value":50}`
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/goals/%s/progress", goalID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "PHASE_RESTRICTED")
}

func TestValidateWeights_InvalidSum(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	weightSvc := &mockWeightValidationService{
		validateFunc: func(ctx context.Context, id uuid.UUID) (*dtogoal.WeightValidationResponse, error) {
			return &dtogoal.WeightValidationResponse{
				Valid:       false,
				CategorySum: 90.0,
				ExpectedSum: 100.0,
				Deficit:     10.0,
				GoalSums: []dtogoal.CategoryGoalSum{
					{CategoryID: "cat-1", CategoryName: "Q1", Sum: 90.0, ExpectedSum: 100.0, Deficit: 10.0},
				},
			}, nil
		},
	}

	h := newTestHandler(nil, nil, nil, nil, weightSvc, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Post("/employees/{empId}/validate-weights", h.ValidateWeights)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/employees/%s/validate-weights", empID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp dtogoal.WeightValidationResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Valid)
	assert.InDelta(t, 90.0, resp.CategorySum, 0.001)
}

func TestCreateGoal_NotFound(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	catID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	goalSvc := &mockGoalService{
		createFunc: func(ctx context.Context, eid, cid uuid.UUID, req dtogoal.CreateGoalRequest) (*repogoal.GoalRow, error) {
			return nil, pkgerrors.ErrCategoryNotFound
		},
	}

	h := newTestHandler(nil, goalSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Post("/employees/{empId}/categories/{catId}/goals", h.CreateGoal)

	body := `{"name":"Missing","unit":"porcentaje","weight":50,"target_value":100}`
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/employees/%s/categories/%s/goals", empID, catID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "CATEGORY_NOT_FOUND")
}

// ---------------------------------------------------------------------------
// Response time benchmarks
// ---------------------------------------------------------------------------

func TestListCategories_ResponseTime(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	catSvc := &mockCategoryService{
		listFunc: func(ctx context.Context, id uuid.UUID) ([]*repogoal.CategoryRow, error) {
			return []*repogoal.CategoryRow{
				{ID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), EmployeeID: empID, Name: "Q1", Weight: 50},
			}, nil
		},
	}

	h := newTestHandler(catSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Get("/employees/{empId}/categories", h.ListCategories)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/employees/%s/categories", empID), nil)
	w := httptest.NewRecorder()

	start := time.Now()
	r.ServeHTTP(w, req)
	elapsed := time.Since(start)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Less(t, elapsed, 200*time.Millisecond, "ListCategories took too long")
}

func TestValidateWeights_ResponseTime(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	weightSvc := &mockWeightValidationService{
		validateFunc: func(ctx context.Context, id uuid.UUID) (*dtogoal.WeightValidationResponse, error) {
			return &dtogoal.WeightValidationResponse{Valid: true, CategorySum: 100.0, ExpectedSum: 100.0, Deficit: 0}, nil
		},
	}

	h := newTestHandler(nil, nil, nil, nil, weightSvc, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Post("/employees/{empId}/validate-weights", h.ValidateWeights)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/employees/%s/validate-weights", empID), nil)
	w := httptest.NewRecorder()

	start := time.Now()
	r.ServeHTTP(w, req)
	elapsed := time.Since(start)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Less(t, elapsed, 150*time.Millisecond, "ValidateWeights took too long")
}

// ---------------------------------------------------------------------------
// Concurrency
// ---------------------------------------------------------------------------

func TestCreateGoal_ConcurrentWeightOverflow(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	catID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	var callCount int64
	goalSvc := &mockGoalService{
		createFunc: func(ctx context.Context, eid, cid uuid.UUID, req dtogoal.CreateGoalRequest) (*repogoal.GoalRow, error) {
			atomic.AddInt64(&callCount, 1)
			// Simulate weight overflow for all concurrent requests
			return nil, pkgerrors.ErrGoalWeightOverflow
		},
	}

	h := newTestHandler(nil, goalSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Post("/employees/{empId}/categories/{catId}/goals", h.CreateGoal)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body := `{"name":"Concurrent","unit":"porcentaje","weight":10,"target_value":100}`
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/employees/%s/categories/%s/goals", empID, catID), strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		}()
	}
	wg.Wait()

	assert.Equal(t, int64(100), atomic.LoadInt64(&callCount))
}

func TestBatchGoals_Concurrent(t *testing.T) {
	empID := mustParseUUID("11111111-1111-1111-1111-111111111111")
	_ = empID

	var callCount int64
	batchSvc := &mockBatchService{
		batchFunc: func(ctx context.Context, id uuid.UUID, req dtogoal.BatchGoalRequest) ([]*repogoal.GoalRow, error) {
			atomic.AddInt64(&callCount, 1)
			return []*repogoal.GoalRow{}, nil
		},
	}

	h := newTestHandler(nil, nil, nil, nil, nil, batchSvc, nil, nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Post("/goals/batch", h.BatchGoals)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body := `{"items":[{"operation":"create","category_id":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","goal":{"name":"Batch","unit":"porcentaje","weight":100,"target_value":100}}]}`
			req := httptest.NewRequest(http.MethodPost, "/goals/batch", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}()
	}
	wg.Wait()

	assert.Equal(t, int64(50), atomic.LoadInt64(&callCount))
}
