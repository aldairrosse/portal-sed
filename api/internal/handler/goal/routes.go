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
// Expected mount point: /api/v1
//
// Middleware stacks per endpoint:
//
//	GET    /employees/{empId}/categories         → RequireAuth → RequirePermission(read) → RateLimit(read)
//	POST   /employees/{empId}/categories         → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
//	PUT    /employees/{empId}/categories/{catId}  → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
//	DELETE /employees/{empId}/categories/{catId}  → RequireAuth → RequirePermission(write) → RateLimit(write)
//	POST   /employees/{empId}/categories/{catId}/goals → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
//	PUT    /goals/{goalId}                        → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
//	DELETE /goals/{goalId}                        → RequireAuth → RequirePermission(write) → RateLimit(write)
//	PATCH  /goals/{goalId}/progress               → RequireAuth → RequirePermission(write) → RateLimit(write)
//	POST   /goals/batch                           → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
//	POST   /employees/{empId}/validate-weights    → RequireAuth → RequirePermission(read) → RateLimit(read)
//	GET    /kpis                                  → RequireAuth → RequirePermission(read) → RateLimit(read)
//	POST   /kpis                                  → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
//	PUT    /kpis/{kpiId}                          → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
//	DELETE /kpis/{kpiId}                          → RequireAuth → RequirePermission(write) → RateLimit(write)
//	PATCH  /kpis/{kpiId}/value                    → RequireAuth → RequirePermission(write) → RateLimit(write)
//	POST   /goals/{goalId}/kpis                   → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
//	DELETE /goals/{goalId}/kpis/{kpiId}           → RequireAuth → RequirePermission(write) → RateLimit(write)
//	GET    /employees/{empId}/assignments         → RequireAuth → RequirePermission(read) → RateLimit(read)
//	GET    /employees/{empId}/score               → RequireAuth → RequirePermission(read) → RateLimit(read)
//	POST   /employees/{empId}/assignments         → RequireAuth → RequirePermission(write) → RateLimit(write) → Idempotency
func NewRouter(handler *GoalHandler, authSvc *authsvc.AuthService) chi.Router {
	r := chi.NewRouter()
	RegisterRoutes(r, handler, authSvc)
	return r
}

// RegisterRoutes registers all goals-api endpoints on an existing router.
// Use this when the router is already mounted under a prefix like /api/v1.
// All middleware is scoped inside r.Group to avoid Chi's "middleware after routes" panic
// when the caller has already registered other handlers on the same router.
func RegisterRoutes(r chi.Router, handler *GoalHandler, authSvc *authsvc.AuthService) {

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

	// All goal endpoints share the RequireAuth guard, scoped inside a Group
	// so middleware can be added after other handlers have registered on the parent router.
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(authSvc))

		// --- Category endpoints ---

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalRead))
			r.Use(middleware.RateLimit(readRateLimit))
			r.Get("/employees/{empId}/categories", handler.ListCategories)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyPermission(writePerms...))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
			r.Post("/employees/{empId}/categories", handler.CreateCategory)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyPermission(writePerms...))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
			r.Put("/employees/{empId}/categories/{catId}", handler.UpdateCategory)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyPermission(writePerms...))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Delete("/employees/{empId}/categories/{catId}", handler.DeleteCategory)
		})

		// --- Goal endpoints ---

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyPermission(writePerms...))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
			r.Post("/employees/{empId}/categories/{catId}/goals", handler.CreateGoal)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyPermission(writePerms...))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
			r.Put("/goals/{goalId}", handler.UpdateGoal)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyPermission(writePerms...))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Delete("/goals/{goalId}", handler.DeleteGoal)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyPermission(writePerms...))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Patch("/goals/{goalId}/progress", handler.UpdateGoalProgress)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyPermission(writePerms...))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
			r.Post("/goals/batch", handler.BatchGoals)
		})

		// --- Weight validation endpoint ---

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalRead))
			r.Use(middleware.RateLimit(readRateLimit))
			r.Post("/employees/{empId}/validate-weights", handler.ValidateWeights)
		})

		// --- KPI endpoints ---

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalRead))
			r.Use(middleware.RateLimit(readRateLimit))
			r.Get("/kpis", handler.ListKPIs)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyPermission(writePerms...))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
			r.Post("/kpis", handler.CreateKPI)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyPermission(writePerms...))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
			r.Put("/kpis/{kpiId}", handler.UpdateKPI)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyPermission(writePerms...))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Delete("/kpis/{kpiId}", handler.DeleteKPI)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyPermission(writePerms...))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Patch("/kpis/{kpiId}/value", handler.UpdateKPIValue)
		})

		// --- KPI linking endpoints ---

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyPermission(writePerms...))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
			r.Post("/goals/{goalId}/kpis", handler.LinkKPI)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyPermission(writePerms...))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Delete("/goals/{goalId}/kpis/{kpiId}", handler.UnlinkKPI)
		})

		// --- Scoring endpoints ---

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalRead))
			r.Use(middleware.RateLimit(readRateLimit))
			r.Get("/employees/{empId}/score", handler.GetEmployeeScore)
		})

		// --- Assignment endpoints ---

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalRead))
			r.Use(middleware.RateLimit(readRateLimit))
			r.Get("/employees/{empId}/assignments", handler.GetAssignment)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyPermission(writePerms...))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Use(middleware.Idempotency(idempStore, 24*time.Hour))
			r.Post("/employees/{empId}/assignments", handler.CreateAssignment)
		})
	})
}

// NewSubRouter creates a Chi router that can be mounted under an existing router.
// This is useful for composing with other API routers.
func NewSubRouter(handler *GoalHandler, authSvc *authsvc.AuthService) http.Handler {
	return NewRouter(handler, authSvc)
}
