package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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
func (s *SessionStore) Create(ctx context.Context, employeeID uuid.UUID, ip, ua string) (*Session, string, error) {
	token, err := GenerateToken()
	if err != nil {
		return nil, "", err
	}

	tokenHash := HashToken(token)
	now := time.Now().UTC()
	expiresAt := now.Add(24 * time.Hour)

	var ipPtr, uaPtr *string
	if ip != "" {
		ipPtr = &ip
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
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO sessions (id, employee_id, token_hash, ip_address, user_agent, expires_at, created_at, last_active_at, is_revoked)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		session.ID, session.EmployeeID, session.TokenHash, ipPtr, uaPtr,
		session.ExpiresAt, session.CreatedAt, session.LastActiveAt, session.IsRevoked,
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

	err := s.db.QueryRowContext(ctx,
		`SELECT id, employee_id, token_hash, ip_address, user_agent, expires_at, created_at, last_active_at, is_revoked
		 FROM sessions WHERE token_hash = $1`, tokenHash,
	).Scan(
		&session.ID, &session.EmployeeID, &session.TokenHash,
		&ipPtr, &uaPtr,
		&session.ExpiresAt, &session.CreatedAt, &session.LastActiveAt, &session.IsRevoked,
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

// GenerateToken creates a cryptographically random 32-byte token encoded as hex.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
