package dev

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/sed-evaluacion-desempeno/api/internal/auth"
)

// Routes returns a chi router with dev-only endpoints.
// Handlers themselves guard with 404 when not in development, so mounting
// unconditionally is safe; caller may also conditionally mount for extra safety.
func Routes(db *sql.DB, sessionStore *auth.SessionStore) http.Handler {
	h := NewHandler(db, sessionStore)
	r := chi.NewRouter()
	r.Get("/status", h.Status)
	r.Get("/employees", h.ListEmployees)
	r.Post("/impersonate", h.Impersonate)
	return r
}
