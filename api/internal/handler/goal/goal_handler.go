// Package goal provides HTTP handlers for goals, categories, KPIs, and assignments.
package goal

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	dtogoal "github.com/sed-evaluacion-desempeno/api/internal/dto/goal"
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/cursor"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
	svcgoal "github.com/sed-evaluacion-desempeno/api/internal/service/goal"
	"time"
)

// generateTraceID generates a short trace ID for error responses.
func generateTraceID() string {
	id := uuid.New().String()
	return id[:8] + "-" + id[9:13]
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("handler: failed to encode JSON response: %v", err)
	}
}

// writeError writes a structured error response.
func writeError(w http.ResponseWriter, err error) {
	traceID := generateTraceID()
	status := pkgerrors.HTTPStatus(err)

	var de *pkgerrors.DomainError
	if pkgerrors.AsDomainError(err, &de) {
		writeJSON(w, status, pkgerrors.NewAPIErrorResponse(de, traceID))
		return
	}

	writeJSON(w, status, pkgerrors.NewAPIErrorResponse(
		pkgerrors.NewDomainError(pkgerrors.InvalidRequest, err.Error(), err),
		traceID,
	))
}

// GoalHandler holds all HTTP handlers for the goals bounded context.
type GoalHandler struct {
	catService   svcgoal.CategoryServicer
	goalService  svcgoal.GoalServicer
	progressSvc  svcgoal.ProgressServicer
	kpiService   svcgoal.KpiServicer
	weightSvc    svcgoal.WeightValidationServicer
	batchService svcgoal.BatchServicer
	catRepo      svcgoal.CategoryRepository
	goalRepo     svcgoal.GoalRepository
	kpiRepo      svcgoal.KPIRepository
	linkRepo     svcgoal.LinkKPIRepository
	assignRepo   svcgoal.AssignmentRepository
}

// NewGoalHandler creates a new GoalHandler.
func NewGoalHandler(
	catService svcgoal.CategoryServicer,
	goalService svcgoal.GoalServicer,
	progressSvc svcgoal.ProgressServicer,
	kpiService svcgoal.KpiServicer,
	weightSvc svcgoal.WeightValidationServicer,
	batchService svcgoal.BatchServicer,
	catRepo svcgoal.CategoryRepository,
	goalRepo svcgoal.GoalRepository,
	kpiRepo svcgoal.KPIRepository,
	linkRepo svcgoal.LinkKPIRepository,
	assignRepo svcgoal.AssignmentRepository,
) *GoalHandler {
	return &GoalHandler{
		catService:   catService,
		goalService:  goalService,
		progressSvc:  progressSvc,
		kpiService:   kpiService,
		weightSvc:    weightSvc,
		batchService: batchService,
		catRepo:      catRepo,
		goalRepo:     goalRepo,
		kpiRepo:      kpiRepo,
		linkRepo:     linkRepo,
		assignRepo:   assignRepo,
	}
}

// ============================================================================
// Category Handlers
// ============================================================================

// parseEmpID extracts and parses the employee ID from URL params.
func parseEmpID(r *http.Request) (uuid.UUID, error) {
	empIDStr := chi.URLParam(r, "empId")
	if empIDStr == "" {
		// Fall back to context
		empIDStr = middleware.EmployeeIDFromContext(r.Context())
	}
	if empIDStr == "" {
		return uuid.Nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "employee ID is required", nil)
	}
	return uuid.Parse(empIDStr)
}

// categoryRowToResponse converts a repo CategoryRow to an API response.
func categoryRowToResponse(c *repogoal.CategoryRow) dtogoal.CategoryResponse {
	return dtogoal.CategoryResponse{
		ID:          c.ID.String(),
		EmployeeID:  c.EmployeeID.String(),
		Name:        c.Name,
		Description: c.Description,
		Weight:      c.Weight,
		Goals:       nil,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
	}
}

