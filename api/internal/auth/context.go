package auth

import (
	"context"

	"github.com/google/uuid"
)

// contextKey is used for storing auth-related values in request context.
type contextKey string

const (
	// SessionKey holds the authenticated session.
	SessionKey contextKey = "session"
	// EmployeeIDKey holds the authenticated employee's UUID.
	EmployeeIDKey contextKey = "employee_id"
	// RoleKey holds the authenticated employee's role.
	RoleKey contextKey = "role"
	// ProfileIDKey holds the authenticated employee's evaluation profile UUID.
	ProfileIDKey contextKey = "profile_id"
)

// WithSession stores session information in the context.
func WithSession(ctx context.Context, session *Session, role Role, profileID uuid.UUID) context.Context {
	ctx = context.WithValue(ctx, SessionKey, session)
	ctx = context.WithValue(ctx, EmployeeIDKey, session.EmployeeID)
	ctx = context.WithValue(ctx, RoleKey, role)
	ctx = context.WithValue(ctx, ProfileIDKey, profileID)
	return ctx
}

// GetSession retrieves the session from context.
func GetSession(ctx context.Context) (*Session, bool) {
	v, ok := ctx.Value(SessionKey).(*Session)
	return v, ok
}

// GetEmployeeID retrieves the employee UUID from context.
func GetEmployeeID(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(EmployeeIDKey).(uuid.UUID)
	return v, ok
}

// GetRole retrieves the role from context.
func GetRole(ctx context.Context) (Role, bool) {
	v, ok := ctx.Value(RoleKey).(Role)
	return v, ok
}

// GetProfileID retrieves the profile UUID from context.
func GetProfileID(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(ProfileIDKey).(uuid.UUID)
	return v, ok
}
