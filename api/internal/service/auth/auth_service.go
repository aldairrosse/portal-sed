// Package auth provides the authentication service layer for the SED platform.
// It handles login, session validation, logout, and refresh operations.
package auth

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// EmployeeRow is a lightweight read model for employee data used during auth.
type EmployeeRow struct {
	ID         uuid.UUID
	FirstName  string
	LastName   string
	Email      string
	ProfileID  uuid.UUID
	IsActive   bool
	JobTitle   string
	OrgNodeID  uuid.UUID
}

// EmployeeReader defines the data access interface for employee lookups
// during authentication. This allows testing without a real database.
type EmployeeReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*EmployeeRow, error)
	GetByEmail(ctx context.Context, email string) (*EmployeeRow, error)
}

// employeeReader is the production implementation of EmployeeReader
// backed by direct SQL queries.
type employeeReader struct {
	db *sql.DB
}

// NewEmployeeReader creates an EmployeeReader backed by the given database.
func NewEmployeeReader(db *sql.DB) EmployeeReader {
	return &employeeReader{db: db}
}

func (r *employeeReader) GetByID(ctx context.Context, id uuid.UUID) (*EmployeeRow, error) {
	row := &EmployeeRow{}
	var jobTitle sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, first_name, last_name, email, profile_id, is_active, job_title, org_node_id
		 FROM employees WHERE id = $1`, id,
	).Scan(&row.ID, &row.FirstName, &row.LastName, &row.Email, &row.ProfileID, &row.IsActive, &jobTitle, &row.OrgNodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrEmployeeNotFound
		}
		return nil, err
	}
	row.JobTitle = jobTitle.String
	return row, nil
}

func (r *employeeReader) GetByEmail(ctx context.Context, email string) (*EmployeeRow, error) {
	row := &EmployeeRow{}
	var jobTitle sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, first_name, last_name, email, profile_id, is_active, job_title, org_node_id
		 FROM employees WHERE email = $1`, email,
	).Scan(&row.ID, &row.FirstName, &row.LastName, &row.Email, &row.ProfileID, &row.IsActive, &jobTitle, &row.OrgNodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrEmployeeNotFound
		}
		return nil, err
	}
	row.JobTitle = jobTitle.String
	return row, nil
}

// AuthService provides authentication operations.
type AuthService struct {
	sessionStore *auth.SessionStore
	employeeRepo EmployeeReader
	db           *sql.DB // for direct profile name lookups
}

// NewAuthService creates a new AuthService.
func NewAuthService(sessionStore *auth.SessionStore, employeeRepo EmployeeReader, db *sql.DB) *AuthService {
	return &AuthService{
		sessionStore: sessionStore,
		employeeRepo: employeeRepo,
		db:           db,
	}
}

// LoginResult holds the response data after a successful login.
type LoginResult struct {
	Session *auth.Session
	Token   string
	Role    auth.Role
	Profile ProfileInfo
}

// ProfileInfo holds evaluation profile details.
type ProfileInfo struct {
	ID   uuid.UUID
	Name string
}

// Login authenticates an employee by email in dev mode (no password).
// TODO(auth:prod): Replace with password/SSO authentication.
func (s *AuthService) Login(ctx context.Context, email, ip, ua string) (*LoginResult, error) {
	emp, err := s.employeeRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if !emp.IsActive {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"employee account is inactive", nil)
	}

	// Look up the evaluation profile name to determine the role
	profile, err := s.getProfileName(ctx, emp.ProfileID)
	if err != nil {
		return nil, err
	}

	role := auth.ProfileNameToRole(profile.Name)

	session, token, err := s.sessionStore.Create(ctx, emp.ID, ip, ua)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Session: session,
		Token:   token,
		Role:    role,
		Profile: *profile,
	}, nil
}

// ValidateSessionResult holds the enriched session data after validation.
type ValidateSessionResult struct {
	Session   *auth.Session
	Role      auth.Role
	ProfileID uuid.UUID
}

// ValidateSession validates a raw token and returns the session with
// enriched role and profile information.
func (s *AuthService) ValidateSession(ctx context.Context, token string) (*ValidateSessionResult, error) {
	session, err := s.sessionStore.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid or expired session", nil)
	}

	emp, err := s.employeeRepo.GetByID(ctx, session.EmployeeID)
	if err != nil {
		return nil, err
	}

	profile, err := s.getProfileName(ctx, emp.ProfileID)
	if err != nil {
		return nil, err
	}

	role := auth.ProfileNameToRole(profile.Name)

	return &ValidateSessionResult{
		Session:   session,
		Role:      role,
		ProfileID: emp.ProfileID,
	}, nil
}

// Logout revokes a specific session.
func (s *AuthService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessionStore.Revoke(ctx, sessionID)
}

// Refresh extends the expiry time of a session.
func (s *AuthService) Refresh(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessionStore.Refresh(ctx, sessionID)
}

// Employee retrieves an employee by ID.
func (s *AuthService) Employee(ctx context.Context, id uuid.UUID) (*EmployeeRow, error) {
	return s.employeeRepo.GetByID(ctx, id)
}

// EmployeeByEmail retrieves an employee by email.
func (s *AuthService) EmployeeByEmail(ctx context.Context, email string) (*EmployeeRow, error) {
	return s.employeeRepo.GetByEmail(ctx, email)
}

// EmployeeRoleAndProfile returns the role and profile for an employee based
// on their evaluation profile. Used by the dev-login flow to populate the
// response without going through the full Login() path.
func (s *AuthService) EmployeeRoleAndProfile(ctx context.Context, empID uuid.UUID) (auth.Role, *ProfileInfo, error) {
	emp, err := s.employeeRepo.GetByID(ctx, empID)
	if err != nil {
		return "", nil, err
	}
	profile, err := s.getProfileName(ctx, emp.ProfileID)
	if err != nil {
		return "", nil, err
	}
	return auth.ProfileNameToRole(profile.Name), profile, nil
}

// SessionStore returns the underlying session store for direct access
// (e.g., dev login which bypasses the normal employee lookup).
func (s *AuthService) SessionStore() *auth.SessionStore {
	return s.sessionStore
}

// OrgNodeName returns the name of an org node by ID.
func (s *AuthService) OrgNodeName(ctx context.Context, orgNodeID uuid.UUID) (string, error) {
	var name string
	err := s.db.QueryRowContext(ctx,
		`SELECT name FROM org_nodes WHERE id = $1`, orgNodeID,
	).Scan(&name)
	if err != nil {
		return "", nil // node not found → empty name, not an error
	}
	return name, nil
}

// OrgNodeInfo returns the name and organization_id of an org node by ID.
func (s *AuthService) OrgNodeInfo(ctx context.Context, orgNodeID uuid.UUID) (name string, orgID uuid.UUID, err error) {
	err = s.db.QueryRowContext(ctx,
		`SELECT name, organization_id FROM org_nodes WHERE id = $1`, orgNodeID,
	).Scan(&name, &orgID)
	if err != nil {
		return "", uuid.Nil, nil // node not found → empty, not an error
	}
	return name, orgID, nil
}

// getProfileName retrieves the evaluation profile name for a given profile ID.
func (s *AuthService) getProfileName(ctx context.Context, profileID uuid.UUID) (*ProfileInfo, error) {
	p := &ProfileInfo{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name FROM evaluation_profiles WHERE id = $1`, profileID,
	).Scan(&p.ID, &p.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			// Fallback: use the profile ID as the name with a default role
			p.ID = profileID
			p.Name = "colaborador"
			return p, nil
		}
		return nil, err
	}
	p.Name = strings.ToLower(p.Name)
	return p, nil
}
