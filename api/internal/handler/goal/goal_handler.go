package goal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	dtogoal "github.com/sed-evaluacion-desempeno/api/internal/dto/goal"
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/cursor"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/scoring"
	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
	activitysvc "github.com/sed-evaluacion-desempeno/api/internal/service/activity"
	svcgoal "github.com/sed-evaluacion-desempeno/api/internal/service/goal"
)

func generateTraceID() string {
	id := uuid.New().String()
	return id[:8] + "-" + id[9:13]
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("handler: failed to encode JSON response: %v", err)
	}
}

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

type CycleResolver interface {
	ResolveActiveCycleID(ctx context.Context, employeeID uuid.UUID) (uuid.UUID, error)
}
type GoalHandler struct {
	catService    svcgoal.CategoryServicer
	goalService   svcgoal.GoalServicer
	progressSvc   svcgoal.ProgressServicer
	kpiService    svcgoal.KpiServicer
	scoringSvc    svcgoal.ScoringServicer
	weightSvc     svcgoal.WeightValidationServicer
	batchService  svcgoal.BatchServicer
	proposalSvc   svcgoal.GoalProposalServicer
	catRepo       svcgoal.CategoryRepository
	goalRepo      svcgoal.GoalRepository
	kpiRepo       svcgoal.KPIRepository
	linkRepo      svcgoal.LinkKPIRepository
	assignRepo    svcgoal.AssignmentRepository
	proposalRepo  svcgoal.GoalProposalRepository
	activitySvc   activitysvc.Service
	cycleResolver CycleResolver
}

func NewGoalHandler(
	catService svcgoal.CategoryServicer,
	goalService svcgoal.GoalServicer,
	progressSvc svcgoal.ProgressServicer,
	kpiService svcgoal.KpiServicer,
	scoringSvc svcgoal.ScoringServicer,
	weightSvc svcgoal.WeightValidationServicer,
	batchService svcgoal.BatchServicer,
	proposalSvc svcgoal.GoalProposalServicer,
	catRepo svcgoal.CategoryRepository,
	goalRepo svcgoal.GoalRepository,
	kpiRepo svcgoal.KPIRepository,
	linkRepo svcgoal.LinkKPIRepository,
	assignRepo svcgoal.AssignmentRepository,
	proposalRepo svcgoal.GoalProposalRepository,
	activitySvc activitysvc.Service,
	cycleResolver CycleResolver,
) *GoalHandler {
	return &GoalHandler{
		catService:    catService,
		goalService:   goalService,
		progressSvc:   progressSvc,
		kpiService:    kpiService,
		scoringSvc:    scoringSvc,
		weightSvc:     weightSvc,
		batchService:  batchService,
		proposalSvc:   proposalSvc,
		catRepo:       catRepo,
		goalRepo:      goalRepo,
		kpiRepo:       kpiRepo,
		linkRepo:      linkRepo,
		assignRepo:    assignRepo,
		proposalRepo:  proposalRepo,
		activitySvc:   activitySvc,
		cycleResolver: cycleResolver,
	}
}

// ============================================================================
// Category Handlers
// ============================================================================

func parseEmpID(r *http.Request) (uuid.UUID, error) {
	empIDStr := chi.URLParam(r, "empId")
	if empIDStr == "" {
		empIDStr = middleware.EmployeeIDFromContext(r.Context())
	}
	if empIDStr == "" {
		return uuid.Nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "employee ID is required", nil)
	}
	return uuid.Parse(empIDStr)
}

func callerID(r *http.Request) (uuid.UUID, error) {
	empIDStr := middleware.EmployeeIDFromContext(r.Context())
	if empIDStr == "" {
		return uuid.Nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "authenticated employee ID is required", nil)
	}
	return uuid.Parse(empIDStr)
}

