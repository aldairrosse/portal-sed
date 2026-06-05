// Package auth provides session management, role-based access control (RBAC),
// and context helpers for the SED evaluation platform.
package auth

// Role represents an evaluation profile with associated permissions.
type Role string

const (
	RoleColaborador    Role = "colaborador"
	RoleJefe           Role = "jefe"
	RoleVendedor       Role = "vendedor"
	RoleGerenteTienda  Role = "gerente-tienda"
	RoleDivisional     Role = "divisional"
	RoleRegional       Role = "regional"
	RoleDirector       Role = "director"
	RoleDirectorGeneral Role = "director-general"
	RoleRH             Role = "rh"
)

// Permission represents a specific action that can be authorized.
type Permission string

const (
	// Goal permissions
	PermGoalCreate   Permission = "goal:create"
	PermGoalRead     Permission = "goal:read"
	PermGoalUpdate   Permission = "goal:update"
	PermGoalDelete   Permission = "goal:delete"
	PermGoalProgress Permission = "goal:progress"

	// Competency permissions
	PermCompetencyRead   Permission = "competency:read"
	PermCompetencyWrite  Permission = "competency:write"
	PermCompetencyDelete Permission = "competency:delete"

	// Evaluation permissions
	PermEvalSelf Permission = "eval:self"
	PermEvalRH   Permission = "eval:rh"
	PermEval9x9  Permission = "eval:9x9"
	PermEvalRead Permission = "eval:read"

	// Cycle permissions
	PermCycleRead       Permission = "cycle:read"
	PermCycleTransition Permission = "cycle:transition"

	// Org permissions
	PermOrgRead  Permission = "org:read"
	PermOrgWrite Permission = "org:write"

	// Admin permissions
	PermAdminAll Permission = "admin:all"
)

// RolePermissions maps roles to their allowed permissions.
var RolePermissions = map[Role][]Permission{
	RoleColaborador: {
		PermGoalCreate, PermGoalRead, PermGoalUpdate, PermGoalDelete, PermGoalProgress,
		PermCompetencyRead,
		PermEvalSelf, PermEvalRead,
		PermCycleRead, PermOrgRead,
	},
	RoleJefe: {
		PermGoalRead,
		PermCompetencyRead,
		PermEval9x9, PermEvalRead,
		PermCycleRead, PermOrgRead,
	},
	RoleVendedor: {
		PermGoalCreate, PermGoalRead, PermGoalUpdate, PermGoalDelete, PermGoalProgress,
		PermCompetencyRead,
		PermEvalSelf, PermEvalRead,
		PermCycleRead, PermOrgRead,
	},
	RoleGerenteTienda: {
		PermGoalRead,
		PermCompetencyRead,
		PermEval9x9, PermEvalRead,
		PermCycleRead, PermOrgRead,
	},
	RoleDivisional: {
		PermGoalRead,
		PermCompetencyRead,
		PermEval9x9, PermEvalRead,
		PermCycleRead, PermOrgRead,
	},
	RoleRegional: {
		PermGoalRead,
		PermCompetencyRead,
		PermEval9x9, PermEvalRead,
		PermCycleRead, PermOrgRead,
	},
	RoleDirector: {
		PermGoalRead,
		PermCompetencyRead,
		PermEval9x9, PermEvalRead,
		PermCycleRead, PermCycleTransition,
		PermOrgRead, PermOrgWrite,
	},
	RoleDirectorGeneral: {
		PermGoalRead,
		PermCompetencyRead,
		PermEvalRead,
		PermCycleRead, PermCycleTransition,
		PermOrgRead, PermOrgWrite,
		PermAdminAll,
	},
	RoleRH: {
		PermGoalCreate, PermGoalRead, PermGoalUpdate, PermGoalDelete, PermGoalProgress,
		PermCompetencyRead, PermCompetencyWrite, PermCompetencyDelete,
		PermEvalSelf, PermEvalRH, PermEvalRead,
		PermCycleRead, PermCycleTransition,
		PermOrgRead,
	},
}

// HasPermission checks if a role has a specific permission.
func HasPermission(role Role, perm Permission) bool {
	perms, ok := RolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if a role has any of the given permissions.
func HasAnyPermission(role Role, perms ...Permission) bool {
	for _, perm := range perms {
		if HasPermission(role, perm) {
			return true
		}
	}
	return false
}

// ProfileNameToRole maps an evaluation profile name to a Role constant.
// Returns RoleColaborador as fallback for unknown profiles.
func ProfileNameToRole(name string) Role {
	switch name {
	case "colaborador":
		return RoleColaborador
	case "jefe":
		return RoleJefe
	case "vendedor":
		return RoleVendedor
	case "gerente-tienda":
		return RoleGerenteTienda
	case "divisional":
		return RoleDivisional
	case "regional":
		return RoleRegional
	case "director":
		return RoleDirector
	case "director-general":
		return RoleDirectorGeneral
	case "rh":
		return RoleRH
	default:
		return RoleColaborador
	}
}
