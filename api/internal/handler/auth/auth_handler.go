// Package auth provides HTTP handlers for authentication endpoints.
package auth

import (
	"encoding/json"
	"log"
	"net/http"
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
	Session  SessionInfo  `json:"session"`
	Token    string       `json:"token"`
	Employee EmployeeInfo `json:"employee"`
	Role     string       `json:"role"`
}

// SessionInfo contains session metadata returned to the client.
type SessionInfo struct {
	ID        string `json:"id"`
	ExpiresAt string `json:"expires_at"`
}

// EmployeeInfo contains basic employee information.
type EmployeeInfo struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
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
		Role: string(result.Role),
	}

	writeJSON(w, http.StatusOK, resp)
}

// Logout handles POST /auth/logout.
// Revokes the current session identified by the Authorization header.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.GetSession(r.Context())
	if !ok || session == nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"no authenticated session", nil))
		return
	}

	if err := h.svc.Logout(r.Context(), session.ID); err != nil {
		writeError(w, err)
		return
	}

	// Clear the session cookie
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
	Employee EmployeeInfo `json:"employee"`
	Role     string      `json:"role"`
	Profile  ProfileInfo `json:"profile"`
}

// ProfileInfo holds evaluation profile information for the /me endpoint.
type ProfileInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Me handles GET /auth/me.
// Returns the current authenticated user's information.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	empID, ok := auth.GetEmployeeID(r.Context())
	if !ok {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"no authenticated session", nil))
		return
	}

	role, _ := auth.GetRole(r.Context())
	profileID, _ := auth.GetProfileID(r.Context())

	emp, err := h.svc.Employee(r.Context(), empID)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := MeResponse{
		Employee: EmployeeInfo{
			ID:        emp.ID.String(),
			FirstName: emp.FirstName,
			LastName:  emp.LastName,
			Email:     emp.Email,
		},
		Role: string(role),
		Profile: ProfileInfo{
			ID:   profileID.String(),
			Name: string(role),
		},
	}

	writeJSON(w, http.StatusOK, resp)
}
