// Package cycle provides HTTP handlers for the evaluation lifecycle API.
// Handlers are thin: they validate input, call the service layer, and format
// responses. Business logic lives in the service layer.
package cycle

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	svc "github.com/sed-evaluacion-desempeno/api/internal/service/cycle"
)

// contextKey for handler-specific context values.
type contextKey struct{ name string }

var idempCtxKey = &contextKey{"idempotency-key"}

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

// CycleHandler holds HTTP handlers for cycle operations.
type CycleHandler struct {
	svc          svc.Service
	phaseService svc.PhaseService
}

// NewCycleHandler creates a new CycleHandler.
func NewCycleHandler(svc svc.Service, phaseService svc.PhaseService) *CycleHandler {
	return &CycleHandler{
		svc:          svc,
		phaseService: phaseService,
	}
}

// ListCycles handles GET /api/v1/cycles.
func (h *CycleHandler) ListCycles(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("organization_id")
	if orgID == "" {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"organization_id query parameter is required", nil))
		return
	}

	var year *int
	if y := r.URL.Query().Get("year"); y != "" {
		yv, err := strconv.Atoi(y)
		if err != nil {
			writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
				"year must be a valid integer", err))
			return
		}
		year = &yv
	}

	var currentPhase *string
	if cp := r.URL.Query().Get("current_phase"); cp != "" {
		currentPhase = &cp
	}

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

	req := svc.ListCyclesRequest{
		OrganizationID: orgID,
		Year:           year,
		CurrentPhase:   currentPhase,
		Cursor:         cursor,
		Limit:          limit,
	}

	result, err := h.svc.ListCycles(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// CreateCycle handles POST /api/v1/cycles.
func (h *CycleHandler) CreateCycle(w http.ResponseWriter, r *http.Request) {
	var req svc.CreateCycleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid JSON body", err))
		return
	}

	// Extract idempotency key from context (injected by middleware)
	req.IdempotencyKey = middleware.IdempotencyKeyFromContext(r.Context())

	cycle, err := h.svc.CreateCycle(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, cycle)
}

// GetCycle handles GET /api/v1/cycles/{id}.
func (h *CycleHandler) GetCycle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"cycle id is required", nil))
		return
	}

	if _, err := uuid.Parse(id); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"cycle id must be a valid UUID v4", err))
		return
	}

	cycle, err := h.svc.GetCycle(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, cycle)
}

// TransitionPhase handles PUT /api/v1/cycles/{id}/transition.
func (h *CycleHandler) TransitionPhase(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"cycle id is required", nil))
		return
	}

	if _, err := uuid.Parse(id); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"cycle id must be a valid UUID v4", err))
		return
	}

	// Parse optional body
	var body struct {
		Trigger string `json:"trigger"`
		Reason  string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	// Get expected version from context (injected by OptimisticLock middleware)
	expectedVersion := middleware.ExpectedVersionFromContext(r.Context())

	// Get idempotency key from context
	idempotencyKey := middleware.IdempotencyKeyFromContext(r.Context())

	req := svc.TransitionPhaseRequest{
		CycleID:         id,
		ExpectedVersion: expectedVersion,
		Trigger:         body.Trigger,
		Reason:          body.Reason,
		IdempotencyKey:  idempotencyKey,
	}

	result, err := h.svc.TransitionPhase(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetPhaseDefinitions handles GET /api/v1/phases.
func (h *CycleHandler) GetPhaseDefinitions(w http.ResponseWriter, r *http.Request) {
	defs, etag, err := h.phaseService.GetPhaseDefinitions(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	// Check If-None-Match
	if r.Header.Get("If-None-Match") == `"`+etag+`"` {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("ETag", `"`+etag+`"`)
	w.Header().Set("Cache-Control", "max-age=3600")

	resp := map[string]interface{}{
		"data": defs,
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetAvailableTransitions handles GET /api/v1/cycles/{id}/transitions.
func (h *CycleHandler) GetAvailableTransitions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"cycle id is required", nil))
		return
	}

	if _, err := uuid.Parse(id); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"cycle id must be a valid UUID v4", err))
		return
	}

	transitions, err := h.phaseService.GetAvailableTransitions(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := map[string]interface{}{
		"data": transitions,
	}
	writeJSON(w, http.StatusOK, resp)
}


