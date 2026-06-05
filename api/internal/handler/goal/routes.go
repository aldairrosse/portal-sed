package goal

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"
)

// NewRouter creates a Chi router with all goals-api endpoints registered.
//
// Middleware stacks per endpoint:
//
//	GET    /api/v1/employees/{empId}/categories         → Auth → RateLimit(read)
//	POST   /api/v1/employees/{empId}/categories         → Auth → RateLimit(write) → Idempotency
//	PUT    /api/v1/employees/{empId}/categories/{catId}  → Auth → RateLimit(write) → Idempotency
//	DELETE /api/v1/employees/{empId}/categories/{catId}  → Auth → RateLimit(write)
//	POST   /api/v1/employees/{empId}/categories/{catId}/goals → Auth → RateLimit(write) → Idempotency
//	PUT    /api/v1/goals/{goalId}                        → Auth → RateLimit(write) → Idempotency
//	DELETE /api/v1/goals/{goalId}                        → Auth → RateLimit(write)
//	PATCH  /api/v1/goals/{goalId}/progress               → Auth → RateLimit(write)
//	POST   /api/v1/goals/batch                           → Auth → RateLimit(write) → Idempotency
//	POST   /api/v1/employees/{empId}/validate-weights    → Auth → RateLimit(read)
//	GET    /api/v1/kpis                                  → Auth → RateLimit(read)
//	POST   /api/v1/kpis                                  → Auth → RateLimit(write) → Idempotency
//	PUT    /api/v1/kpis/{kpiId}                          → Auth → RateLimit(write) → Idempotency
//	DELETE /api/v1/kpis/{kpiId}                          → Auth → RateLimit(write)
//	POST   /api/v1/goals/{goalId}/kpis                   → Auth → RateLimit(write) → Idempotency
//	DELETE /api/v1/goals/{goalId}/kpis/{kpiId}           → Auth → RateLimit(write)
//	GET    /api/v1/employees/{empId}/assignments         → Auth → RateLimit(read)
//	POST   /api/v1/employees/{empId}/assignments         → Auth → RateLimit(write) → Idempotency
func NewRouter(handler *GoalHandler) chi.Router {
	r := chi.NewRouter()

	// Shared middleware for auth (TODO(auth:C7): replace with real auth)
	r.Use(middleware.AuthPlaceholder)

	// Rate limit configurations
	readRateLimit := middleware.RateLimitConfig{
		Window:   time.Minute,
		MaxCount: 1000,
		Store:    middleware.NewInMemoryRateLimitStore(),
	}

	writeRateLimit := middleware.RateLimitConfig{
		Window:   time.Minute,
		MaxCount: 100,
		Store:    middleware.NewInMemoryRateLimitStore(),
	}

	// Idempotency middleware (in-memory for dev; Redis in production)
	idempStore := middleware.NewInMemoryIdempotencyStore()

	// --- Category endpoints ---

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Get("/api/v1/employees/{empId}/categories", handler.ListCategories)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/api/v1/employees/{empId}/categories", handler.CreateCategory)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Put("/api/v1/employees/{empId}/categories/{catId}", handler.UpdateCategory)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Delete("/api/v1/employees/{empId}/categories/{catId}", handler.DeleteCategory)
	})

	// --- Goal endpoints ---

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/api/v1/employees/{empId}/categories/{catId}/goals", handler.CreateGoal)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Put("/api/v1/goals/{goalId}", handler.UpdateGoal)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Delete("/api/v1/goals/{goalId}", handler.DeleteGoal)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Patch("/api/v1/goals/{goalId}/progress", handler.UpdateGoalProgress)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/api/v1/goals/batch", handler.BatchGoals)
	})

	// --- Weight validation endpoint ---

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Post("/api/v1/employees/{empId}/validate-weights", handler.ValidateWeights)
	})

	// --- KPI endpoints ---

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Get("/api/v1/kpis", handler.ListKPIs)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/api/v1/kpis", handler.CreateKPI)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Put("/api/v1/kpis/{kpiId}", handler.UpdateKPI)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Delete("/api/v1/kpis/{kpiId}", handler.DeleteKPI)
	})

	// --- KPI linking endpoints ---

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/api/v1/goals/{goalId}/kpis", handler.LinkKPI)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Delete("/api/v1/goals/{goalId}/kpis/{kpiId}", handler.UnlinkKPI)
	})

	// --- Assignment endpoints ---

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Get("/api/v1/employees/{empId}/assignments", handler.GetAssignment)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/api/v1/employees/{empId}/assignments", handler.CreateAssignment)
	})

	return r
}

// NewSubRouter creates a Chi router that can be mounted under an existing router.
// This is useful for composing with other API routers.
func NewSubRouter(handler *GoalHandler) http.Handler {
	return NewRouter(handler)
}