// ListCategories handles GET /api/v1/employees/{empId}/categories.
func (h *GoalHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	empID, err := parseEmpID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	cats, err := h.catService.ListCategories(r.Context(), empID)
	if err != nil {
		writeError(w, err)
		return
	}

	items := make([]dtogoal.CategoryResponse, len(cats))
	for i, c := range cats {
		items[i] = categoryRowToResponse(c)
	}

	resp := dtogoal.CategoryListResponse{
		Items: items,
	}
	writeJSON(w, http.StatusOK, resp)
}

// CreateCategory handles POST /api/v1/employees/{empId}/categories.
func (h *GoalHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	empID, err := parseEmpID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req dtogoal.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid JSON body", err))
		return
	}

	cat, err := h.catService.CreateCategory(r.Context(), empID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, categoryRowToResponse(cat))
}

// UpdateCategory handles PUT /api/v1/employees/{empId}/categories/{catId}.
func (h *GoalHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	empID, err := parseEmpID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	catIDStr := chi.URLParam(r, "catId")
	catID, err := uuid.Parse(catIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid category ID", err))
		return
	}

	var req dtogoal.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid JSON body", err))
		return
	}

	cat, err := h.catService.UpdateCategory(r.Context(), empID, catID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, categoryRowToResponse(cat))
}

// DeleteCategory handles DELETE /api/v1/employees/{empId}/categories/{catId}.
func (h *GoalHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	empID, err := parseEmpID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	catIDStr := chi.URLParam(r, "catId")
	catID, err := uuid.Parse(catIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid category ID", err))
		return
	}

	if err := h.catService.DeleteCategory(r.Context(), empID, catID); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// Goal Handlers
// ============================================================================

// goalRowToResponse converts a repo GoalRow to an API response.
func goalRowToResponse(g *repogoal.GoalRow) dtogoal.GoalResponse {
	return dtogoal.GoalResponse{
		ID:           g.ID.String(),
		CategoryID:   g.CategoryID.String(),
		Name:         g.Name,
		Description:  g.Description,
		Unit:         g.Unit,
		Weight:       g.Weight,
		TargetValue:  g.TargetValue,
		CurrentValue: g.CurrentValue,
		State:        g.State,
		Version:      g.Version,
		CreatedAt:    g.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    g.UpdatedAt.Format(time.RFC3339),
	}
}

// CreateGoal handles POST /api/v1/employees/{empId}/categories/{catId}/goals.
func (h *GoalHandler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	empID, err := parseEmpID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	catIDStr := chi.URLParam(r, "catId")
	catID, err := uuid.Parse(catIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid category ID", err))
		return
	}

	var req dtogoal.CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid JSON body", err))
		return
	}

	goal, err := h.goalService.CreateGoal(r.Context(), empID, catID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, goalRowToResponse(goal))
}

// UpdateGoal handles PUT /api/v1/goals/{goalId}.
func (h *GoalHandler) UpdateGoal(w http.ResponseWriter, r *http.Request) {
	empID, err := parseEmpID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	goalIDStr := chi.URLParam(r, "goalId")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid goal ID", err))
		return
	}

	var req dtogoal.UpdateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid JSON body", err))
		return
	}

	goal, err := h.goalService.UpdateGoal(r.Context(), empID, goalID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, goalRowToResponse(goal))
}

// DeleteGoal handles DELETE /api/v1/goals/{goalId}.
func (h *GoalHandler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	empID, err := parseEmpID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	goalIDStr := chi.URLParam(r, "goalId")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid goal ID", err))
		return
	}

	if err := h.goalService.DeleteGoal(r.Context(), empID, goalID); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateGoalProgress handles PATCH /api/v1/goals/{goalId}/progress.
func (h *GoalHandler) UpdateGoalProgress(w http.ResponseWriter, r *http.Request) {
	empID, err := parseEmpID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	goalIDStr := chi.URLParam(r, "goalId")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid goal ID", err))
		return
	}

	var req dtogoal.UpdateProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid JSON body", err))
		return
	}

	goal, err := h.progressSvc.UpdateGoalProgress(r.Context(), empID, goalID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, goalRowToResponse(goal))
}

