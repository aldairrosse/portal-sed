package evaluation

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// GetExport handles GET /api/v1/evaluations/export.
// Returns the 7-column backend-computed export over the full viewer scope
// (all active employees for RH, report subtree otherwise). cycle_id is
// optional: empty defaults to the active cycle; phase defaults to current_phase.
func (h *EvaluationHandler) GetExport(w http.ResponseWriter, r *http.Request) {
	if h.exportSvc == nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.DomainCode("INTERNAL"), "export service not configured", nil))
		return
	}
	var cycleID uuid.UUID
	if s := r.URL.Query().Get("cycle_id"); s != "" {
		var err error
		cycleID, err = uuid.Parse(s)
		if err != nil {
			writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
				"cycle_id must be a valid UUID v4", err))
			return
		}
	}
	phase, err := parsePhaseParam(r.URL.Query().Get("phase"))
	if err != nil {
		writeError(w, err)
		return
	}
	viewerID, ok := auth.GetEmployeeID(r.Context())
	if !ok {
		writeError(w, pkgerrors.ErrForbidden)
		return
	}
	viewerRole, ok := auth.GetRole(r.Context())
	if !ok {
		writeError(w, pkgerrors.ErrForbidden)
		return
	}
	result, err := h.exportSvc.Export(r.Context(), cycleID, phase, r.URL.Query().Get("q"), viewerID, viewerRole)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
