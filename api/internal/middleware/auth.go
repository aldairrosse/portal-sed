package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	svc "github.com/sed-evaluacion-desempeno/api/internal/service/auth"
)

// Context key types for backward-compatible context values.
type ctxKey string

const (
	employeeIDKey ctxKey = "employee_id"
	orgIDKey      ctxKey = "organization_id"
	rolesKey      ctxKey = "roles"
)

// OrgIDFromContext extracts the organization ID from the context.
func OrgIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(orgIDKey).(string)
	return v
}

// EmployeeIDFromContext extracts the employee ID as a string from the context.
// Deprecated: Use auth.GetEmployeeID for typed access.
func EmployeeIDFromContext(ctx context.Context) string {
	id, ok := auth.GetEmployeeID(ctx)
	if !ok {
		return ""
	}
	return id.String()
}

// RolesFromContext extracts the roles slice from the context.
// Deprecated: Use auth.GetRole for typed access.
func RolesFromContext(ctx context.Context) []string {
	role, ok := auth.GetRole(ctx)
	if !ok {
		return nil
	}
	return []string{string(role)}
}

// RequireAuth extracts the Bearer token from the Authorization header,
// validates the session, and injects auth info into the request context.
//
// Usage:
//
//	r.Use(middleware.RequireAuth(authSvc))
func RequireAuth(authSvc *svc.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := extractBearerToken(r)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				de := pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
					"missing or malformed Authorization header", err)
				ae := pkgerrors.NewAPIErrorResponse(de, "")
				_, _ = w.Write(ae.MustMarshalJSON())
				return
			}

			result, err := authSvc.ValidateSession(r.Context(), token)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				de := pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
					"invalid or expired session", err)
				ae := pkgerrors.NewAPIErrorResponse(de, "")
				_, _ = w.Write(ae.MustMarshalJSON())
				return
			}

			if result == nil || result.Session == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				de := pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
					"invalid or expired session", nil)
				ae := pkgerrors.NewAPIErrorResponse(de, "")
				_, _ = w.Write(ae.MustMarshalJSON())
				return
			}

			// Inject auth context
			ctx := auth.WithSession(r.Context(), result.Session, result.Role, result.ProfileID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission returns middleware that checks if the authenticated role
// has the required permission. Must be used after RequireAuth.
func RequirePermission(perm auth.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := auth.GetRole(r.Context())
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				de := pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
					"no authenticated session", nil)
				ae := pkgerrors.NewAPIErrorResponse(de, "")
				_, _ = w.Write(ae.MustMarshalJSON())
				return
			}

			if !auth.HasPermission(role, perm) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				de := pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
					"insufficient permissions", nil).
					WithDetails("required permission: " + string(perm))
				ae := pkgerrors.NewAPIErrorResponse(de, "")
				_, _ = w.Write(ae.MustMarshalJSON())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission returns middleware that checks if the authenticated role
// has any of the required permissions. Must be used after RequireAuth.
func RequireAnyPermission(perms ...auth.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := auth.GetRole(r.Context())
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				de := pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
					"no authenticated session", nil)
				ae := pkgerrors.NewAPIErrorResponse(de, "")
				_, _ = w.Write(ae.MustMarshalJSON())
				return
			}

			if !auth.HasAnyPermission(role, perms...) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				de := pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
					"insufficient permissions", nil)
				ae := pkgerrors.NewAPIErrorResponse(de, "")
				_, _ = w.Write(ae.MustMarshalJSON())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractBearerToken extracts the Bearer token from the Authorization header.
func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// Fallback: try the session_token cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			return "", err
		}
		return cookie.Value, nil
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errMissingAuth
	}

	return parts[1], nil
}

var errMissingAuth = pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
	"missing or malformed Authorization header", nil)

// AuthPlaceholder injects a mock identity into the request context.
// Deprecated: Use RequireAuth instead. Kept for backward compatibility
// with C2–C6 routes; will be removed in C8 (wire-api-replace-mocks).
func AuthPlaceholder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), employeeIDKey, "00000000-0000-0000-0000-000000000000")
		ctx = context.WithValue(ctx, orgIDKey, r.URL.Query().Get("organization_id"))
		ctx = context.WithValue(ctx, rolesKey, []string{"rh"})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