// ============================================================================
// Batch Handlers
// ============================================================================

// BatchGoals handles POST /api/v1/goals/batch.
func (h *GoalHandler) BatchGoals(w http.ResponseWriter, r *http.Request) {
	empID, err := parseEmpID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req dtogoal.BatchGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid JSON body", err))
		return
	}

	results, err := h.batchService.BatchCreateUpdateGoals(r.Context(), empID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	items := make([]dtogoal.GoalResponse, len(results))
	for i, g := range results {
		items[i] = goalRowToResponse(g)
	}

	writeJSON(w, http.StatusOK, dtogoal.BatchGoalResponse{Items: items})
}

// ============================================================================
// Weight Validation Handlers
// ============================================================================

// ValidateWeights handles POST /api/v1/employees/{empId}/validate-weights.
func (h *GoalHandler) ValidateWeights(w http.ResponseWriter, r *http.Request) {
	empID, err := parseEmpID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	result, err := h.weightSvc.ValidateDoubleWeighting(r.Context(), empID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ============================================================================
// KPI Handlers
// ============================================================================

// kpiRowToResponse converts a repo KpiRow to an API response.
func kpiRowToResponse(k *repogoal.KpiRow) dtogoal.KpiResponse {
	return dtogoal.KpiResponse{
		ID:          k.ID.String(),
		Name:        k.Name,
		Unit:        k.Unit,
		Description: k.Description,
		CreatedAt:   k.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   k.UpdatedAt.Format(time.RFC3339),
	}
}

// ListKPIs handles GET /api/v1/kpis.
func (h *GoalHandler) ListKPIs(w http.ResponseWriter, r *http.Request) {
	kpis, err := h.kpiService.ListKPIs(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	items := make([]dtogoal.KpiResponse, len(kpis))
	for i, k := range kpis {
		items[i] = kpiRowToResponse(k)
	}

	resp := dtogoal.KpiListResponse{
		Items: items,
	}

	// Handle cursor pagination
	cursorStr := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		lv, err := strconv.Atoi(limitStr)
		if err == nil && lv > 0 && lv <= 100 {
			limit = lv
		}
	}
	_ = cursorStr

	// Apply cursor-based pagination
	if len(items) > limit {
		hasMore := true
		items = items[:limit]
		resp.Items = items

		if len(items) > 0 {
			last := items[len(items)-1]
			c := &cursor.Cursor{
				ID:        uuid.MustParse(last.ID),
				UpdatedAt: time.Now(), // simplified; real impl uses actual timestamp
			}
			next, err := c.Encode()
			if err == nil {
				resp.NextCursor = &next
			}
		}
		_ = hasMore
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreateKPI handles POST /api/v1/kpis.
func (h *GoalHandler) CreateKPI(w http.ResponseWriter, r *http.Request) {
	var req dtogoal.CreateKpiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid JSON body", err))
		return
	}

	kpi, err := h.kpiService.CreateKPI(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, kpiRowToResponse(kpi))
}

// UpdateKPI handles PUT /api/v1/kpis/{kpiId}.
func (h *GoalHandler) UpdateKPI(w http.ResponseWriter, r *http.Request) {
	kpiIDStr := chi.URLParam(r, "kpiId")
	kpiID, err := uuid.Parse(kpiIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid KPI ID", err))
		return
	}

	var req dtogoal.UpdateKpiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid JSON body", err))
		return
	}

	kpi, err := h.kpiService.UpdateKPI(r.Context(), kpiID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, kpiRowToResponse(kpi))
}

// DeleteKPI handles DELETE /api/v1/kpis/{kpiId}.
func (h *GoalHandler) DeleteKPI(w http.ResponseWriter, r *http.Request) {
	kpiIDStr := chi.URLParam(r, "kpiId")
	kpiID, err := uuid.Parse(kpiIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid KPI ID", err))
		return
	}

	if err := h.kpiService.DeleteKPI(r.Context(), kpiID); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// KPI Linking Handlers
// ============================================================================

// LinkKPI handles POST /api/v1/goals/{goalId}/kpis.
func (h *GoalHandler) LinkKPI(w http.ResponseWriter, r *http.Request) {
	empID, err := parseEmpID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	goalIDStr := chi.URLParam(r, "goalId")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid goal ID", err))
		return
	}

	var req dtogoal.LinkKpiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid JSON body", err))
		return
	}

	kpiID, err := uuid.Parse(req.KpiID)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid KPI ID", err))
		return
	}

	if err := h.kpiService.LinkKPI(r.Context(), empID, goalID, kpiID); err != nil {
		writeError(w, err)
		return
	}

	// Return the updated goal
	goal, err := h.goalRepo.GetGoal(r.Context(), goalID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, goalRowToResponse(goal))
}

