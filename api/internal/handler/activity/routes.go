package activity

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"
	authsvc "github.com/sed-evaluacion-desempeno/api/internal/service/auth"
)

// RegisterActivityRoutes registers all activity log endpoints on the given router.
//
// Middleware stack:
//
//	GET /employees/{employeeId}/activity-logs → RequireAuth → RateLimit(read)
func RegisterActivityRoutes(r chi.Router, handler *ActivityHandler, authSvc *authsvc.AuthService) {
	// Rate limit configuration for reads
	readRateLimit := middleware.RateLimitConfig{
		Window:   time.Minute,
		MaxCount: 2000,
		Store:    middleware.NewInMemoryRateLimitStore(),
	}

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(authSvc))

		r.Group(func(r chi.Router) {
			r.Use(middleware.RateLimit(readRateLimit))
			r.Get("/employees/{employeeId}/activity-logs", handler.ListActivityLogs)
		})
	})
}

// NewRouter creates a Chi router with all activity log endpoints registered.
func NewRouter(handler *ActivityHandler, authSvc *authsvc.AuthService) chi.Router {
	r := chi.NewRouter()
	RegisterActivityRoutes(r, handler, authSvc)
	return r
}
