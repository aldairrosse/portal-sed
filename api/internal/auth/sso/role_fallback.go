package sso

import (
	"os"

	"github.com/sed-evaluacion-desempeno/api/internal/auth"
)

func isDevFallbackEnabled() bool {
	return os.Getenv("APP_ENV") == "development"
}

// ponytail: dev-only fallback — never evaluated in UAT/production
func Requires2FAFromLocalRole(role auth.Role) bool {
	if !isDevFallbackEnabled() {
		return false
	}
	switch role {
	case auth.RoleRH, auth.RoleDirector, auth.RoleDirectorGeneral:
		return true
	}
	return false
}
