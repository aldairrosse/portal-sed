// Package evaluation provides HTTP handlers for the evaluation lifecycle
// and 9×9 matrix API. Handlers are thin: decode, validate, call service, encode.
//
// # TODO(auth:C7)
//
// All endpoints are decorated with TODO(auth:C7) markers for future
// authentication and RBAC integration. Currently, AuthPlaceholder middleware
// injects a mock "rh" identity.
package evaluation

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/evaluation"
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// contextKey for handler-specific context values.
type contextKey struct{ name string }

// generateTraceID creates a short trace ID for error responses.
func generateTraceID() string {
	id := uuid.New().String()
	return id[:8] + "-" + id[9:13]
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("handler/evaluation: failed to encode JSON response: %v", err)
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

// EvaluationHandler holds HTTP handlers for evaluation operations.
type EvaluationHandler struct {
	evalSvc      EvalService
	nineBoxSvc   BoxService
	dashboardSvc DashService
}

// NewEvaluationHandler creates a new EvaluationHandler.
func NewEvaluationHandler(
	evalSvc EvalService,
	nineBoxSvc BoxService,
	dashboardSvc DashService,
) *EvaluationHandler {
	return &EvaluationHandler{
		evalSvc:      evalSvc,
		nineBoxSvc:   nineBoxSvc,
		dashboardSvc: dashboardSvc,
	}
}

// --- Evaluation Endpoints ---

// ListEvaluations handles GET /api/v1/evaluations
// TODO(auth:C7): Restrict to rh, admin roles.
func (h *EvaluationHandler) ListEvaluations(w http.ResponseWriter, r *http.Request) {
	cycleIDStr := r.URL.Query().Get("cycle_id")
	if cycleIDStr == "" {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"cycle_id query parameter is required", nil))
		return
	}

	cycleID, err := uuid.Parse(cycleIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"cycle_id must be a valid UUID v4", err))
		return
	}

	state := r.URL.Query().Get("state")
	cursor := r.URL.Query().Get("cursor")

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		lv, err := strconv.Atoi(l)
		if err != nil {
			writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
				"limit must be a valid integer", err))
			return
		}
		if lv < 1 || lv > 100 {
			writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
				"limit must be between 1 and 100", nil))
			return
		}
		limit = lv
	}

	result, err := h.evalSvc.ListEvaluations(r.Context(), cycleID, state, cursor, limit)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetEvaluation handles GET /api/v1/evaluations/{id}
// TODO(auth:C7): Restrict to owner, manager, rh roles.
func (h *EvaluationHandler) GetEvaluation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"evaluation id must be a valid UUID v4", err))
		return
	}

	result, err := h.evalSvc.GetEvaluation(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// SubmitSelfEvaluation handles POST /api/v1/evaluations/{id}/self-evaluation
// TODO(auth:C7): Restrict to evaluation owner.
func (h *EvaluationHandler) SubmitSelfEvaluation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"evaluation id must be a valid UUID v4", err))
		return
	}

	var req dto.SelfEvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid JSON body", err))
		return
	}

	if len(req.Competencies) == 0 {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"at least one competency rating is required", nil))
		return
	}

	for i, c := range req.Competencies {
		if c.Rating < 1 || c.Rating > 5 {
			writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
				"competency rating must be between 1 and 5", nil).
				WithDetails("index: "+strconv.Itoa(i), "rating: "+strconv.Itoa(c.Rating)))
			return
		}
	}

	idempotencyKey := middleware.IdempotencyKeyFromContext(r.Context())

	result, err := h.evalSvc.SubmitSelfEvaluation(r.Context(), id, req, idempotencyKey)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// UpdateSelfEvaluation handles PUT /api/v1/evaluations/{id}/self-evaluation
