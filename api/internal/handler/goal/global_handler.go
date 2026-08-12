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
		errors.WriteError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	goal, err := h.service.CreateGlobalGoal(r.Context(), req)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, "Failed to create global goal", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(goal)
}

// List handles GET /api/v1/goals/global
func (h *GlobalGoalHandler) List(w http.ResponseWriter, r *http.Request) {
	cycleIDStr := r.URL.Query().Get("cycleId")
	var cycleID uuid.UUID
	if cycleIDStr != "" {
		var err error
		cycleID, err = uuid.Parse(cycleIDStr)
		if err != nil {
			errors.WriteError(w, http.StatusBadRequest, "Invalid cycle ID", err)
			return
		}
	}

	goals, err := h.service.ListGlobalGoals(r.Context(), cycleID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, "Failed to list global goals", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goals)
}

// Get handles GET /api/v1/goals/global/{goalID}
func (h *GlobalGoalHandler) Get(w http.ResponseWriter, r *http.Request) {
	goalIDStr := chi.URLParam(r, "goalID")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		errors.WriteError(w, http.StatusBadRequest, "Invalid goal ID", err)
		return
	}

	goal, err := h.service.GetGlobalGoal(r.Context(), goalID)
	if err != nil {
		errors.WriteError(w, http.StatusNotFound, "Global goal not found", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goal)
}

// Update handles PUT /api/v1/goals/global/{goalID}
func (h *GlobalGoalHandler) Update(w http.ResponseWriter, r *http.Request) {
	goalIDStr := chi.URLParam(r, "goalID")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		errors.WriteError(w, http.StatusBadRequest, "Invalid goal ID", err)
		return
	}

	var req servicegoal.UpdateGlobalGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	goal, err := h.service.UpdateGlobalGoal(r.Context(), goalID, req)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, "Failed to update global goal", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goal)
}

// Delete handles DELETE /api/v1/goals/global/{goalID}
func (h *GlobalGoalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	goalIDStr := chi.URLParam(r, "goalID")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		errors.WriteError(w, http.StatusBadRequest, "Invalid goal ID", err)
		return
	}

	if err := h.service.DeleteGlobalGoal(r.Context(), goalID); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, "Failed to delete global goal", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ExecuteRules handles POST /api/v1/goals/global/{goalID}/execute-rules
func (h *GlobalGoalHandler) ExecuteRules(w http.ResponseWriter, r *http.Request) {
	goalIDStr := chi.URLParam(r, "goalID")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		errors.WriteError(w, http.StatusBadRequest, "Invalid goal ID", err)
		return
	}

	count, err := h.service.ExecuteRules(r.Context(), goalID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, "Failed to execute rules", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"assignments_created": count,
	})
}