func categoryRowToResponse(c *repogoal.CategoryRow) dtogoal.CategoryResponse {
	var pillarID *string
	if c.PillarID != nil {
		s := c.PillarID.String()
		pillarID = &s
	}
	return dtogoal.CategoryResponse{
		ID:          c.ID.String(),
		EmployeeID:  c.EmployeeID.String(),
		Name:        c.Name,
		Description: c.Description,
		Weight:      c.Weight,
		PillarID:    pillarID,
		Goals:       nil,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
	}
}

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
	var allGoalIDs []uuid.UUID
	for i, c := range cats {
		cr := categoryRowToResponse(c)
		goals, _ := h.goalRepo.ListGoalsByCategory(r.Context(), c.ID)
		if goals != nil {
			goalResponses := make([]dtogoal.GoalResponse, len(goals))
			for j, g := range goals {
				kpiLinks, _ := h.linkRepo.ListLinksByGoal(r.Context(), g.ID)
				var kpis []dtogoal.KpiResponse
				for _, kl := range kpiLinks {
					if kl.Kpi != nil {
						kpis = append(kpis, kpiRowToResponse(kl.Kpi))
					}
				}
				goalResponses[j] = goalRowToResponse(g, kpis)
				allGoalIDs = append(allGoalIDs, g.ID)
			}
			cr.Goals = goalResponses
		}
		items[i] = cr
	}

	if len(allGoalIDs) > 0 {
		pendingProps, err := h.proposalRepo.ListPendingByGoalIDs(r.Context(), allGoalIDs)
		if err == nil && len(pendingProps) > 0 {
			pendingByGoal := make(map[string]*repogoal.GoalProposalRow, len(pendingProps))
			for _, p := range pendingProps {
				pendingByGoal[p.GoalID.String()] = p
			}
			for i := range items {
				if items[i].Goals == nil {
					continue
				}
				for j := range items[i].Goals {
					if p, ok := pendingByGoal[items[i].Goals[j].ID]; ok {
						pr := proposalRowToResponse(p)
						items[i].Goals[j].PendingProposal = &pr
					}
				}
			}
		}
	}

	resp := dtogoal.CategoryListResponse{
		Items: items,
	}
	writeJSON(w, http.StatusOK, resp)
}

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

func goalRowToResponse(g *repogoal.GoalRow, kpis ...[]dtogoal.KpiResponse) dtogoal.GoalResponse {
	baselineVal := 0.0
	if g.BaselineValue != nil {
		baselineVal = *g.BaselineValue
	}
	var kpiResponses []dtogoal.KpiResponse
	if len(kpis) > 0 {
		kpiResponses = kpis[0]
	}
	return dtogoal.GoalResponse{
		ID:              g.ID.String(),
		CategoryID:      g.CategoryID.String(),
		Name:            g.Name,
		Description:     g.Description,
		Unit:            g.Unit,
		Weight:          g.Weight,
		TargetValue:     g.TargetValue,
		CurrentValue:    g.CurrentValue,
		Direction:       g.Direction,
		BaselineValue:   g.BaselineValue,
		ProgressPercent: scoring.ProgressPercent(g.CurrentValue, g.TargetValue, baselineVal, g.Direction),
		State:           g.State,
		Version:         g.Version,
		KPIs:            kpiResponses,
		CreatedAt:       g.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       g.UpdatedAt.Format(time.RFC3339),
	}
}

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

	if h.activitySvc != nil {
		_ = h.activitySvc.LogActivity(r.Context(), empID, "goal_progress",
			"Actualizaste el progreso de tu meta", "Metas", nil)
	}

	writeJSON(w, http.StatusOK, goalRowToResponse(goal))
}

// ============================================================================
// Batch Handlers
// ============================================================================

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

func kpiRowToResponse(k *repogoal.KpiRow) dtogoal.KpiResponse {
	progressPercent := 0.0
	var orgNodeID *string
	if k.OrgNodeID != nil {
		s := k.OrgNodeID.String()
		orgNodeID = &s
	}
	return dtogoal.KpiResponse{
		ID:              k.ID.String(),
		Name:            k.Name,
		Unit:            k.Unit,
		Description:     k.Description,
		Direction:       k.Direction,
		TargetValue:     k.TargetValue,
		CurrentValue:    k.CurrentValue,
		ProgressPercent: progressPercent,
		CreatedAt:       k.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       k.UpdatedAt.Format(time.RFC3339),
		OrgNodeID:       orgNodeID,
	}
}

