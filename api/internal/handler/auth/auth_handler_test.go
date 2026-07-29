package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	"github.com/sed-evaluacion-desempeno/api/internal/auth/sso"
	handler "github.com/sed-evaluacion-desempeno/api/internal/handler/auth"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	svc "github.com/sed-evaluacion-desempeno/api/internal/service/auth"
)

type mockSSO struct {
	authURLFunc       func(state, nonce, acr, codeVerifier string) string
	exchangeFunc      func(ctx context.Context, code, codeVerifier string) (string, string, string, error)
	validateTokenFunc func(ctx context.Context, rawIDToken, expectedNonce string) (*sso.SSOUser, error)
	hasAccessFunc     func(ctx context.Context, accessToken string) bool
	getUserFunc       func(ctx context.Context, accessToken string) (*sso.SSOUser, error)
}

func (m *mockSSO) AuthorizationURL(state, nonce, acr, codeVerifier string) string {
	if m.authURLFunc != nil {
		return m.authURLFunc(state, nonce, acr, codeVerifier)
	}
	return "https://sso.example.com/auth?" + state
}

func (m *mockSSO) ExchangeCode(ctx context.Context, code, codeVerifier string) (string, string, string, error) {
	if m.exchangeFunc != nil {
		return m.exchangeFunc(ctx, code, codeVerifier)
	}
	return "acc_token", "id_token", "ref_token", nil
}

func (m *mockSSO) ValidateToken(ctx context.Context, rawIDToken, expectedNonce string) (*sso.SSOUser, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(ctx, rawIDToken, expectedNonce)
	}
	return &sso.SSOUser{ExternalID: "EMP001"}, nil
}

func (m *mockSSO) HasAccess(ctx context.Context, accessToken string) bool {
	if m.hasAccessFunc != nil {
		return m.hasAccessFunc(ctx, accessToken)
	}
	return true
}

func (m *mockSSO) GetUserFromToken(ctx context.Context, accessToken string) (*sso.SSOUser, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(ctx, accessToken)
	}
	return &sso.SSOUser{ExternalID: "EMP001", ACR: "1", Requires2FA: false}, nil
}

func (m *mockSSO) ValidateAccessToken(ctx context.Context, accessToken string) error { return nil }
func (m *mockSSO) GetEndSessionURL(ctx context.Context, idTokenHint string) (string, error) {
	return "https://sso.example.com/logout", nil
}
func (m *mockSSO) RevokeToken(ctx context.Context, token string) error { return nil }

type mockEmployeeReader struct {
	rows map[string]*svc.EmployeeRow
}

func (r *mockEmployeeReader) GetByID(_ context.Context, id uuid.UUID) (*svc.EmployeeRow, error) {
	for _, row := range r.rows {
		if row.ID == id {
			return row, nil
		}
	}
	return nil, pkgerrors.ErrEmployeeNotFound
}

func (r *mockEmployeeReader) GetByEmail(_ context.Context, _ string) (*svc.EmployeeRow, error) {
	return nil, pkgerrors.ErrEmployeeNotFound
}

func (r *mockEmployeeReader) GetByEmployeeNumber(_ context.Context, num string) (*svc.EmployeeRow, error) {
	if row, ok := r.rows[num]; ok {
		return row, nil
	}
	return nil, pkgerrors.ErrEmployeeNotFound
}

