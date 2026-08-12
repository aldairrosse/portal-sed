package goal

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	servicegoal "github.com/sed-evaluacion-desempeno/api/internal/service/goal"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// SharedGoalHandler handles HTTP requests for shared goals.
type SharedGoalHandler struct {
	service servicegoal.SharedGoalServicer
}

// NewSharedGoalHandler creates a new SharedGoalHandler.
func NewSharedGoalHandler(service servicegoal.SharedGoalServicer) *SharedGoalHandler {
	return &SharedGoalHandler{service: service}
}

// RegisterRoutes registers the routes for shared goals.
func (h *SharedGoalHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/goals/shared", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/{goalID}", h.Get)
		r.Put("/{goalID}", h.Update)
		r.Delete("/{goalID}", h.Delete)
		r.Post("/{goalID}/members", h.AddMember)
		r.Delete("/{goalID}/members/{employeeID}", h.RemoveMember)
		r.Put("/{goalID}/progress/{employeeID}", h.UpdateProgress)
	})
}

// Create handles POST /api/v1/goals/shared
func (h *SharedGoalHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req servicegoal.CreateSharedGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid request body", err))
		return
	}

	goal, err := h.service.CreateSharedGoal(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, goal)
}

// List handles GET /api/v1/goals/shared
func (h *SharedGoalHandler) List(w http.ResponseWriter, r *http.Request) {
	viewType := r.URL.Query().Get("view")
	
	var goals []interface{}
	var err error

	switch viewType {
	case "creator":
		creatorGoals, err := h.service.ListSharedGoalsAsCreator(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}
		for _, g := range creatorGoals {
			goals = append(goals, g)
		}
	case "member":
		memberGoals, err := h.service.ListSharedGoalsAsMember(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}
		for _, g := range memberGoals {
			goals = append(goals, g)
		}
	default:
		// List both
		creatorGoals, _ := h.service.ListSharedGoalsAsCreator(r.Context())
		memberGoals, _ := h.service.ListSharedGoalsAsMember(r.Context())
		for _, g := range creatorGoals {
			goals = append(goals, g)
		}
		for _, g := range memberGoals {
			goals = append(goals, g)
		}
	}

	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, goals)
}

// Get handles GET /api/v1/goals/shared/{goalID}
func (h *SharedGoalHandler) Get(w http.ResponseWriter, r *http.Request) {
	goalIDStr := chi.URLParam(r, "goalID")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid goal ID", err))
		return
	}

	goal, err := h.service.GetSharedGoal(r.Context(), goalID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, goal)
}

// Update handles PUT /api/v1/goals/shared/{goalID}
func (h *SharedGoalHandler) Update(w http.ResponseWriter, r *http.Request) {
	goalIDStr := chi.URLParam(r, "goalID")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid goal ID", err))
		return
	}

	var req servicegoal.UpdateSharedGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid request body", err))
		return
	}

	goal, err := h.service.UpdateSharedGoal(r.Context(), goalID, req)
	if err != nil {
		if err == servicegoal.ErrNotCreator {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "Only the creator can modify this shared goal", err))
			return
		}
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, goal)
}

// Delete handles DELETE /api/v1/goals/shared/{goalID}
func (h *SharedGoalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	goalIDStr := chi.URLParam(r, "goalID")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid goal ID", err))
		return
	}

	if err := h.service.DeleteSharedGoal(r.Context(), goalID); err != nil {
		if err == servicegoal.ErrNotCreator {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "Only the creator can delete this shared goal", err))
			return
		}
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddMember handles POST /api/v1/goals/shared/{goalID}/members
func (h *SharedGoalHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	goalIDStr := chi.URLParam(r, "goalID")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid goal ID", err))
		return
	}

	var req servicegoal.AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid request body", err))
		return
	}

	member, err := h.service.AddMember(r.Context(), goalID, req)
	if err != nil {
		if err == servicegoal.ErrNotCreator {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "Only the creator can add members", err))
			return
		}
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, member)
}

// RemoveMember handles DELETE /api/v1/goals/shared/{goalID}/members/{employeeID}
func (h *SharedGoalHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	goalIDStr := chi.URLParam(r, "goalID")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid goal ID", err))
		return
	}

	employeeIDStr := chi.URLParam(r, "employeeID")
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid employee ID", err))
		return
	}

	if err := h.service.RemoveMember(r.Context(), goalID, employeeID); err != nil {
		if err == servicegoal.ErrNotCreator {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "Only the creator can remove members", err))
			return
		}
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateProgress handles PUT /api/v1/goals/shared/{goalID}/progress/{employeeID}
func (h *SharedGoalHandler) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	goalIDStr := chi.URLParam(r, "goalID")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid goal ID", err))
		return
	}

	employeeIDStr := chi.URLParam(r, "employeeID")
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid employee ID", err))
		return
	}

	var req servicegoal.UpdateProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "Invalid request body", err))
		return
	}

	if err := h.service.UpdateProgress(r.Context(), goalID, employeeID, req); err != nil {
		if err == servicegoal.ErrNotCreator {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "Only the creator can update progress", err))
			return
		}
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "updated",
	})
}
