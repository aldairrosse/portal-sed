package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
)

// Session represents an authenticated user session stored in the database.
type Session struct {
	ID                uuid.UUID
	EmployeeID        uuid.UUID
	TokenHash         string
	IPAddress         *string
	UserAgent         *string
	ExpiresAt         time.Time
	CreatedAt         time.Time
	LastActiveAt      time.Time
	IsRevoked         bool
	IDToken           string
	AccessToken       string
	RefreshToken      string
	ACR               string
	Requires2FA       bool
	TokenExpiresAt    time.Time
	RefreshExpiresAt  time.Time
}

// SessionStore provides database operations for session management.
type SessionStore struct {
	db *sql.DB
}

// NewSessionStore creates a new SessionStore.
func NewSessionStore(db *sql.DB) *SessionStore {
	return &SessionStore{db: db}
}

// Create generates a new session, stores it in the database, and returns
// the raw token (only shown once at creation time).
func (s *SessionStore) Create(ctx context.Context, employeeID uuid.UUID, ip, ua, idToken, accessToken, refreshToken, acr string, requires2FA bool, tokenExpiresAt, refreshExpiresAt time.Time) (*Session, string, error) {
	token, err := GenerateToken()
	if err != nil {
		return nil, "", err
	}

	tokenHash := HashToken(token)
	now := time.Now().UTC()
	expiresAt := now.Add(24 * time.Hour)

	var ipPtr, uaPtr *string
	if host := stripPort(ip); host != "" {
		ipPtr = &host
	}
	if ua != "" {
		uaPtr = &ua
	}

	session := &Session{
		ID:                uuid.New(),
		EmployeeID:        employeeID,
		TokenHash:         tokenHash,
		IPAddress:         ipPtr,
		UserAgent:         uaPtr,
		ExpiresAt:         expiresAt,
		CreatedAt:         now,
		LastActiveAt:      now,
		IsRevoked:         false,
		IDToken:           idToken,
		AccessToken:       accessToken,
		RefreshToken:      refreshToken,
		ACR:               acr,
		Requires2FA:       requires2FA,
		TokenExpiresAt:    tokenExpiresAt,
		RefreshExpiresAt:  refreshExpiresAt,
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO sessions (id, employee_id, token_hash, ip_address, user_agent, expires_at, created_at, last_active_at, is_revoked, id_token, access_token, refresh_token, acr, requires_2fa, token_expires_at, refresh_expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
		session.ID, session.EmployeeID, session.TokenHash, ipPtr, uaPtr,
		session.ExpiresAt, session.CreatedAt, session.LastActiveAt, session.IsRevoked,
		nullString(idToken), nullString(accessToken), nullString(refreshToken),
		nullString(acr), requires2FA,
		nullTime(tokenExpiresAt), nullTime(refreshExpiresAt),
	)
	if err != nil {
		return nil, "", err
	}

	return session, token, nil
}

// GetByToken retrieves a session by its raw token. Returns nil if the session
// is expired, revoked, or not found.
func (s *SessionStore) GetByToken(ctx context.Context, token string) (*Session, error) {
	tokenHash := HashToken(token)

	session := &Session{}
	var ipPtr, uaPtr sql.NullString
	var idTok, accTok, refTok sql.NullString
	var acrTok sql.NullString

	err := s.db.QueryRowContext(ctx,
		`SELECT id, employee_id, token_hash, ip_address, user_agent,
		        expires_at, created_at, last_active_at, is_revoked,
		        id_token, access_token, refresh_token, acr, requires_2fa,
		        token_expires_at, refresh_expires_at
		 FROM sessions WHERE token_hash = $1`, tokenHash,
	).Scan(
		&session.ID, &session.EmployeeID, &session.TokenHash,
		&ipPtr, &uaPtr,
		&session.ExpiresAt, &session.CreatedAt, &session.LastActiveAt, &session.IsRevoked,
		&idTok, &accTok, &refTok, &acrTok, &session.Requires2FA,
		&session.TokenExpiresAt, &session.RefreshExpiresAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("session store: GetByToken no_rows tokenHash(last8)=%s", tokenHash[len(tokenHash)-8:])
			return nil, nil
		}
		return nil, err
	}

	if ipPtr.Valid {
		session.IPAddress = &ipPtr.String
	}
	if uaPtr.Valid {
		session.UserAgent = &uaPtr.String
	}
	if idTok.Valid {
		session.IDToken = idTok.String
	}
	if accTok.Valid {
		session.AccessToken = accTok.String
	}
	if refTok.Valid {
		session.RefreshToken = refTok.String
	}
	if acrTok.Valid {
		session.ACR = acrTok.String
	}

	// Check expiry and revocation
	if session.IsRevoked || time.Now().UTC().After(session.ExpiresAt) {
		return nil, nil
	}

	return session, nil
}