func TestInvalidState(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sessionStore := auth.NewSessionStore(db)
	authSvc := svc.NewAuthService(sessionStore, &mockEmployeeReader{rows: map[string]*svc.EmployeeRow{}}, db)
	txStore := sso.NewTransactionStore(10 * time.Minute)
	h := handler.NewAuthHandlerWithStore(authSvc, &mockSSO{}, txStore)

	req := httptest.NewRequest("GET", "/sso-callback?code=c&state=nonexistent", nil)
	w := httptest.NewRecorder()
	h.SSOCallback(w, req)
	resp := w.Result()
	resp.Body.Close()
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.True(t, strings.Contains(resp.Header.Get("Location"), "estado_invalido"))
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestNonceMismatch(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sessionStore := auth.NewSessionStore(db)
	authSvc := svc.NewAuthService(sessionStore, &mockEmployeeReader{rows: map[string]*svc.EmployeeRow{}}, db)
	txStore := sso.NewTransactionStore(10 * time.Minute)

	mockSSO := &mockSSO{
		validateTokenFunc: func(_ context.Context, _, _ string) (*sso.SSOUser, error) {
			return nil, assert.AnError
		},
	}
	h := handler.NewAuthHandlerWithStore(authSvc, mockSSO, txStore)

	txStore.Store("n_bad", &sso.OIDCTransaction{
		State: "n_bad", Nonce: "real", CodeVerifier: "cv",
	})

	req := httptest.NewRequest("GET", "/sso-callback?code=c&state=n_bad", nil)
	w := httptest.NewRecorder()
	h.SSOCallback(w, req)
	resp := w.Result()
	resp.Body.Close()
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.True(t, strings.Contains(resp.Header.Get("Location"), "id_token_invalido"))
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestMultipleTabs(t *testing.T) {
	txStore := sso.NewTransactionStore(10 * time.Minute)
	txStore.Store("tab_a", &sso.OIDCTransaction{State: "tab_a", Nonce: "n_a", CodeVerifier: "cv_a"})
	txStore.Store("tab_b", &sso.OIDCTransaction{State: "tab_b", Nonce: "n_b", CodeVerifier: "cv_b"})

	gotA, okA := txStore.Consume("tab_a")
	assert.True(t, okA)
	assert.Equal(t, "n_a", gotA.Nonce)

	gotB, okB := txStore.Consume("tab_b")
	assert.True(t, okB)
	assert.Equal(t, "n_b", gotB.Nonce)

	_, okAgain := txStore.Consume("tab_a")
	assert.False(t, okAgain)
}

func TestExpiredTransaction(t *testing.T) {
	fastStore := sso.NewTransactionStore(50 * time.Millisecond)
	fastStore.Store("exp", &sso.OIDCTransaction{State: "exp"})
	time.Sleep(100 * time.Millisecond)
	_, ok := fastStore.Consume("exp")
	assert.False(t, ok)

	fastStore.Store("fresh", &sso.OIDCTransaction{State: "fresh"})
	_, ok = fastStore.Consume("fresh")
	assert.True(t, ok)
}

func TestSSOStepUp_GeneratesNonce(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	var capturedNonce, capturedState string
	mockSSO := &mockSSO{
		authURLFunc: func(state, nonce, acr, codeVerifier string) string {
			capturedState = state
			capturedNonce = nonce
			return "https://sso.example.com/step-up?" + state
		},
	}
	sessionStore := auth.NewSessionStore(db)
	authSvc := svc.NewAuthService(sessionStore, &mockEmployeeReader{rows: map[string]*svc.EmployeeRow{}}, db)
	h := handler.NewAuthHandlerWithStore(authSvc, mockSSO, sso.NewTransactionStore(10*time.Minute))

	req := httptest.NewRequest("GET", "/sso-step-up?return_to=/secure", nil)
	w := httptest.NewRecorder()
	h.SSOStepUp(w, req)
	resp := w.Result()
	resp.Body.Close()
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.NotEmpty(t, capturedNonce)
	assert.NotEmpty(t, capturedState)
	assert.NotEqual(t, capturedState, capturedNonce)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestSSOStepUp_InvalidReturnTo(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sessionStore := auth.NewSessionStore(db)
	authSvc := svc.NewAuthService(sessionStore, &mockEmployeeReader{rows: map[string]*svc.EmployeeRow{}}, db)
	h := handler.NewAuthHandlerWithStore(authSvc, &mockSSO{}, sso.NewTransactionStore(10*time.Minute))

	tests := []struct{ name, ret string }{
		{"empty", ""},
		{"external", "http://evil.com"},
		{"double_slash", "//evil.com"},
		{"backslash", "/path\\evil"},
		{"no_slash", "path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/sso-step-up?return_to="+tt.ret, nil)
			w := httptest.NewRecorder()
			h.SSOStepUp(w, req)
			assert.Equal(t, http.StatusFound, w.Code)
		})
	}
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestSSOLoginRedirect_GeneratesNonce(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	var capturedNonce, capturedState string
	mockSSO := &mockSSO{
		authURLFunc: func(state, nonce, acr, codeVerifier string) string {
			capturedState = state
			capturedNonce = nonce
			return "https://sso.example.com/auth?" + state
		},
	}
	sessionStore := auth.NewSessionStore(db)
	authSvc := svc.NewAuthService(sessionStore, &mockEmployeeReader{rows: map[string]*svc.EmployeeRow{}}, db)
	h := handler.NewAuthHandlerWithStore(authSvc, mockSSO, sso.NewTransactionStore(10*time.Minute))

	req := httptest.NewRequest("GET", "/sso-login", nil)
	w := httptest.NewRecorder()
	h.SSOLoginRedirect(w, req)
	resp := w.Result()
	resp.Body.Close()
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.NotEmpty(t, capturedNonce)
	assert.NotEmpty(t, capturedState)
	assert.NotEqual(t, capturedState, capturedNonce)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestLogout(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sessionStore := auth.NewSessionStore(db)
	authSvc := svc.NewAuthService(sessionStore, &mockEmployeeReader{rows: map[string]*svc.EmployeeRow{}}, db)
	h := handler.NewAuthHandlerWithStore(authSvc, &mockSSO{}, sso.NewTransactionStore(10*time.Minute))

	req := httptest.NewRequest("GET", "/logout", nil)
	w := httptest.NewRecorder()
	h.Logout(w, req)
	resp := w.Result()
	resp.Body.Close()
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestStepUpIncomplete(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	empID := uuid.New()
	proID := uuid.New()

	mockSSO := &mockSSO{
		getUserFunc: func(_ context.Context, _ string) (*sso.SSOUser, error) {
			return &sso.SSOUser{ExternalID: "EMP004", ACR: "1", Requires2FA: true}, nil
		},
	}
	sessionStore := auth.NewSessionStore(db)
	reader := &mockEmployeeReader{rows: map[string]*svc.EmployeeRow{
		"EMP004": {ID: empID, EmployeeNumber: "EMP004", IsActive: true, ProfileID: proID, OrgNodeID: uuid.New()},
	}}
	authSvc := svc.NewAuthService(sessionStore, reader, db)
	txStore := sso.NewTransactionStore(10 * time.Minute)
	h := handler.NewAuthHandlerWithStore(authSvc, mockSSO, txStore)

	txStore.Store("fail", &sso.OIDCTransaction{
		State: "fail", Nonce: "n_fail", CodeVerifier: "cv_fail",
		RequestedACR: "mobo-2fa",
	})

	req := httptest.NewRequest("GET", "/sso-callback?code=c_fail&state=fail", nil)
	w := httptest.NewRecorder()
	h.SSOCallback(w, req)
	resp := w.Result()
	resp.Body.Close()
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	loc := resp.Header.Get("Location")
	assert.True(t, strings.Contains(loc, "stepup"), "step-up error: "+loc)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestLoginProtectedRole(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	empID := uuid.New()
	proID := uuid.New()

	mockSSO := &mockSSO{
		getUserFunc: func(_ context.Context, _ string) (*sso.SSOUser, error) {
			return &sso.SSOUser{ExternalID: "EMP002", ACR: "1", Requires2FA: true}, nil
		},
	}
	sessionStore := auth.NewSessionStore(db)
	reader := &mockEmployeeReader{rows: map[string]*svc.EmployeeRow{
		"EMP002": {ID: empID, EmployeeNumber: "EMP002", IsActive: true, ProfileID: proID, OrgNodeID: uuid.New()},
	}}
	authSvc := svc.NewAuthService(sessionStore, reader, db)
	txStore := sso.NewTransactionStore(10 * time.Minute)
	h := handler.NewAuthHandlerWithStore(authSvc, mockSSO, txStore)

	txStore.Store("s_otp", &sso.OIDCTransaction{
		State: "s_otp", Nonce: "n_otp",
		CodeVerifier: "cv_otp", ReturnTo: "/dashboard",
	})

	req := httptest.NewRequest("GET", "/sso-callback?code=c_otp&state=s_otp", nil)
	w := httptest.NewRecorder()
	h.SSOCallback(w, req)
	resp := w.Result()
	resp.Body.Close()
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	loc := resp.Header.Get("Location")
	assert.True(t, strings.Contains(loc, "sso"), "redirect to SSO: "+loc)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}
