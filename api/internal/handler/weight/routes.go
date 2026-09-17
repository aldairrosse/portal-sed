package weight

import (
	"github.com/go-chi/chi/v5"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"
	authsvc "github.com/sed-evaluacion-desempeno/api/internal/service/auth"
)

func RegisterRoutes(r chi.Router, h *Handler, authSvc *authsvc.AuthService) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(authSvc))
		r.Use(middleware.RequireLoA2())
		// Authenticated GETs: any user can view weight config.
		r.Get("/weights/cycle-config", h.GetCycleConfig)
		r.Get("/weights/team-config", h.GetTeamConfig)
		// RH only: G% (P derived)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalGlobal))
			r.Put("/weights/cycle-config", h.PutCycleConfig)
		})
		// Jefe / director: J% (PJ derived) — team_id query param optional, fallback to caller's node
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyPermission(auth.PermGoalShared, auth.PermGoalGlobal))
			r.Put("/weights/team-config", h.PutTeamConfig)
		})
	})
}