func (h *GoalHandler) ListKPIs(w http.ResponseWriter, r *http.Request) {
	empID, err := callerID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	kpis, err := h.kpiService.ListKPIs(r.Context(), empID)
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

	if len(items) > limit {
		hasMore := true
		items = items[:limit]
		resp.Items = items

		if len(items) > 0 {
			last := items[len(items)-1]
			c := &cursor.Cursor{
				ID:        uuid.MustParse(last.ID),
				UpdatedAt: time.Now(),
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

func (h *GoalHandler) CreateKPI(w http.ResponseWriter, r *http.Request) {
	var req dtogoal.CreateKpiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid JSON body", err))
		return
	}

	empID, err := callerID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	kpi, err := h.kpiService.CreateKPI(r.Context(), req, empID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, kpiRowToResponse(kpi))
}

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

func (h *GoalHandler) UpdateKPIValue(w http.ResponseWriter, r *http.Request) {
	kpiIDStr := chi.URLParam(r, "kpiId")
	kpiID, err := uuid.Parse(kpiIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid KPI ID", err))
		return
	}

	var req dtogoal.KpiUpdateValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid JSON body", err))
		return
	}

	kpi, err := h.kpiService.UpdateKPIValue(r.Context(), kpiID, req.CurrentValue)
	if err != nil {
		writeError(w, err)
		return
	}

	if h.activitySvc != nil {
		if empID, ok := auth.GetEmployeeID(r.Context()); ok {
			_ = h.activitySvc.LogActivity(r.Context(), empID, "goal_progress",
				"Actualizaste el valor de un KPI", "Metas", nil)
		}
	}

	writeJSON(w, http.StatusOK, kpiRowToResponse(kpi))
}

// ============================================================================
// KPI Linking Handlers
// ============================================================================

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

	goal, err := h.goalRepo.GetGoal(r.Context(), goalID)
	if err != nil {
		writeError(w, err)
		return
	}
	kpiLinks, _ := h.linkRepo.ListLinksByGoal(r.Context(), goalID)
	var kpis []dtogoal.KpiResponse
	for _, kl := range kpiLinks {
		if kl.Kpi != nil {
			kpis = append(kpis, kpiRowToResponse(kl.Kpi))
		}
	}

	writeJSON(w, http.StatusOK, goalRowToResponse(goal, kpis))
}

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
// Scoring Handlers
// ============================================================================

func (h *GoalHandler) GetEmployeeScore(w http.ResponseWriter, r *http.Request) {
	empID, err := parseEmpID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	score, err := h.scoringSvc.GetEmployeeScore(r.Context(), empID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]float64{"score": score})
}

// ============================================================================
// Assignment Handlers
// ============================================================================

func assignmentRowToResponse(a *repogoal.AssignmentRow) dtogoal.AssignmentResponse {
	var submittedAt *string
	if a.SubmittedAt != nil {
		s := a.SubmittedAt.Format(time.RFC3339)
		submittedAt = &s
	}
	status := a.Status
	if status == "" {
		status = "borrador"
	}
	return dtogoal.AssignmentResponse{
		ID:          a.ID.String(),
		EmployeeID:  a.EmployeeID.String(),
		CycleID:     a.CycleID.String(),
		Status:      status,
		SubmittedAt: submittedAt,
		CreatedAt:   a.CreatedAt.Format(time.RFC3339),
	}
}

func assignedGoalResponse(id string, name, description, unit, direction, kind string, weight, target, current float64, baseline *float64, state, source string) dtogoal.AssignedGoalResponse {
	baselineValue := 0.0
	if baseline != nil {
		baselineValue = *baseline
	}
	return dtogoal.AssignedGoalResponse{
		ID: id, Name: name, Description: description, Unit: unit, Direction: direction,
		GoalKind: kind, Weight: weight, TargetValue: target, BaselineValue: baseline,
		CurrentValue: current, ProgressPercent: scoring.ProgressPercent(current, target, baselineValue, direction),
		State: state, Source: source,
	}
}

