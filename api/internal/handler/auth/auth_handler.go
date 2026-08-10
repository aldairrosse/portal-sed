package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	"github.com/sed-evaluacion-desempeno/api/internal/auth/sso"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	svc "github.com/sed-evaluacion-desempeno/api/internal/service/auth"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("auth handler: failed to encode JSON response: %v", err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	traceID := uuid.New().String()[:8]

	var de *pkgerrors.DomainError
	if pkgerrors.AsDomainError(err, &de) {
		writeJSON(w, pkgerrors.HTTPStatus(err), pkgerrors.NewAPIErrorResponse(de, traceID))
		return
	}
	writeJSON(w, http.StatusInternalServerError, pkgerrors.NewAPIErrorResponse(
		pkgerrors.NewDomainError(pkgerrors.InvalidRequest, err.Error(), err),
		traceID,
	))
}

type AuthHandler struct {
	svc     *svc.AuthService
	sso     sso.SSOAdapter
	txStore *sso.TransactionStore
}

func NewAuthHandler(svc *svc.AuthService, ssoAdapter sso.SSOAdapter) *AuthHandler {
	return NewAuthHandlerWithStore(svc, ssoAdapter, sso.NewTransactionStore(10*time.Minute))
}

func NewAuthHandlerWithStore(svc *svc.AuthService, ssoAdapter sso.SSOAdapter, store *sso.TransactionStore) *AuthHandler {
	return &AuthHandler{svc: svc, sso: ssoAdapter, txStore: store}
}

func (h *AuthHandler) redirectError(w http.ResponseWriter, r *http.Request, code string) {
	frontendHost := os.Getenv("CORS_ORIGINS")
	if frontendHost == "" {
		frontendHost = "http://localhost:5173"
	}
	host := strings.Split(frontendHost, ",")[0]
	http.Redirect(w, r, host+"/login?sso_error="+code, http.StatusFound)
}

func (h *AuthHandler) writeStepUpError(w http.ResponseWriter, r *http.Request) {
	log.Printf("sso callback: step-up FAILED, redirecting sso_error=stepup_fallido")
	frontendHost := os.Getenv("CORS_ORIGINS")
	if frontendHost == "" {
		frontendHost = "http://localhost:5173"
	}
	host := strings.Split(frontendHost, ",")[0]
	http.Redirect(w, r, host+"/login?sso_error=stepup_fallido", http.StatusFound)
}

func (h *AuthHandler) SSOLoginRedirect(w http.ResponseWriter, r *http.Request) {
	state, err := auth.GenerateToken()
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "Error al iniciar SSO", err))
		return
	}
	nonce, err := auth.GenerateToken()
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "Error al generar nonce", err))
		return
	}

	codeVerifier, _, err := sso.GeneratePKCE()
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "Error al generar PKCE", err))
		return
	}

	returnTo := r.URL.Query().Get("return_to")
	if !validReturnTo(returnTo) {
		returnTo = "/"
	}

	h.txStore.Store(state, &sso.OIDCTransaction{
		State:        state,
		Nonce:        nonce,
		CodeVerifier: codeVerifier,
		ReturnTo:     returnTo,
	})

	authURL := h.sso.AuthorizationURL(state, nonce, "", codeVerifier)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *AuthHandler) SSOStepUp(w http.ResponseWriter, r *http.Request) {
	state, err := auth.GenerateToken()
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "Error al iniciar step-up", err))
		return
	}
	nonce, err := auth.GenerateToken()
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "Error al generar nonce", err))
		return
	}

	codeVerifier, _, err := sso.GeneratePKCE()
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "Error al generar PKCE", err))
		return
	}

	returnTo := r.URL.Query().Get("return_to")
	if !validReturnTo(returnTo) {
		returnTo = "/"
	}

	h.txStore.Store(state, &sso.OIDCTransaction{
		State:        state,
		Nonce:        nonce,
		CodeVerifier: codeVerifier,
		ReturnTo:     returnTo,
		RequestedACR: "mobo-2fa",
	})

	authURL := h.sso.AuthorizationURL(state, nonce, "mobo-2fa", codeVerifier)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func validReturnTo(value string) bool {
	if value == "" {
		return false
	}
	// Reject CR/LF and other control chars: decoded %0d/%0a would pass the
	// path checks below but http.Redirect writes them into the Location
	// header verbatim (header injection).
	if strings.ContainsFunc(value, func(r rune) bool {
		return r < 0x20 || r == 0x7f
	}) {
		return false
	}
	if !strings.HasPrefix(value, "/") {
		return false
	}
	if strings.HasPrefix(value, "//") {
		return false
	}
	if strings.Contains(value, "\\") {
		return false
	}
	return true
}