// TODO(auth:C7): Restrict to evaluation owner.
func (h *EvaluationHandler) UpdateSelfEvaluation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"evaluation id must be a valid UUID v4", err))
		return
	}

	ifMatch := middleware.ExpectedVersionFromContext(r.Context())

	var req dto.SelfEvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid JSON body", err))
		return
	}

	if len(req.Competencies) == 0 {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"at least one competency rating is required", nil))
		return
	}

	result, err := h.evalSvc.UpdateSelfEvaluation(r.Context(), id, req, ifMatch)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// SubmitRHEvaluation handles POST /api/v1/evaluations/{id}/rh-evaluation
// TODO(auth:C7): Restrict to rh, admin roles.
func (h *EvaluationHandler) SubmitRHEvaluation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"evaluation id must be a valid UUID v4", err))
		return
	}

	var req dto.RHEvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid JSON body", err))
		return
	}

	if len(req.Competencies) == 0 {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"at least one competency rating is required", nil))
		return
	}

	idempotencyKey := middleware.IdempotencyKeyFromContext(r.Context())

	result, err := h.evalSvc.SubmitRHEvaluation(r.Context(), id, req, idempotencyKey)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// UpdateRHEvaluation handles PUT /api/v1/evaluations/{id}/rh-evaluation
// TODO(auth:C7): Restrict to rh, admin roles.
func (h *EvaluationHandler) UpdateRHEvaluation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"evaluation id must be a valid UUID v4", err))
		return
	}

	ifMatch := middleware.ExpectedVersionFromContext(r.Context())

	var req dto.RHEvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid JSON body", err))
		return
	}

	if len(req.Competencies) == 0 {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"at least one competency rating is required", nil))
		return
	}

	result, err := h.evalSvc.UpdateRHEvaluation(r.Context(), id, req, ifMatch)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// FinalizeEvaluation handles POST /api/v1/evaluations/{id}/finalize
