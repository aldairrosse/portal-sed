package sso

import "github.com/sed-evaluacion-desempeno/api/internal/auth"

// ponytail: local role mapping for dev/local envs without Keycloak
func Requires2FAFromLocalRole(role auth.Role) bool {
	switch role {
	case auth.RoleRH, auth.RoleDirector, auth.RoleDirectorGeneral:
		return true
	}
	return false
}