func (h *AuthHandler) SSOCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		h.redirectError(w, r, "parametros_invalidos")
		return
	}

	tx, ok := h.txStore.Consume(state)
	if !ok {
		h.redirectError(w, r, "estado_invalido")
		return
	}

	accessToken, rawIDToken, refreshToken, err := h.sso.ExchangeCode(r.Context(), code, tx.CodeVerifier)
	if err != nil {
		log.Printf("sso callback: exchange failed: %v", err)
		h.redirectError(w, r, "token_exchange_fallido")
		return
	}

	if _, err := h.sso.ValidateToken(r.Context(), rawIDToken, tx.Nonce); err != nil {
		log.Printf("sso callback: id_token invalid: %v", err)
		h.redirectError(w, r, "id_token_invalido")
		return
	}

	if !h.sso.HasAccess(r.Context(), accessToken) {
		log.Printf("sso callback: access denied")
		h.redirectError(w, r, "sin_acceso")
		return
	}

	ssoUser, err := h.sso.GetUserFromToken(r.Context(), accessToken)
	if err != nil {
		log.Printf("sso callback: get user failed: %v", err)
		h.redirectError(w, r, "error_usuario")
		return
	}

	requires2FA := ssoUser.Requires2FA
	hasLoA2 := ssoUser.ACR == "mobo-2fa"
	log.Printf("sso callback: step-up decide ext=%s acr=%q requires2FA=%v isStepUp=%v hasLoA2=%v", ssoUser.ExternalID, ssoUser.ACR, requires2FA, tx.IsStepUp(), hasLoA2)

	emp, err := h.svc.EmployeeByEmployeeNumber(r.Context(), ssoUser.ExternalID)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrEmployeeNotFound) {
			h.redirectError(w, r, "usuario_no_encontrado")
			return
		}
		log.Printf("sso callback: employee lookup failed: %v", err)
		h.redirectError(w, r, "error_bd")
		return
	}

	if !emp.IsActive {
		h.redirectError(w, r, "usuario_inactivo")
		return
	}

	if tx.IsStepUp() {
		if !hasLoA2 {
			log.Printf("sso callback: step-up demanded but acr is %q — rejecting", ssoUser.ACR)
			h.writeStepUpError(w, r)
			return
		}
	} else {
		if requires2FA && !hasLoA2 {
			log.Printf("sso callback: otp_required detected, initiating step-up for %s", ssoUser.ExternalID)
			h.initiateStepUp(w, r, tx.ReturnTo)
			return
		}
	}

	ip := r.RemoteAddr
	ua := r.UserAgent()
	session, token, err := h.svc.SessionStore().Create(r.Context(), emp.ID, ip, ua,
		rawIDToken, accessToken, refreshToken, ssoUser.ACR, requires2FA, time.Time{}, time.Time{})
	if err != nil {
		log.Printf("sso callback: session creation failed: %v", err)
		h.redirectError(w, r, "error_sesion")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name: "session_token", Value: token, Path: "/",
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode, Expires: session.ExpiresAt,
	})

	frontendHost := os.Getenv("CORS_ORIGINS")
	if frontendHost == "" {
		frontendHost = "http://localhost:5173"
	}
	host := strings.Split(frontendHost, ",")[0]

	redirectTo := tx.ReturnTo
	if !validReturnTo(redirectTo) {
		redirectTo = "/"
	}
	log.Printf("sso callback: session created id=%s acr=%q requires2FA=%v redirecting to %s", session.ID, ssoUser.ACR, requires2FA, host+redirectTo)
	http.Redirect(w, r, host+redirectTo, http.StatusFound)
}

