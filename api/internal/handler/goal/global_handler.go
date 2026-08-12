package goal

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	servicegoal "github.com/sed-evaluacion-desempeno/api/internal/service/goal"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// GlobalGoalHandler handles HTTP requests for global goals.
type GlobalGoalHandler struct {
	service servicegoal.GlobalGoalServicer
}

// NewGlobalGoalHandler creates a new GlobalGoalHandler.
func NewGlobalGoalHandler(service servicegoal.GlobalGoalServicer) *GlobalGoalHandler {
	return &GlobalGoalHandler{service: service}
}

// RegisterRoutes registers the routes for global goals.
func (h *GlobalGoalHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/goals/global", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/{goalID}", h.Get)
		r.Put("/{goalID}", h.Update)
		r.Delete("/{goalID}", h.Delete)
		r.Post("/{goalID}/execute-rules", h.ExecuteRules)
	})
}

// Create handles POST /api/v1/goals/global
func (h *GlobalGoalHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req servicegoal.CreateGlobalGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid request body", err))
		return
	}

	goal, err := h.service.CreateGlobalGoal(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, goal)
}

// List handles GET /api/v1/goals/global
func (h *GlobalGoalHandler) List(w http.ResponseWriter, r *http.Request) {
	cycleIDStr := r.URL.Query().Get("cycleId")
	var cycleID uuid.UUID
	if cycleIDStr != "" {
		var err error
		cycleID, err = uuid.Parse(cycleIDStr)
		if err != nil {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid cycle ID", err))
			return
		}
	}

	goals, err := h.service.ListGlobalGoals(r.Context(), cycleID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, goals)
}

// Get handles GET /api/v1/goals/global/{goalID}
func (h *GlobalGoalHandler) Get(w http.ResponseWriter, r *http.Request) {
	goalIDStr := chi.URLParam(r, "goalID")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid goal ID", err))
		return
	}

	goal, err := h.service.GetGlobalGoal(r.Context(), goalID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, goal)
}

// Update handles PUT /api/v1/goals/global/{goalID}
func (h *GlobalGoalHandler) Update(w http.ResponseWriter, r *http.Request) {
	goalIDStr := chi.URLParam(r, "goalID")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid goal ID", err))
		return
	}

	var req servicegoal.UpdateGlobalGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid request body", err))
		return
	}

	goal, err := h.service.UpdateGlobalGoal(r.Context(), goalID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, goal)
}

// Delete handles DELETE /api/v1/goals/global/{goalID}
func (h *GlobalGoalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	goalIDStr := chi.URLParam(r, "goalID")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid goal ID", err))
		return
	}

	if err := h.service.DeleteGlobalGoal(r.Context(), goalID); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ExecuteRules handles POST /api/v1/goals/global/{goalID}/execute-rules
func (h *GlobalGoalHandler) ExecuteRules(w http.ResponseWriter, r *http.Request) {
	goalIDStr := chi.URLParam(r, "goalID")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid goal ID", err))
		return
	}

	count, err := h.service.ExecuteRules(r.Context(), goalID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"assignments_created": count,
	})
}
