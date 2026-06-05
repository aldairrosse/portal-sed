package competency

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"
)

// Dependencies holds the shared resources needed by the competency handlers.
// In production, these would be wired via dependency injection.
type Dependencies struct {
	Handler *Handler
}

// RegisterRoutes mounts all 16 competency endpoints on the given router using
// chi's subrouter pattern. Middleware is applied per endpoint group:
//
//	GET endpoints:  AuthPlaceholder → RateLimit(read) → read replica
//	POST endpoints: AuthPlaceholder → RateLimit(write) → Idempotency
//	PUT endpoints:  AuthPlaceholder → RateLimit(write) → OptimisticLock
//	DELETE endpoints: AuthPlaceholder → RateLimit(write)
//
// TODO(auth:C7): Replace AuthPlaceholder with real RBAC middleware.
func RegisterRoutes(r chi.Router, deps *Dependencies) {
	// Rate-limit configurations
	readRateLimit := middleware.RateLimitConfig{
		Window:   time.Minute,
		MaxCount: 500,
		Store:    middleware.NewInMemoryRateLimitStore(),
	}

	writeRateLimit := middleware.RateLimitConfig{
		Window:   time.Minute,
		MaxCount: 50,
		Store:    middleware.NewInMemoryRateLimitStore(),
	}

	// Idempotency store (in-memory for dev; Redis in production)
	idempStore := middleware.NewInMemoryIdempotencyStore()

	// Common auth placeholder (applied to all routes below)
	r.Use(middleware.AuthPlaceholder)

	// -----------------------------------------------------------------------
	// Pillar endpoints
	// -----------------------------------------------------------------------

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/api/v1/pillars", deps.Handler.ListPillars)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/api/v1/pillars", deps.Handler.CreatePillar)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/api/v1/pillars/{id}", deps.Handler.GetPillar)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.OptimisticLock)
		r.Put("/api/v1/pillars/{id}", deps.Handler.UpdatePillar)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Delete("/api/v1/pillars/{id}", deps.Handler.DeletePillar)
	})

	// -----------------------------------------------------------------------
	// Competency endpoints (nested under pillars)
	// -----------------------------------------------------------------------

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/api/v1/pillars/{pillarId}/competencies", deps.Handler.ListCompetenciesByPillar)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/api/v1/pillars/{pillarId}/competencies", deps.Handler.CreateCompetency)
	})

	// -----------------------------------------------------------------------
	// Competency standalone endpoints
	// -----------------------------------------------------------------------

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/api/v1/competencies/{id}", deps.Handler.GetCompetency)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.OptimisticLock)
		r.Put("/api/v1/competencies/{id}", deps.Handler.UpdateCompetency)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Delete("/api/v1/competencies/{id}", deps.Handler.DeleteCompetency)
	})

	// -----------------------------------------------------------------------
	// Scale criteria endpoints
	// -----------------------------------------------------------------------

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/api/v1/competencies/{id}/scale-criteria", deps.Handler.GetScaleCriteria)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/api/v1/competencies/{id}/scale-criteria", deps.Handler.UpsertScaleCriteria)
	})

	// -----------------------------------------------------------------------
	// Static catalog endpoints
	// -----------------------------------------------------------------------

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/api/v1/levels", deps.Handler.ListLevels)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/api/v1/profiles", deps.Handler.ListProfiles)
	})

	// -----------------------------------------------------------------------
	// Acceptance level endpoints
	// -----------------------------------------------------------------------

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/api/v1/acceptance-levels", deps.Handler.ListAcceptanceLevels)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Post("/api/v1/acceptance-levels", deps.Handler.UpsertAcceptanceLevel)
	})
}

// readReplicaMiddleware sets the db role hint to "replica" so repository
// methods route GET queries to a read replica when available.
func readReplicaMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctxKeyDBRole, "replica")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// context key type for db role.
type ctxKeyDBRoleType struct{}

var ctxKeyDBRole = ctxKeyDBRoleType{}

// DBRoleFromContext extracts the db role from context.
func DBRoleFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyDBRole).(string)
	if v == "" {
		return "primary"
	}
	return v
}