func (h *AuthHandler) initiateStepUp(w http.ResponseWriter, r *http.Request, returnTo string) {
	state, err := auth.GenerateToken()
	if err != nil {
		log.Printf("sso step-up: failed to generate state: %v", err)
		h.redirectError(w, r, "stepup_error")
		return
	}
	nonce, err := auth.GenerateToken()
	if err != nil {
		log.Printf("sso step-up: failed to generate nonce: %v", err)
		h.redirectError(w, r, "stepup_error")
		return
	}

	codeVerifier, _, err := sso.GeneratePKCE()
	if err != nil {
		log.Printf("sso step-up: failed to generate PKCE: %v", err)
		h.redirectError(w, r, "stepup_error")
		return
	}

	h.txStore.Store(state, &sso.OIDCTransaction{
		State:        state,
		Nonce:        nonce,
		CodeVerifier: codeVerifier,
		ReturnTo:     returnTo,
		RequestedACR: "mobo-2fa",
	})

	authURL := h.sso.AuthorizationURL(state, nonce, "mobo-2fa", codeVerifier)
	http.Redirect(w, r, authURL, http.StatusFound)
}

type LoginResponse struct {
	Session        SessionInfo  `json:"session"`
	Token          string       `json:"token"`
	Employee       EmployeeInfo `json:"employee"`
	Role           string       `json:"role"`
	Profile        ProfileInfo  `json:"profile"`
	OrganizationID string       `json:"organization_id"`
}

type SessionInfo struct {
	ID        string `json:"id"`
	ExpiresAt string `json:"expires_at"`
}

