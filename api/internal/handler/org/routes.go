package org

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/org"
)

// NewRouter creates a Chi router with all org hierarchy endpoints registered.
//
// Endpoint catalog (14 endpoints):
//
//   GET    /api/v1/org-trees                       → ListOrgTrees
//   GET    /api/v1/org-trees/{treeId}              → GetOrgTree
//   GET    /api/v1/org-trees/{treeId}/nodes         → GetOrgTreeNodes
//   GET    /api/v1/org-trees/{treeId}/export        → ExportOrgTree
//   GET    /api/v1/org-nodes/{nodeId}               → GetOrgNode
//   POST   /api/v1/org-nodes                       → CreateOrgNode
//   PUT    /api/v1/org-nodes/{nodeId}               → UpdateOrgNode
//   DELETE /api/v1/org-nodes/{nodeId}               → DeleteOrgNode
//   POST   /api/v1/org-nodes/{nodeId}/move          → MoveOrgNode
//   GET    /api/v1/employees                        → ListEmployees
//   GET    /api/v1/employees/{empId}                → GetEmployee
//   GET    /api/v1/employees/{empId}/evaluatees     → GetMyEvaluatees
//   GET    /api/v1/employees/{empId}/manager        → GetManager
//   GET    /api/v1/employees/{empId}/ancestors      → GetAncestors
//   POST   /api/v1/employees/batch                 → BatchLookupEmployees
//   GET    /api/v1/employees/search                → SearchEmployees
//   GET    /api/v1/evaluator-scopes                 → GetEvaluatorScope
//   GET    /api/v1/evaluator-scopes/{scopeId}       → GetEvaluatorScopeByID
func NewRouter(handler *OrgHandler) chi.Router {
	r := chi.NewRouter()
	RegisterRoutes(r, handler)
	return r
}

// RegisterRoutes registers all org hierarchy endpoints on an existing router.
func RegisterRoutes(r chi.Router, handler *OrgHandler) {

	// Rate limit configurations
	readRateLimit := middleware.RateLimitConfig{
		Window:   time.Minute,
		MaxCount: 2000, // 2000 req/s for reads
		Store:    middleware.NewInMemoryRateLimitStore(),
	}

	writeRateLimit := middleware.RateLimitConfig{
		Window:   time.Minute,
		MaxCount: 100, // 100 req/s for writes
		Store:    middleware.NewInMemoryRateLimitStore(),
	}

	// --- Organization Tree endpoints (read) ---

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/org-trees", handler.ListOrgTrees)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/org-trees/{treeId}", handler.GetOrgTree)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/org-trees/{treeId}/nodes", handler.GetOrgTreeNodes)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/org-trees/{treeId}/export", handler.ExportOrgTree)
	})

	// --- Org Node endpoints (read + write) ---

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/org-nodes/{nodeId}", handler.GetOrgNode)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Post("/org-nodes", handler.CreateOrgNode)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Put("/org-nodes/{nodeId}", handler.UpdateOrgNode)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Delete("/org-nodes/{nodeId}", handler.DeleteOrgNode)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Post("/org-nodes/{nodeId}/move", handler.MoveOrgNode)
	})

	// --- Employee endpoints (read + batch) ---

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/employees", handler.ListEmployees)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/employees/{empId}", handler.GetEmployee)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/employees/{empId}/evaluatees", handler.GetMyEvaluatees)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/employees/{empId}/manager", handler.GetManager)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/employees/{empId}/ancestors", handler.GetAncestors)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(writeRateLimit))
		r.Post("/employees/batch", handler.BatchLookupEmployees)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/employees/search", handler.SearchEmployees)
	})

	// --- Evaluator Scope endpoints (read) ---

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/evaluator-scopes", handler.GetEvaluatorScope)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(readRateLimit))
		r.Use(readReplicaMiddleware)
		r.Get("/evaluator-scopes/{scopeId}", handler.GetEvaluatorScopeByID)
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
