package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net"
	"time"

	"github.com/google/uuid"
)

// Session represents an authenticated user session stored in the database.
type Session struct {
	ID           uuid.UUID
	EmployeeID   uuid.UUID
	TokenHash    string
	IPAddress    *string
	UserAgent    *string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	LastActiveAt time.Time
	IsRevoked    bool
	IDToken      string
	AccessToken  string
	RefreshToken string
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
func (s *SessionStore) Create(ctx context.Context, employeeID uuid.UUID, ip, ua, idToken, accessToken, refreshToken string) (*Session, string, error) {
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
		ID:           uuid.New(),
		EmployeeID:   employeeID,
		TokenHash:    tokenHash,
		IPAddress:    ipPtr,
		UserAgent:    uaPtr,
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
		LastActiveAt: now,
		IsRevoked:    false,
		IDToken:      idToken,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO sessions (id, employee_id, token_hash, ip_address, user_agent, expires_at, created_at, last_active_at, is_revoked, id_token, access_token, refresh_token)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		session.ID, session.EmployeeID, session.TokenHash, ipPtr, uaPtr,
		session.ExpiresAt, session.CreatedAt, session.LastActiveAt, session.IsRevoked,
		nullString(idToken), nullString(accessToken), nullString(refreshToken),
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

	err := s.db.QueryRowContext(ctx,
		`SELECT id, employee_id, token_hash, ip_address, user_agent,
		        expires_at, created_at, last_active_at, is_revoked,
		        id_token, access_token, refresh_token
		 FROM sessions WHERE token_hash = $1`, tokenHash,
	).Scan(
		&session.ID, &session.EmployeeID, &session.TokenHash,
		&ipPtr, &uaPtr,
		&session.ExpiresAt, &session.CreatedAt, &session.LastActiveAt, &session.IsRevoked,
		&idTok, &accTok, &refTok,
	)
	if err != nil {
		if err == sql.ErrNoRows {
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
		        id_token, access_token, refresh_token
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

		if err := rows.Scan(
			&session.ID, &session.EmployeeID, &session.TokenHash,
			&ipPtr, &uaPtr,
			&session.ExpiresAt, &session.CreatedAt, &session.LastActiveAt, &session.IsRevoked,
			&idTok, &accTok, &refTok,
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

// nullString returns a *string suitable for sql.NullString-like scanning,
// or nil if the value is empty.
func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