type EmployeeInfo struct {
	ID             string `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Email          string `json:"email"`
	JobTitle       string `json:"job_title"`
	OrgNodeID      string `json:"org_node_id"`
	OrgNodeName    string `json:"org_node_name"`
	OrganizationID string `json:"organization_id"`
}

type ProfileInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := extractSessionToken(r)
	var idTokenHint string
	if token != "" {
		if sess, err := h.svc.SessionStore().GetByToken(r.Context(), token); err == nil && sess != nil {
			_ = h.svc.Logout(r.Context(), sess.ID)
			idTokenHint = sess.IDToken
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name: "session_token", Value: "", Path: "/",
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode, MaxAge: -1,
	})

	endSessionURL, err := h.sso.GetEndSessionURL(r.Context(), idTokenHint)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	http.Redirect(w, r, endSessionURL, http.StatusFound)
}

func (h *AuthHandler) LogoutComplete(w http.ResponseWriter, r *http.Request) {
	frontendHost := os.Getenv("CORS_ORIGINS")
	if frontendHost == "" {
		frontendHost = "http://localhost:5173"
	}
	host := strings.Split(frontendHost, ",")[0]
	http.Redirect(w, r, host+"/login", http.StatusFound)
}

func (h *AuthHandler) RevokeEmployeeSessions(w http.ResponseWriter, r *http.Request) {
	adminKey := os.Getenv("ADMIN_REVOKE_KEY")
	if adminKey == "" || r.Header.Get("X-Admin-Revoke-Key") != adminKey {
		writeError(w, pkgerrors.NewDomainError("FORBIDDEN", "Acceso no autorizado", nil))
		return
	}

	empIDStr := chi.URLParam(r, "empId")
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "ID de empleado inválido", err))
		return
	}

	sessions, err := h.svc.SessionStore().ListByEmployeeID(r.Context(), empID)
	if err != nil {
		log.Printf("admin revoke: failed to list sessions for %s: %v", empID, err)
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "Error al buscar sesiones", err))
		return
	}

	for _, sess := range sessions {
		if sess.RefreshToken != "" {
			if err := h.sso.RevokeToken(r.Context(), sess.RefreshToken); err != nil {
				log.Printf("admin revoke: SSO revoke failed for session %s: %v", sess.ID, err)
			}
		}
	}

	if err := h.svc.SessionStore().RevokeAllEmployeeSessions(r.Context(), empID); err != nil {
		log.Printf("admin revoke: local revoke failed for %s: %v", empID, err)
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "Error al revocar sesiones", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message":        "Sesiones revocadas",
		"employee_id":    empID.String(),
		"sessions_count": fmt.Sprintf("%d", len(sessions)),
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.GetSession(r.Context())
	if !ok || session == nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "Sesión no autenticada", nil))
		return
	}

	result, err := h.svc.RefreshTokens(r.Context(), session)
	if err != nil {
		if isFatalRefreshError(err) {
			// Destroy local session and cookie
			_ = h.svc.Logout(r.Context(), session.ID)
			http.SetCookie(w, &http.Cookie{
				Name: "session_token", Value: "", Path: "/",
				HttpOnly: true,
				Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
				SameSite: http.SameSiteLaxMode, MaxAge: -1,
			})
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "SSO_SESSION_EXPIRED",
			})
			return
		}
		// Transient error (network/timeout) → 502, session intact
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error": "Token refresh failed, try again",
		})
		return
	}

	// Persist new tokens (access + refresh expirations, atomically)
	tokenExpiresAt := time.Now().UTC().Add(time.Duration(result.ExpiresIn) * time.Second)
	refreshExpiresAt := time.Now().UTC().Add(time.Duration(result.RefreshExpiresIn) * time.Second)
	if err := h.svc.SessionStore().UpdateTokens(r.Context(), session.ID, result.AccessToken, result.RefreshToken, tokenExpiresAt, refreshExpiresAt); err != nil {
		log.Printf("auth handler: failed to persist refreshed tokens: %v", err)
		writeError(w, err)
		return
	}

	// Extend local session expiry
	if err := h.svc.SessionStore().Refresh(r.Context(), session.ID); err != nil {
		log.Printf("auth handler: failed to extend session expiry: %v", err)
	}

	newExpiry := time.Now().UTC().Add(24 * time.Hour)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":            "Sesión actualizada",
		"expires_at":         newExpiry.Format(time.RFC3339),
		"token_expires_at":   tokenExpiresAt.Format(time.RFC3339),
		"refresh_expires_at": refreshExpiresAt.Format(time.RFC3339),
		"refresh_expires_in": result.RefreshExpiresIn,
	})
}

// isFatalRefreshError returns true when the error indicates the refresh token
// is permanently invalid and the local session should be destroyed.
func isFatalRefreshError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, phrase := range []string{
		"invalid_grant",
		"session not active",
		"refresh token expired",
		"refresh token revoked",
	} {
		if strings.Contains(msg, phrase) {
			return true
		}
	}
	return false
}

type MeResponse struct {
	Employee       EmployeeInfo `json:"employee"`
	Role           string       `json:"role"`
	Profile        ProfileInfo  `json:"profile"`
	OrganizationID string       `json:"organization_id"`
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	token := extractSessionToken(r)
	if token == "" {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "Sesión no autenticada", nil))
		return
	}

	result, err := h.svc.ValidateSession(r.Context(), token)
	if err != nil || result == nil || result.Session == nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "Sesión inválida o expirada", err))
		return
	}

	emp, err := h.svc.Employee(r.Context(), result.Session.EmployeeID)
	if err != nil {
		writeError(w, err)
		return
	}

	orgNodeName, orgID, _ := h.svc.OrgNodeInfo(r.Context(), emp.OrgNodeID)

	resp := MeResponse{
		Employee: EmployeeInfo{
			ID:             emp.ID.String(),
			FirstName:      emp.FirstName,
			LastName:       emp.LastName,
			Email:          emp.Email,
			JobTitle:       emp.JobTitle,
			OrgNodeID:      emp.OrgNodeID.String(),
			OrgNodeName:    orgNodeName,
			OrganizationID: orgID.String(),
		},
		Role: string(result.Role),
		Profile: ProfileInfo{
			ID:   result.ProfileID.String(),
			Name: string(result.Role),
		},
		OrganizationID: orgID.String(),
	}

	writeJSON(w, http.StatusOK, resp)
}

func extractSessionToken(r *http.Request) string {
	if c, err := r.Cookie("session_token"); err == nil && c.Value != "" {
		return c.Value
	}
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}
