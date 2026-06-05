package evaluation

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/evaluation"
)

// NewRouter creates a Chi router with all evaluation and 9×9 endpoints registered.
//
// # Middleware stacks
//
//	GET    /api/v1/evaluations                    → AuthPlaceholder → RateLimit(read) → ReadReplica
//	GET    /api/v1/evaluations/{id}               → AuthPlaceholder → RateLimit(read) → ReadReplica
//	POST   /api/v1/evaluations/{id}/self-evaluation → AuthPlaceholder → RateLimit(write) → Idempotency
//	PUT    /api/v1/evaluations/{id}/self-evaluation → AuthPlaceholder → RateLimit(write) → OptimisticLock
//	POST   /api/v1/evaluations/{id}/rh-evaluation   → AuthPlaceholder → RateLimit(write) → Idempotency
//	PUT    /api/v1/evaluations/{id}/rh-evaluation   → AuthPlaceholder → RateLimit(write) → OptimisticLock
//	POST   /api/v1/evaluations/{id}/finalize        → AuthPlaceholder → RateLimit(write)
//	GET    /api/v1/evaluations/summary              → AuthPlaceholder → RateLimit(read) → ReadReplica
//	GET    /api/v1/nine-box/matrices                → AuthPlaceholder → RateLimit(read) → ReadReplica
//	POST   /api/v1/nine-box/matrices                → AuthPlaceholder → RateLimit(write)
//	GET    /api/v1/nine-box/matrices/{matrixId}     → AuthPlaceholder → RateLimit(read) → ReadReplica
//	GET    /api/v1/nine-box/matrices/{matrixId}/entries → AuthPlaceholder → RateLimit(read) → ReadReplica
//	POST   /api/v1/nine-box/matrices/{matrixId}/entries → AuthPlaceholder → RateLimit(write)
//	PUT    /api/v1/nine-box/entries/{entryId}       → AuthPlaceholder → RateLimit(write) → OptimisticLock
//	POST   /api/v1/nine-box/batch                   → AuthPlaceholder → RateLimit(write)
//	GET    /api/v1/nine-box/scales                  → AuthPlaceholder → RateLimit(read) → ReadReplica
//	GET    /api/v1/nine-box/quadrants               → AuthPlaceholder → RateLimit(read) → ReadReplica
func NewRouter(handler *EvaluationHandler) chi.Router {
	r := chi.NewRouter()

	// Shared middleware for auth (TODO(auth:C7): replace with real auth)
	r.Use(middleware.AuthPlaceholder)

	// Rate limit configurations
	readRateLimit := middleware.RateLimitConfig{
		Window:   time.Minute,
		MaxCount: 2000, // 2000 req/s per org for reads
		Store:    middleware.NewInMemoryRateLimitStore(),
	}

	writeRateLimit := middleware.RateLimitConfig{
		Window:   time.Minute,
		MaxCount: 100, // 100 req/s per org for writes
		Store:    middleware.NewInMemoryRateLimitStore(),
	}

	// Idempotency middleware (in-memory for dev; Redis in production)
	idempStore := middleware.NewInMemoryIdempotencyStore()

	// --- Evaluation Endpoints ---

	// GET /api/v1/evaluations
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/evaluations", handler.ListEvaluations)
	})

	// GET /api/v1/evaluations/{id}
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/evaluations/{id}", handler.GetEvaluation)
	})

	// POST /api/v1/evaluations/{id}/self-evaluation
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/evaluations/{id}/self-evaluation", handler.SubmitSelfEvaluation)
	})

	// PUT /api/v1/evaluations/{id}/self-evaluation
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.OptimisticLock)
		r.Put("/evaluations/{id}/self-evaluation", handler.UpdateSelfEvaluation)
	})

	// POST /api/v1/evaluations/{id}/rh-evaluation
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/evaluations/{id}/rh-evaluation", handler.SubmitRHEvaluation)
	})

	// PUT /api/v1/evaluations/{id}/rh-evaluation
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.OptimisticLock)
		r.Put("/evaluations/{id}/rh-evaluation", handler.UpdateRHEvaluation)
	})

	// POST /api/v1/evaluations/{id}/finalize
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Post("/evaluations/{id}/finalize", handler.FinalizeEvaluation)
	})

	// GET /api/v1/evaluations/summary
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/evaluations/summary", handler.GetEvaluationSummary)
	})

	// --- Nine-Box Endpoints ---

	// GET /api/v1/nine-box/matrices
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/nine-box/matrices", handler.ListMatrices)
	})

	// POST /api/v1/nine-box/matrices
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Post("/nine-box/matrices", handler.CreateMatrix)
	})

	// GET /api/v1/nine-box/matrices/{matrixId}
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/nine-box/matrices/{matrixId}", handler.GetMatrix)
	})

	// GET /api/v1/nine-box/matrices/{matrixId}/entries
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/nine-box/matrices/{matrixId}/entries", handler.ListMatrixEntries)
	})

	// POST /api/v1/nine-box/matrices/{matrixId}/entries
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Post("/nine-box/matrices/{matrixId}/entries", handler.UpsertMatrixEntry)
	})

	// PUT /api/v1/nine-box/entries/{entryId}
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.OptimisticLock)
		r.Put("/nine-box/entries/{entryId}", handler.UpdateEntry)
	})

	// POST /api/v1/nine-box/batch
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Post("/nine-box/batch", handler.BatchSubmitEntries)
	})

	// GET /api/v1/nine-box/scales
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/nine-box/scales", handler.GetScales)
	})

	// GET /api/v1/nine-box/quadrants
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/nine-box/quadrants", handler.GetQuadrants)
	})

	return r
}

// readReplicaMiddleware sets the db role hint to "replica" so repository
// methods route GET queries to a read replica when available.
func readReplicaMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := repo.WithDBRole(r.Context(), repo.DBRoleReplica)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
