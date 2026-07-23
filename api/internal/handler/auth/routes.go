package auth

import (
	"github.com/go-chi/chi/v5"
)

// AuthRoutes creates a Chi router with all auth endpoints registered.
//
// Routes:
//
//	GET   /sso-login        — redirect to Keycloak SSO authorization
//	GET   /sso-callback     — handle OIDC callback from Keycloak
//	GET   /logout           — revoke current session and redirect to Keycloak logout
//	GET   /logout-complete   — redirects to frontend login after SSO logout completes
//	POST  /refresh          — extend session expiry
//	GET   /me               — get current user info
//	POST  /admin/revoke-employee/{empId} — revoke SSO tokens + local sessions
//
// Expected mount point: /api/v1/auth
func AuthRoutes(handler *AuthHandler) chi.Router {
	r := chi.NewRouter()

	r.Get("/sso-login", handler.SSOLoginRedirect)
	r.Get("/sso-callback", handler.SSOCallback)
	r.Get("/logout", handler.Logout)
	r.Get("/logout-complete", handler.LogoutComplete)
	r.Post("/refresh", handler.Refresh)
	r.Get("/me", handler.Me)
	r.Post("/admin/revoke-employee/{empId}", handler.RevokeEmployeeSessions)

	return r
}