// Refresh extends the session expiry and updates last_active_at.
func (s *SessionStore) Refresh(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now().UTC()
	expiresAt := now.Add(24 * time.Hour)

	result, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET expires_at = $1, last_active_at = $2 WHERE id = $3 AND NOT is_revoked`,
		expiresAt, now, sessionID,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// UpdateSecurityContext updates the OIDC tokens, ACR level, and 2FA flag
// for an existing session. Used after token refresh/revalidation when the
// security context (roles, ACR) may have changed server-side.
func (s *SessionStore) UpdateSecurityContext(
	ctx context.Context,
	sessionID uuid.UUID,
	idToken string,
	accessToken string,
	refreshToken string,
	acr string,
	requires2FA bool,
) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET id_token = $1, access_token = $2, refresh_token = $3, acr = $4, requires_2fa = $5
		 WHERE id = $6 AND NOT is_revoked`,
		nullString(idToken), nullString(accessToken), nullString(refreshToken),
		nullString(acr), requires2FA, sessionID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateTokens updates the access_token, refresh_token, and token_expires_at
// for an existing session. Used after a successful token refresh.
func (s *SessionStore) UpdateTokens(ctx context.Context, sessionID uuid.UUID, accessToken, refreshToken string, tokenExpiresAt, refreshExpiresAt time.Time) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET access_token = $1, refresh_token = $2, token_expires_at = $3, refresh_expires_at = $4
		 WHERE id = $5 AND NOT is_revoked`,
		nullString(accessToken), nullString(refreshToken), tokenExpiresAt, nullTime(refreshExpiresAt), sessionID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Revoke marks a session as revoked (logout).
func (s *SessionStore) Revoke(ctx context.Context, sessionID uuid.UUID) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET is_revoked = true WHERE id = $1`, sessionID,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ListByEmployeeID returns all active (non-expired, non-revoked) sessions
// for a given employee. Used by admin revoke to collect SSO tokens.
func (s *SessionStore) ListByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]*Session, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, employee_id, token_hash, ip_address, user_agent,
		        expires_at, created_at, last_active_at, is_revoked,
		        id_token, access_token, refresh_token, acr, requires_2fa,
		        token_expires_at, refresh_expires_at
		 FROM sessions
		 WHERE employee_id = $1 AND NOT is_revoked AND expires_at > NOW()`,
		employeeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		session := &Session{}
		var ipPtr, uaPtr sql.NullString
		var idTok, accTok, refTok sql.NullString
		var acrTok sql.NullString

		if err := rows.Scan(
			&session.ID, &session.EmployeeID, &session.TokenHash,
			&ipPtr, &uaPtr,
			&session.ExpiresAt, &session.CreatedAt, &session.LastActiveAt, &session.IsRevoked,
			&idTok, &accTok, &refTok, &acrTok, &session.Requires2FA,
			&session.TokenExpiresAt, &session.RefreshExpiresAt,
		); err != nil {
			return nil, err
		}

		if ipPtr.Valid {
			session.IPAddress = &ipPtr.String
		}
		if uaPtr.Valid {
			session.UserAgent = &uaPtr.String
		}
		if idTok.Valid {
			session.IDToken = idTok.String
		}
		if accTok.Valid {
			session.AccessToken = accTok.String
		}
		if refTok.Valid {
			session.RefreshToken = refTok.String
		}
		if acrTok.Valid {
			session.ACR = acrTok.String
		}

		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// RevokeAllEmployeeSessions revokes all active sessions for an employee.
func (s *SessionStore) RevokeAllEmployeeSessions(ctx context.Context, employeeID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET is_revoked = true WHERE employee_id = $1 AND NOT is_revoked`,
		employeeID,
	)
	return err
}

// HashToken computes the SHA-256 hex digest of a token.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// stripPort removes the port from a host:port address.
// r.RemoteAddr returns "ip:port" but the INET column stores the host only.
func stripPort(addr string) string {
	if addr == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return addr
}

// GenerateToken creates a cryptographically random 32-byte token encoded as hex.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// nullTime returns a *time.Time suitable for nullable timestamp columns,
// or nil if the value is zero.
func nullTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// nullString returns a *string suitable for sql.NullString-like scanning,
// or nil if the value is empty.
func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
