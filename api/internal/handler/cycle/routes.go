package cycle

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/cycle"
)

// NewRouter creates a Chi router with all cycle/phase endpoints registered.
//
// Middleware stacks per endpoint (from design section 3.1):
//
//	GET    /api/v1/cycles                  → AuthPlaceholder → RateLimit(read) → ReadReplica
//	POST   /api/v1/cycles                  → AuthPlaceholder → RateLimit(write) → Idempotency
//	GET    /api/v1/cycles/{id}             → AuthPlaceholder → RateLimit(read) → ReadReplica
//	PUT    /api/v1/cycles/{id}/transition  → AuthPlaceholder → RateLimit(write) → Idempotency → OptimisticLock
//	GET    /api/v1/phases                  → AuthPlaceholder → RateLimit(read) → ReadReplica
//	GET    /api/v1/cycles/{id}/transitions → AuthPlaceholder → RateLimit(read) → ReadReplica
func NewRouter(handler *CycleHandler) chi.Router {
	r := chi.NewRouter()
	RegisterRoutes(r, handler)
	return r
}

// RegisterRoutes registers all cycle/phase endpoints on an existing router.
// The caller is responsible for applying AuthPlaceholder and any other shared middleware.
func RegisterRoutes(r chi.Router, handler *CycleHandler) {

	// Rate limit configurations
	readRateLimit := middleware.RateLimitConfig{
		Window:   time.Minute,
		MaxCount: 1000, // 1000 req/s per org for reads
		Store:    middleware.NewInMemoryRateLimitStore(),
	}

	writeRateLimit := middleware.RateLimitConfig{
		Window:   time.Minute,
		MaxCount: 100, // 100 req/s per org for writes
		Store:    middleware.NewInMemoryRateLimitStore(),
	}

	// Idempotency middleware (in-memory for dev; Redis in production)
	idempStore := middleware.NewInMemoryIdempotencyStore()

	// --- Cycle endpoints ---

	// GET /api/v1/cycles
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/cycles", handler.ListCycles)
	})

	// POST /api/v1/cycles
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/cycles", handler.CreateCycle)
	})

	// GET /api/v1/cycles/{id}
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/cycles/{id}", handler.GetCycle)
	})

	// PUT /api/v1/cycles/{id}/transition
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Use(middleware.OptimisticLock)
		r.Put("/cycles/{id}/transition", handler.TransitionPhase)
	})

	// --- Phase endpoints ---

	// GET /api/v1/phases
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/phases", handler.GetPhaseDefinitions)
	})

	// GET /api/v1/cycles/{id}/transitions
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/cycles/{id}/transitions", handler.GetAvailableTransitions)
	})
}

// readReplicaMiddleware sets the db role hint to "replica" so repository
// methods route GET queries to a read replica when available.
func readReplicaMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := repo.WithDBRole(r.Context(), repo.DBRoleReplica)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
