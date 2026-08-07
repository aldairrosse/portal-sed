package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	svc "github.com/sed-evaluacion-desempeno/api/internal/service/auth"
)

// failingValidator always rejects the access token against the SSO provider.
type failingValidator struct{}

func (*failingValidator) ValidateAccessToken(context.Context, string) error {
	return errors.New("access token expired")
}

// unusedEmployeeReader satisfies EmployeeReader; employee lookup is not
// reached when SSO validation fails.
type unusedEmployeeReader struct{}

func (*unusedEmployeeReader) GetByID(context.Context, uuid.UUID) (*svc.EmployeeRow, error) {
	return nil, errors.New("unused")
}
func (*unusedEmployeeReader) GetByEmail(context.Context, string) (*svc.EmployeeRow, error) {
	return nil, errors.New("unused")
}
func (*unusedEmployeeReader) GetByEmployeeNumber(context.Context, string) (*svc.EmployeeRow, error) {
	return nil, errors.New("unused")
}

func newSessionRow(id, empID uuid.UUID) *sqlmock.Rows {
	now := time.Now()
	return sqlmock.NewRows([]string{
		"id", "employee_id", "token_hash", "ip_address", "user_agent",
		"expires_at", "created_at", "last_active_at", "is_revoked",
		"id_token", "access_token", "refresh_token", "acr", "requires_2fa",
		"token_expires_at", "refresh_expires_at",
	}).AddRow(id, empID, "hash", nil, nil, now.Add(time.Hour), now, now, false,
		nil, "acc-token", "ref-token", "", false, nil, nil)
}

// TestValidateSession_InvalidAccessTokenDoesNotRevoke proves the key invariant:
// when the SSO access token fails validation, ValidateSession returns
// ErrAccessTokenInvalid and does NOT revoke the local session (a revoked
// session would be invisible to the subsequent refresh attempt in RequireAuth).
func TestValidateSession_InvalidAccessTokenDoesNotRevoke(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sessionID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	authSvc := svc.NewAuthService(auth.NewSessionStore(db), &unusedEmployeeReader{}, db).
		WithSSOValidator(&failingValidator{})

	// Only the session lookup is expected. If the old revoke-on-invalid
	// behavior were back, an unexpected UPDATE ... WHERE id = $1 would hit
	// this mock and surface as an error from the store.
	mock.ExpectQuery(`FROM sessions WHERE token_hash = \$1`).
		WithArgs(auth.HashToken("raw-token")).
		WillReturnRows(newSessionRow(sessionID, uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")))

	_, err = authSvc.ValidateSession(context.Background(), "raw-token")
	require.Error(t, err)

	var invalid *svc.ErrAccessTokenInvalid
	require.ErrorAs(t, err, &invalid)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestErrAccessTokenInvalid_ErrorAndUnwrap(t *testing.T) {
	cause := errors.New("token expired")
	e := &svc.ErrAccessTokenInvalid{Err: cause}
	assert.Contains(t, e.Error(), "access token invalid")
	assert.ErrorIs(t, e, cause)
}
