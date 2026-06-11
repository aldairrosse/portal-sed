package goal

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"
	authsvc "github.com/sed-evaluacion-desempeno/api/internal/service/auth"
)

// NewRouter creates a Chi router with all goals-api endpoints registered.
//
// Middleware stacks per endpoint:
//
//	GET    /api/v1/employees/{empId}/categories         → RequireAuth → RequirePermission(read) → RateLimit(read)
//	POST   /api/v1/employees/{empId}/categories         → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
//	PUT    /api/v1/employees/{empId}/categories/{catId}  → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
//	DELETE /api/v1/employees/{empId}/categories/{catId}  → RequireAuth → RequirePermission(write) → RateLimit(write)
//	POST   /api/v1/employees/{empId}/categories/{catId}/goals → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
//	PUT    /api/v1/goals/{goalId}                        → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
//	DELETE /api/v1/goals/{goalId}                        → RequireAuth → RequirePermission(write) → RateLimit(write)
//	PATCH  /api/v1/goals/{goalId}/progress               → RequireAuth → RequirePermission(write) → RateLimit(write)
//	POST   /api/v1/goals/batch                           → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
//	POST   /api/v1/employees/{empId}/validate-weights    → RequireAuth → RequirePermission(read) → RateLimit(read)
//	GET    /api/v1/kpis                                  → RequireAuth → RequirePermission(read) → RateLimit(read)
//	POST   /api/v1/kpis                                  → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
//	PUT    /api/v1/kpis/{kpiId}                          → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
//	DELETE /api/v1/kpis/{kpiId}                          → RequireAuth → RequirePermission(write) → RateLimit(write)
//	POST   /api/v1/goals/{goalId}/kpis                   → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
//	DELETE /api/v1/goals/{goalId}/kpis/{kpiId}           → RequireAuth → RequirePermission(write) → RateLimit(write)
//	GET    /api/v1/employees/{empId}/assignments         → RequireAuth → RequirePermission(read) → RateLimit(read)
//	POST   /api/v1/employees/{empId}/assignments         → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
func NewRouter(handler *GoalHandler, authSvc *authsvc.AuthService) chi.Router {
	r := chi.NewRouter()

	// Shared middleware for auth
	r.Use(middleware.RequireAuth(authSvc))

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

	// writePerms defines the set of permissions required for write operations.
	writePerms := []auth.Permission{auth.PermGoalCreate, auth.PermGoalUpdate, auth.PermGoalDelete, auth.PermGoalProgress}

	// --- Category endpoints ---

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequirePermission(auth.PermGoalRead))
		r.Use(middleware.RateLimit(readRateLimit))
		r.Get("/api/v1/employees/{empId}/categories", handler.ListCategories)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAnyPermission(writePerms...))
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/api/v1/employees/{empId}/categories", handler.CreateCategory)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAnyPermission(writePerms...))
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Put("/api/v1/employees/{empId}/categories/{catId}", handler.UpdateCategory)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAnyPermission(writePerms...))
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Delete("/api/v1/employees/{empId}/categories/{catId}", handler.DeleteCategory)
	})

	// --- Goal endpoints ---

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAnyPermission(writePerms...))
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/api/v1/employees/{empId}/categories/{catId}/goals", handler.CreateGoal)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAnyPermission(writePerms...))
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Put("/api/v1/goals/{goalId}", handler.UpdateGoal)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAnyPermission(writePerms...))
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Delete("/api/v1/goals/{goalId}", handler.DeleteGoal)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAnyPermission(writePerms...))
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Patch("/api/v1/goals/{goalId}/progress", handler.UpdateGoalProgress)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAnyPermission(writePerms...))
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/api/v1/goals/batch", handler.BatchGoals)
	})

	// --- Weight validation endpoint ---

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequirePermission(auth.PermGoalRead))
		r.Use(middleware.RateLimit(readRateLimit))
		r.Post("/api/v1/employees/{empId}/validate-weights", handler.ValidateWeights)
	})

	// --- KPI endpoints ---

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequirePermission(auth.PermGoalRead))
		r.Use(middleware.RateLimit(readRateLimit))
		r.Get("/api/v1/kpis", handler.ListKPIs)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAnyPermission(writePerms...))
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/api/v1/kpis", handler.CreateKPI)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAnyPermission(writePerms...))
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Put("/api/v1/kpis/{kpiId}", handler.UpdateKPI)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAnyPermission(writePerms...))
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Delete("/api/v1/kpis/{kpiId}", handler.DeleteKPI)
	})

	// --- KPI linking endpoints ---

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAnyPermission(writePerms...))
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/api/v1/goals/{goalId}/kpis", handler.LinkKPI)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAnyPermission(writePerms...))
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Delete("/api/v1/goals/{goalId}/kpis/{kpiId}", handler.UnlinkKPI)
	})

	// --- Assignment endpoints ---

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequirePermission(auth.PermGoalRead))
		r.Use(middleware.RateLimit(readRateLimit))
		r.Get("/api/v1/employees/{empId}/assignments", handler.GetAssignment)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAnyPermission(writePerms...))
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
		r.Post("/api/v1/employees/{empId}/assignments", handler.CreateAssignment)
	})

	return r
}

// NewSubRouter creates a Chi router that can be mounted under an existing router.
// This is useful for composing with other API routers.
func NewSubRouter(handler *GoalHandler, authSvc *authsvc.AuthService) http.Handler {
	return NewRouter(handler, authSvc)
}
