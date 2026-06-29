// Package activity provides HTTP handlers for the activity logs API.
// Handlers are thin: they validate input, call the service layer, and format
// responses. Business logic lives in the service layer.
package activity

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	svc "github.com/sed-evaluacion-desempeno/api/internal/service/activity"
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

// ActivityHandler holds HTTP handlers for activity log operations.
type ActivityHandler struct {
	svc svc.Service
}

// NewActivityHandler creates a new ActivityHandler.
func NewActivityHandler(svc svc.Service) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

// ListActivityLogs handles GET /api/v1/employees/{employeeId}/activity-logs.
func (h *ActivityHandler) ListActivityLogs(w http.ResponseWriter, r *http.Request) {
	employeeIDStr := chi.URLParam(r, "employeeId")
	if employeeIDStr == "" {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"employeeId path parameter is required", nil))
		return
	}

	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"employeeId must be a valid UUID v4", err))
		return
	}

	// Parse limit: default 50, max 100
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		lv, err := strconv.Atoi(l)
		if err != nil {
			writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
				"limit must be a valid integer", err))
			return
		}
		if lv < 1 {
			lv = 1
		} else if lv > 100 {
			lv = 100
		}
		limit = lv
	}

	logs, err := h.svc.ListByEmployee(r.Context(), employeeID, limit)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := map[string]interface{}{
		"data":  logs,
		"total": len(logs),
	}
	writeJSON(w, http.StatusOK, resp)
}
