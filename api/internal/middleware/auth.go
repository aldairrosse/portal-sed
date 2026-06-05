package middleware

import (
	"context"
	"net/http"
)

// context key types to avoid collisions.
type ctxKey string

const (
	employeeIDKey ctxKey = "employee_id"
	orgIDKey      ctxKey = "organization_id"
	rolesKey      ctxKey = "roles"
)

// AuthPlaceholder injects a mock identity into the request context.
// TODO(auth:C7): Replace with real session/JWT validation and claim extraction.
func AuthPlaceholder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), employeeIDKey, "00000000-0000-0000-0000-000000000000")
		ctx = context.WithValue(ctx, orgIDKey, r.URL.Query().Get("organization_id"))
		ctx = context.WithValue(ctx, rolesKey, []string{"rh"})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// EmployeeIDFromContext extracts the employee ID from the context.
func EmployeeIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(employeeIDKey).(string)
	return v
}

// OrgIDFromContext extracts the organization ID from the context.
func OrgIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(orgIDKey).(string)
	return v
}

// RolesFromContext extracts the roles slice from the context.
func RolesFromContext(ctx context.Context) []string {
	v, _ := ctx.Value(rolesKey).([]string)
	return v
}
