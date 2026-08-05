// Package dev provides a temporary authentication service for development
// without a real SSO provider. It exposes preset users with preconfigured
// roles and permissions to enable full-stack development.
//
// Only active when ENV=development. Not exposed in production.
package dev

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
)

// DevUser represents a preset development user with roles and permissions.
type DevUser struct {
	EmployeeID     string
	Email          string
	Name           string
	Profile        string
	OrganizationID string
	Roles          []string
	Permissions    []string
}

// presetUsers maps email to DevUser for quick lookup.
var presetUsers = map[string]*DevUser{
	"dev-rh@empresa.com": {
		EmployeeID:     "00000000-0000-0000-0000-000000000001",
		Email:          "dev-rh@empresa.com",
		Name:           "Frankil Perez",
		Profile:        "rh",
		OrganizationID: "00000000-0000-0000-0000-000000000001",
		Roles:          []string{"admin", "evaluator"},
		Permissions:    []string{"goal:read", "goal:write", "evaluation:read", "evaluation:write", "competency:read"},
	},
	"dev-jefe@empresa.com": {
		EmployeeID:     "00000000-0000-0000-0000-000000000002",
		Email:          "dev-jefe@empresa.com",
		Name:           "Juan Carlos",
		Profile:        "jefe",
		OrganizationID: "00000000-0000-0000-0000-000000000001",
		Roles:          []string{"manager", "evaluator"},
		Permissions:    []string{"goal:read", "goal:write", "evaluation:read", "evaluation:write"},
	},
	"dev-colaborador@empresa.com": {
		EmployeeID:     "00000000-0000-0000-0000-000000000003",
		Email:          "dev-colaborador@empresa.com",
		Name:           "Maria Lopez",
		Profile:        "colaborador",
		OrganizationID: "00000000-0000-0000-0000-000000000001",
		Roles:          []string{"collaborator"},
		Permissions:    []string{"goal:read", "evaluation:read"},
	},
}

// GetUserByEmail returns a preset DevUser for the given email.
// Returns nil if no preset user matches the email.
func GetUserByEmail(email string) *DevUser {
	return presetUsers[email]
}

// DevLoginResult holds the data returned after a dev login attempt.
type DevLoginResult struct {
	Session *auth.Session
	Token   string
	Role    auth.Role
}

// CreateDevSession creates a real database session for a dev user
// without going through the SSO flow. The session is stored in the
// sessions table and will be validated by RequireAuth middleware.
func CreateDevSession(ctx context.Context, sessionStore *auth.SessionStore, user *DevUser, ip, ua string) (*DevLoginResult, error) {
	empID, err := uuid.Parse(user.EmployeeID)
	if err != nil {
		return nil, err
	}

	// Map the dev user profile to an auth Role
	role := auth.ProfileNameToRole(user.Profile)

	// Create a real session in the sessions table
	session, token, err := sessionStore.Create(ctx, empID, ip, ua, "", "", "", "", false, time.Time{})
	if err != nil {
		return nil, err
	}

	return &DevLoginResult{
		Session: session,
		Token:   token,
		Role:    role,
	}, nil
}