func (h *GoalHandler) GetAssignment(w http.ResponseWriter, r *http.Request) {
	empID, err := parseEmpID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	assignment, err := h.assignRepo.GetAssignment(r.Context(), empID)
	if err != nil {
		var de *pkgerrors.DomainError
		if !pkgerrors.AsDomainError(err, &de) || de.Code != pkgerrors.GoalNotFound {
			writeError(w, err)
			return
		}
		// No personal assignment: fall through so global/shared goals are still returned.
		assignment = nil
	}

	var resp dtogoal.AssignmentResponse
	if assignment != nil {
		resp = assignmentRowToResponse(assignment)
	} else {
		resp = dtogoal.AssignmentResponse{EmployeeID: empID.String()}
		if h.cycleResolver != nil {
			if cid, cerr := h.cycleResolver.ResolveActiveCycleID(r.Context(), empID); cerr == nil {
				resp.CycleID = cid.String()
			}
		}
	}
	resp.Categories = []dtogoal.CategoryResponse{}
	globalGoals, err := h.assignRepo.ListGlobalGoalsByEmployee(r.Context(), empID)
	if err != nil {
		writeError(w, err)
		return
	}
	for _, g := range globalGoals {
		if len(g.Assignments) > 0 {
			a := g.Assignments[0]
			resp.GlobalGoals = append(resp.GlobalGoals, assignedGoalResponse(g.ID.String(), g.Name, g.Description, g.Unit, g.Direction, g.GoalKind, a.Weight, a.TargetValue, g.CurrentValue, a.BaselineValue, g.State, "global"))
		}
	}
	sharedGoals, err := h.assignRepo.ListSharedGoalsAsMember(r.Context(), empID)
	if err != nil {
		writeError(w, err)
		return
	}
	for _, g := range sharedGoals {
		if len(g.Members) > 0 {
			m := g.Members[0]
			resp.SharedGoals = append(resp.SharedGoals, assignedGoalResponse(g.ID.String(), g.Name, g.Description, g.Unit, g.Direction, g.GoalKind, m.Weight, m.TargetValue, g.CurrentValue, m.BaselineValue, g.State, "shared"))
		}
	}

	cats, _ := h.catRepo.ListCategoriesByEmployee(r.Context(), empID)
	if cats != nil {
		catResponses := make([]dtogoal.CategoryResponse, len(cats))
		var allGoalIDs []uuid.UUID
		for i, c := range cats {
			cr := categoryRowToResponse(c)
			goals, _ := h.goalRepo.ListGoalsByCategory(r.Context(), c.ID)
			if goals != nil {
				goalResponses := make([]dtogoal.GoalResponse, len(goals))
				for j, g := range goals {
					kpiLinks, _ := h.linkRepo.ListLinksByGoal(r.Context(), g.ID)
					var kpis []dtogoal.KpiResponse
					for _, kl := range kpiLinks {
						if kl.Kpi != nil {
							kpis = append(kpis, kpiRowToResponse(kl.Kpi))
						}
					}
					goalResponses[j] = goalRowToResponse(g, kpis)
					allGoalIDs = append(allGoalIDs, g.ID)
				}
				cr.Goals = goalResponses
			}
			catResponses[i] = cr
		}

		if len(allGoalIDs) > 0 {
			pendingProps, err := h.proposalRepo.ListPendingByGoalIDs(r.Context(), allGoalIDs)
			if err == nil && len(pendingProps) > 0 {
				pendingByGoal := make(map[string]*repogoal.GoalProposalRow, len(pendingProps))
				for _, p := range pendingProps {
					pendingByGoal[p.GoalID.String()] = p
				}
				for i := range catResponses {
					if catResponses[i].Goals == nil {
						continue
					}
					for j := range catResponses[i].Goals {
						if p, ok := pendingByGoal[catResponses[i].Goals[j].ID]; ok {
							pr := proposalRowToResponse(p)
							catResponses[i].Goals[j].PendingProposal = &pr
						}
					}
				}
			}
		}

		resp.Categories = catResponses
	}

	writeJSON(w, http.StatusOK, resp)
}

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

	var cycleID uuid.UUID
	if req.CycleID != nil && *req.CycleID != "" {
		cycleID, err = uuid.Parse(*req.CycleID)
		if err != nil {
			writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid cycle ID", err))
			return
		}
	} else if h.cycleResolver != nil {
		cycleID, err = h.cycleResolver.ResolveActiveCycleID(r.Context(), empID)
		if err != nil {
			writeError(w, err)
			return
		}
	} else {
		writeError(w, pkgerrors.ErrCycleNotFound)
		return
	}

	assignment, err := h.assignRepo.CreateAssignment(r.Context(), empID, cycleID)
	if err != nil {
		writeError(w, err)
		return
	}

	// Weight validation gate: if Double 100% passes, transition to 'enviada'.
	if h.weightSvc != nil {
		v, verr := h.weightSvc.ValidateDoubleWeighting(r.Context(), empID)
		if verr != nil {
			writeError(w, verr)
			return
		}
		if v != nil && !v.Valid {
			var details []string
			if math.Abs(v.CategorySum-100.0) > 0.01 {
				details = append(details, fmt.Sprintf("La suma de pesos de las categorías es %.1f%% (debe ser 100%%)", v.CategorySum))
			}
			for _, gs := range v.GoalSums {
				if math.Abs(gs.Sum-100.0) > 0.01 {
					details = append(details, fmt.Sprintf("Las metas en la categoría '%s' suman %.1f%% (debe ser 100%%)", gs.CategoryName, gs.Sum))
				}
			}
			if len(details) == 0 {
				details = append(details, "Las categorías y/o metas deben sumar 100%")
			}
			writeError(w, pkgerrors.NewDomainError(pkgerrors.WeightSumInvalid, "La asignación no cumple con la regla de Double 100% de pesos", nil).WithDetails(details...))
			return
		}

		now := time.Now().UTC()
		updated, uerr := h.assignRepo.UpdateAssignmentStatus(r.Context(), empID, cycleID, "enviada", &now)
		if uerr != nil {
			writeError(w, uerr)
			return
		}
		if updated != nil {
			assignment = updated
		}
	}

	resp := assignmentRowToResponse(assignment)
	resp.Categories = []dtogoal.CategoryResponse{}

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

