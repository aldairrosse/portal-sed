package weight

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	weightsvc "github.com/sed-evaluacion-desempeno/api/internal/service/weight"
)

type Handler struct {
	svc *weightsvc.Service
}

func NewHandler(svc *weightsvc.Service) *Handler { return &Handler{svc: svc} }

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, err error) {
	code := pkgerrors.HTTPStatus(err)
	var de *pkgerrors.DomainError
	if pkgerrors.AsDomainError(err, &de) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(pkgerrors.NewAPIErrorResponse(de, ""))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(pkgerrors.NewAPIErrorResponse(pkgerrors.NewDomainError(pkgerrors.InvalidRequest, err.Error(), err), ""))
}

func (h *Handler) GetCycleConfig(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.resolveOrgID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	weights, _, err := h.svc.GetCycleConfig(r.Context(), orgID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, weights)
}

func (h *Handler) PutCycleConfig(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.resolveOrgID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var body struct {
		GWeight *float64 `json:"g_weight"`
		G       *float64 `json:"g"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "JSON inválido", err))
		return
	}
	var g float64
	if body.GWeight != nil {
		g = *body.GWeight
	} else if body.G != nil {
		g = *body.G
	} else {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "g_weight es requerido", nil))
		return
	}
	weights, err := h.svc.SaveCycleConfig(r.Context(), orgID, g)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, weights)
}

func (h *Handler) GetTeamConfig(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.resolveOrgID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	callerID, _ := auth.GetEmployeeID(r.Context())
	teamID := parseTeamID(r)
	weights, _, err := h.svc.GetTeamConfig(r.Context(), orgID, teamID, callerID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, weights)
}

func (h *Handler) PutTeamConfig(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.resolveOrgID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	callerID, _ := auth.GetEmployeeID(r.Context())
	teamID := parseTeamID(r)
	var body struct {
		JWeight *float64 `json:"j_weight"`
		J       *float64 `json:"j"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "JSON inválido", err))
		return
	}
	var j float64
	if body.JWeight != nil {
		j = *body.JWeight
	} else if body.J != nil {
		j = *body.J
	} else {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "j_weight es requerido", nil))
		return
	}
	weights, err := h.svc.SaveTeamConfig(r.Context(), orgID, teamID, callerID, j)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, weights)
}

func parseTeamID(r *http.Request) uuid.UUID {
	s := r.URL.Query().Get("team_id")
	if s == "" {
		s = r.URL.Query().Get("teamId")
	}
	if s == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

// resolveOrgID tries middleware context, then session employee -> org lookup via DB through service helper.
func (h *Handler) resolveOrgID(r *http.Request) (uuid.UUID, error) {
	if oidStr := middleware.OrgIDFromContext(r.Context()); oidStr != "" {
		if id, err := uuid.Parse(oidStr); err == nil {
			return id, nil
		}
	}
	if hdr := r.Header.Get("X-Organization-Id"); hdr != "" {
		if id, err := uuid.Parse(hdr); err == nil {
			return id, nil
		}
	}
	// Fallback: use auth handler's OrgNodeInfo pattern via direct DB query through service's DB
	// We can ask svc to resolve org from caller employee.
	if empID, ok := auth.GetEmployeeID(r.Context()); ok {
		// Try to fetch organization_id via service db query (exposed via helper)
		if orgID, err := h.svc.ResolveOrgIDForEmployee(r.Context(), empID); err == nil && orgID != uuid.Nil {
			return orgID, nil
		}
	}
	return uuid.Nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "organization_id no disponible en sesión", nil)
}