// TODO(auth:C7): Restrict to rh, admin roles.
func (h *EvaluationHandler) FinalizeEvaluation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"evaluation id must be a valid UUID v4", err))
		return
	}

	var req dto.FinalizeEvaluationRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // body is optional

	result, err := h.evalSvc.FinalizeEvaluation(r.Context(), id, req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetEvaluationSummary handles GET /api/v1/evaluations/summary
// TODO(auth:C7): Restrict to rh, admin roles.
func (h *EvaluationHandler) GetEvaluationSummary(w http.ResponseWriter, r *http.Request) {
	cycleIDStr := r.URL.Query().Get("cycle_id")
	if cycleIDStr == "" {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"cycle_id query parameter is required", nil))
		return
	}

	cycleID, err := uuid.Parse(cycleIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"cycle_id must be a valid UUID v4", err))
		return
	}

	result, err := h.dashboardSvc.GetSummary(r.Context(), cycleID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// --- Nine-Box Endpoints ---

// ListMatrices handles GET /api/v1/nine-box/matrices
// TODO(auth:C7): Restrict to evaluator, rh roles.
func (h *EvaluationHandler) ListMatrices(w http.ResponseWriter, r *http.Request) {
	var cycleID, evaluatorID uuid.UUID

	if c := r.URL.Query().Get("cycle_id"); c != "" {
		var err error
		cycleID, err = uuid.Parse(c)
		if err != nil {
			writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
				"cycle_id must be a valid UUID v4", err))
			return
		}
	}

	if e := r.URL.Query().Get("evaluator_id"); e != "" {
		var err error
		evaluatorID, err = uuid.Parse(e)
		if err != nil {
			writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
				"evaluator_id must be a valid UUID v4", err))
			return
		}
	}

	result, err := h.nineBoxSvc.ListMatrices(r.Context(), cycleID, evaluatorID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// CreateMatrix handles POST /api/v1/nine-box/matrices
// TODO(auth:C7): Restrict to evaluator, rh roles.
func (h *EvaluationHandler) CreateMatrix(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CycleID     uuid.UUID `json:"cycleId"`
		EvaluatorID uuid.UUID `json:"evaluatorId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid JSON body", err))
		return
	}

	if req.CycleID == uuid.Nil || req.EvaluatorID == uuid.Nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"cycleId and evaluatorId are required", nil))
		return
	}

	result, err := h.nineBoxSvc.CreateMatrix(r.Context(), req.CycleID, req.EvaluatorID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// GetMatrix handles GET /api/v1/nine-box/matrices/{matrixId}
// TODO(auth:C7): Restrict to evaluator owner, rh roles.
func (h *EvaluationHandler) GetMatrix(w http.ResponseWriter, r *http.Request) {
	matrixID, err := uuid.Parse(chi.URLParam(r, "matrixId"))
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"matrixId must be a valid UUID v4", err))
		return
	}

	result, err := h.nineBoxSvc.GetMatrix(r.Context(), matrixID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ListMatrixEntries handles GET /api/v1/nine-box/matrices/{matrixId}/entries
// TODO(auth:C7): Restrict to evaluator owner, rh roles.
func (h *EvaluationHandler) ListMatrixEntries(w http.ResponseWriter, r *http.Request) {
	matrixID, err := uuid.Parse(chi.URLParam(r, "matrixId"))
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"matrixId must be a valid UUID v4", err))
		return
	}

	// Get the full matrix which includes entries
	result, err := h.nineBoxSvc.GetMatrix(r.Context(), matrixID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result.Entries)
}

// UpsertMatrixEntry handles POST /api/v1/nine-box/matrices/{matrixId}/entries
// TODO(auth:C7): Restrict to evaluator owner.
func (h *EvaluationHandler) UpsertMatrixEntry(w http.ResponseWriter, r *http.Request) {
	matrixID, err := uuid.Parse(chi.URLParam(r, "matrixId"))
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"matrixId must be a valid UUID v4", err))
		return
	}

	var req dto.NineBoxEntryInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid JSON body", err))
		return
	}

	if req.EvaluateeID == uuid.Nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"evaluateeId is required", nil))
		return
	}

	result, err := h.nineBoxSvc.UpsertEntry(r.Context(), matrixID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// UpdateEntry handles PUT /api/v1/nine-box/entries/{entryId}
// TODO(auth:C7): Restrict to evaluator owner.
func (h *EvaluationHandler) UpdateEntry(w http.ResponseWriter, r *http.Request) {
	entryID, err := uuid.Parse(chi.URLParam(r, "entryId"))
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"entryId must be a valid UUID v4", err))
		return
	}

	ifMatch := middleware.ExpectedVersionFromContext(r.Context())

	var req dto.NineBoxEntryInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid JSON body", err))
		return
	}

	result, err := h.nineBoxSvc.UpdateEntry(r.Context(), entryID, req, ifMatch)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// BatchSubmitEntries handles POST /api/v1/nine-box/batch
// TODO(auth:C7): Restrict to evaluator owner.
func (h *EvaluationHandler) BatchSubmitEntries(w http.ResponseWriter, r *http.Request) {
	var req dto.NineBoxBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid JSON body", err))
		return
	}

	if len(req.Entries) == 0 {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"at least one entry is required", nil))
		return
	}

	// The batch request must be for a specific matrix; get from query param
	matrixIDStr := r.URL.Query().Get("matrixId")
	if matrixIDStr == "" {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"matrixId query parameter is required for batch submission", nil))
		return
	}

	matrixID, err := uuid.Parse(matrixIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"matrixId must be a valid UUID v4", err))
		return
	}

	result, err := h.nineBoxSvc.BatchSubmitEntries(r.Context(), matrixID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetScales handles GET /api/v1/nine-box/scales
// TODO(auth:C7): Any authenticated user.
func (h *EvaluationHandler) GetScales(w http.ResponseWriter, r *http.Request) {
	result, err := h.nineBoxSvc.GetScales(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetQuadrants handles GET /api/v1/nine-box/quadrants
// TODO(auth:C7): Any authenticated user.
func (h *EvaluationHandler) GetQuadrants(w http.ResponseWriter, r *http.Request) {
	result, err := h.nineBoxSvc.GetQuadrants(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}