// ============================================================================
// Goal Proposal Handlers
// ============================================================================

func proposalRowToResponse(p *repogoal.GoalProposalRow) dtogoal.GoalProposalResponse {
	var kpiIDs []string
	var reviewedBy *string
	var reviewedAt *string

	if p.ReviewedBy != nil {
		s := p.ReviewedBy.String()
		reviewedBy = &s
	}
	if p.ReviewedAt != nil {
		s := p.ReviewedAt.Format(time.RFC3339)
		reviewedAt = &s
	}

	return dtogoal.GoalProposalResponse{
		ID:            p.ID.String(),
		GoalID:        p.GoalID.String(),
		RequestedBy:   p.RequestedBy.String(),
		Name:          p.Name,
		Description:   p.Description,
		Unit:          p.Unit,
		Weight:        p.Weight,
		TargetValue:   p.TargetValue,
		Direction:     p.Direction,
		BaselineValue: p.BaselineValue,
		KpiIDs:        kpiIDs,
		Status:        p.Status,
		ReviewedBy:    reviewedBy,
		ReviewedAt:    reviewedAt,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *GoalHandler) CreateGoalProposal(w http.ResponseWriter, r *http.Request) {
	callerIDStr := middleware.EmployeeIDFromContext(r.Context())
	callerID, err := uuid.Parse(callerIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid authenticated employee ID", err))
		return
	}

	goalIDStr := chi.URLParam(r, "goalId")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid goal ID", err))
		return
	}

	var req dtogoal.CreateGoalProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid JSON body", err))
		return
	}

	proposal, err := h.proposalSvc.CreateProposal(r.Context(), callerID, goalID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, proposalRowToResponse(proposal))
}

func (h *GoalHandler) ListGoalProposals(w http.ResponseWriter, r *http.Request) {
	goalIDStr := chi.URLParam(r, "goalId")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid goal ID", err))
		return
	}

	proposals, err := h.proposalRepo.ListByGoal(r.Context(), goalID)
	if err != nil {
		writeError(w, err)
		return
	}

	items := make([]dtogoal.GoalProposalResponse, len(proposals))
	for i, p := range proposals {
		items[i] = proposalRowToResponse(p)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func (h *GoalHandler) UpdateGoalProposal(w http.ResponseWriter, r *http.Request) {
	callerIDStr := middleware.EmployeeIDFromContext(r.Context())
	callerID, err := uuid.Parse(callerIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid authenticated employee ID", err))
		return
	}

	goalIDStr := chi.URLParam(r, "goalId")
	_, err = uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid goal ID", err))
		return
	}

	propIDStr := chi.URLParam(r, "propId")
	propID, err := uuid.Parse(propIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid proposal ID", err))
		return
	}

	var req dtogoal.UpdateGoalProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid JSON body", err))
		return
	}

	switch req.Status {
	case "accepted":
		updatedGoal, err := h.proposalSvc.AcceptProposal(r.Context(), callerID, propID)
		if err != nil {
			writeError(w, err)
			return
		}
		resp := goalRowToResponse(updatedGoal)

		pendingProps, err := h.proposalRepo.ListPendingByGoalIDs(r.Context(), []uuid.UUID{updatedGoal.ID})
		if err == nil && len(pendingProps) > 0 {
			pr := proposalRowToResponse(pendingProps[0])
			resp.PendingProposal = &pr
		}

		writeJSON(w, http.StatusOK, resp)

	case "rejected":
		proposal, err := h.proposalSvc.RejectProposal(r.Context(), callerID, propID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, proposalRowToResponse(proposal))

	default:
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "status must be 'accepted' or 'rejected'", nil))
	}
}
