// Package auth provides HTTP handlers for authentication endpoints.
package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	svc "github.com/sed-evaluacion-desempeno/api/internal/service/auth"
)

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("auth handler: failed to encode JSON response: %v", err)
	}
}

// writeError writes a structured error response.
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

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	svc *svc.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(svc *svc.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// LoginRequest is the JSON body for POST /auth/login.
type LoginRequest struct {
	Email string `json:"email"`
}

// LoginResponse is the JSON body returned after successful login.
type LoginResponse struct {
	Session        SessionInfo  `json:"session"`
	Token          string       `json:"token"`
	Employee       EmployeeInfo `json:"employee"`
	Role           string       `json:"role"`
	Profile        ProfileInfo  `json:"profile"`
	OrganizationID string       `json:"organization_id"`
}

// SessionInfo contains session metadata returned to the client.
type SessionInfo struct {
	ID        string `json:"id"`
	ExpiresAt string `json:"expires_at"`
}

// EmployeeInfo contains basic employee information.
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

// Login handles POST /auth/login.
// Dev mode: accepts email only, no password required.
// TODO(auth:prod): Add password validation and SSO support.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid JSON body", err))
		return
	}

	if req.Email == "" {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"email is required", nil))
		return
	}

	ip := r.RemoteAddr
	ua := r.UserAgent()

	result, err := h.svc.Login(r.Context(), req.Email, ip, ua)
	if err != nil {
		writeError(w, err)
		return
	}

	// Set httpOnly cookie with the session token
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    result.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		Expires:  result.Session.ExpiresAt,
	})

	// Resolve organization_id from the employee's org node
	var orgIDStr string
	if emp, err := h.svc.Employee(r.Context(), result.Session.EmployeeID); err == nil {
		if _, orgID, err := h.svc.OrgNodeInfo(r.Context(), emp.OrgNodeID); err == nil {
			orgIDStr = orgID.String()
		}
	}

	resp := LoginResponse{
		Session: SessionInfo{
			ID:        result.Session.ID.String(),
			ExpiresAt: result.Session.ExpiresAt.Format(time.RFC3339),
		},
		Token: result.Token,
		Employee: EmployeeInfo{
			ID: result.Session.EmployeeID.String(),
			// First and last name not available from session alone;
			// client can GET /auth/me for full details.
		},
		Role:           string(result.Role),
		OrganizationID: orgIDStr,
	}

	writeJSON(w, http.StatusOK, resp)
}

// DevLogin handles POST /auth/dev-login.
// Only available in ENV=development. Looks up an employee by email
// and creates a session without password/SSO — just needs the email
// to exist in the employees table.
func (h *AuthHandler) DevLogin(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("ENV") != "development" {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"dev login not available in production", nil))
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid JSON body", err))
		return
	}

	if req.Email == "" {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"email is required", nil))
		return
	}

	// Look up employee by email in the DB
	emp, err := h.svc.EmployeeByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"employee not found: "+req.Email, err))
		return
	}

	ip := r.RemoteAddr
	ua := r.UserAgent()

	// Create a real session — same as Login but skips password/SSO
	session, token, err := h.svc.SessionStore().Create(r.Context(), emp.ID, ip, ua)
	if err != nil {
		writeError(w, err)
		return
	}

	// Look up the role and profile from the employee's evaluation profile
	role, profile, _ := h.svc.EmployeeRoleAndProfile(r.Context(), emp.ID)

	orgNodeName, orgID, _ := h.svc.OrgNodeInfo(r.Context(), emp.OrgNodeID)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		Expires:  session.ExpiresAt,
	})

	resp := LoginResponse{
		Session: SessionInfo{
			ID:        session.ID.String(),
			ExpiresAt: session.ExpiresAt.Format(time.RFC3339),
		},
		Token: token,
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
		Role: string(role),
		Profile: ProfileInfo{
			ID:   profile.ID.String(),
			Name: string(role),
		},
		OrganizationID: orgID.String(),
	}

	writeJSON(w, http.StatusOK, resp)
}

// Logout handles POST /auth/logout.
// Revokes the current session identified by the session_token cookie
// or Authorization: Bearer header. Self-contained for the same reason
// as /me — clients need to be able to log out without first going
// through RequireAuth (which would 401 a stale or expired session).
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := extractSessionToken(r)

	// If we have a valid session, revoke it server-side
	if token != "" {
		if result, err := h.svc.ValidateSession(r.Context(), token); err == nil && result != nil && result.Session != nil {
			_ = h.svc.Logout(r.Context(), result.Session.ID)
		}
	}

	// Always clear the session cookie, even if no session was found —
	// the client should not see a stale cookie after a logout attempt.
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "session revoked"})
}

// Refresh handles POST /auth/refresh.
// Extends the expiry time of the current session.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.GetSession(r.Context())
	if !ok || session == nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"no authenticated session", nil))
		return
	}

	if err := h.svc.Refresh(r.Context(), session.ID); err != nil {
		writeError(w, err)
		return
	}

	newExpiry := time.Now().UTC().Add(24 * time.Hour)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "session refreshed",
		"expires_at": newExpiry.Format(time.RFC3339),
	})
}

// MeResponse is the JSON body returned by GET /auth/me.
type MeResponse struct {
	Employee       EmployeeInfo `json:"employee"`
	Role           string      `json:"role"`
	Profile        ProfileInfo `json:"profile"`
	OrganizationID string      `json:"organization_id"`
}

// ProfileInfo holds evaluation profile information for the /me endpoint.
type ProfileInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Me handles GET /auth/me.
// Returns the current authenticated user's information.
// Self-contained: extracts the session token from the session_token cookie
// or Authorization header, then validates it directly. (The /me endpoint
// runs outside the RequireAuth middleware by design — clients need to be
// able to ask "who am I?" to recover from a stale session.)
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	token := extractSessionToken(r)
	if token == "" {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"no authenticated session", nil))
		return
	}

	result, err := h.svc.ValidateSession(r.Context(), token)
	if err != nil || result == nil || result.Session == nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid or expired session", err))
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

// extractSessionToken pulls the session token from the session_token cookie
// or the Authorization: Bearer header, in that order.
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
