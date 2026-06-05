package auth

import (
	"github.com/go-chi/chi/v5"
)

// AuthRoutes creates a Chi router with all auth endpoints registered.
//
// Routes:
//
//	POST /login    — authenticate with email (dev mode)
//	POST /logout   — revoke current session
//	POST /refresh  — extend session expiry
//	GET  /me       — get current user info
//
// Expected mount point: /api/v1/auth
func AuthRoutes(handler *AuthHandler) chi.Router {
	r := chi.NewRouter()

	r.Post("/login", handler.Login)
	r.Post("/logout", handler.Logout)
	r.Post("/refresh", handler.Refresh)
	r.Get("/me", handler.Me)

	return r
}