// UnlinkKPI handles DELETE /api/v1/goals/{goalId}/kpis/{kpiId}.
func (h *GoalHandler) UnlinkKPI(w http.ResponseWriter, r *http.Request) {
	empID, err := parseEmpID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	goalIDStr := chi.URLParam(r, "goalId")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid goal ID", err))
		return
	}

	kpiIDStr := chi.URLParam(r, "kpiId")
	kpiID, err := uuid.Parse(kpiIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid KPI ID", err))
		return
	}

	if err := h.kpiService.UnlinkKPI(r.Context(), empID, goalID, kpiID); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// Assignment Handlers
// ============================================================================

// assignmentRowToResponse converts a repo AssignmentRow to an API response.
func assignmentRowToResponse(a *repogoal.AssignmentRow) dtogoal.AssignmentResponse {
	return dtogoal.AssignmentResponse{
		ID:         a.ID.String(),
		EmployeeID: a.EmployeeID.String(),
		CycleID:    a.CycleID.String(),
		CreatedAt:  a.CreatedAt.Format(time.RFC3339),
	}
}

// GetAssignment handles GET /api/v1/employees/{empId}/assignments.
func (h *GoalHandler) GetAssignment(w http.ResponseWriter, r *http.Request) {
	empID, err := parseEmpID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	assignment, err := h.assignRepo.GetAssignment(r.Context(), empID)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := assignmentRowToResponse(assignment)

	// Fetch categories for the assignment
	cats, _ := h.catRepo.ListCategoriesByEmployee(r.Context(), empID)
	if cats != nil {
		catResponses := make([]dtogoal.CategoryResponse, len(cats))
		for i, c := range cats {
			cr := categoryRowToResponse(c)
			// Fetch goals for each category
			goals, _ := h.goalRepo.ListGoalsByCategory(r.Context(), c.ID)
			if goals != nil {
				goalResponses := make([]dtogoal.GoalResponse, len(goals))
				for j, g := range goals {
					goalResponses[j] = goalRowToResponse(g)
				}
				cr.Goals = goalResponses
			}
			catResponses[i] = cr
		}
		resp.Categories = catResponses
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreateAssignment handles POST /api/v1/employees/{empId}/assignments.
func (h *GoalHandler) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	empID, err := parseEmpID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req dtogoal.CreateAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid JSON body", err))
		return
	}

	cycleID, err := uuid.Parse(req.CycleID)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid cycle ID", err))
		return
	}

	assignment, err := h.assignRepo.CreateAssignment(r.Context(), empID, cycleID)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := assignmentRowToResponse(assignment)

	// Fetch categories
	cats, _ := h.catRepo.ListCategoriesByEmployee(r.Context(), empID)
	if cats != nil {
		catResponses := make([]dtogoal.CategoryResponse, len(cats))
		for i, c := range cats {
			catResponses[i] = categoryRowToResponse(c)
		}
		resp.Categories = catResponses
	}

	writeJSON(w, http.StatusCreated, resp)
}
