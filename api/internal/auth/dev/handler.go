package dev

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/sed-evaluacion-desempeno/api/internal/auth"
)

// Handler provides dev-only endpoints. All handlers return 404 when not in development.
type Handler struct {
	db           *sql.DB
	sessionStore *auth.SessionStore
}

func NewHandler(db *sql.DB, sessionStore *auth.SessionStore) *Handler {
	return &Handler{db: db, sessionStore: sessionStore}
}

func isDevEnabled() bool {
	return os.Getenv("ENV") == "development" || os.Getenv("APP_ENV") == "development"
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Status returns 200 when dev mode is enabled, 404 otherwise.
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	if !isDevEnabled() {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": true})
}

type employeeRow struct {
	ID          string `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	ProfileName string `json:"profile_name"`
	Email       string `json:"email"`
}

// ListEmployees returns active employees with profile info from real DB.
// Supports q/offset/limit for dev unauthenticated picker (scroll + search).
func (h *Handler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	if !isDevEnabled() {
		http.NotFound(w, r)
		return
	}
	q := r.URL.Query().Get("q")
	offset := 0
	limit := 50
	{
		var o int
		if _, err := parseIntParam(r.URL.Query().Get("offset"), &o); err == nil && o >= 0 {
			offset = o
		}
		var l int
		if _, err := parseIntParam(r.URL.Query().Get("limit"), &l); err == nil && l > 0 {
			limit = l
			if limit > 100 {
				limit = 100
			}
		}
	}
	rows, err := h.db.QueryContext(r.Context(), `
		SELECT e.id, COALESCE(e.first_name,''), COALESCE(e.last_name,''), COALESCE(e.email,''),
		       COALESCE(ep.name,'')
		FROM employees e
		LEFT JOIN evaluation_profiles ep ON e.profile_id = ep.id
		WHERE e.is_active = true
		  AND ($1 = '' OR e.first_name ILIKE '%' || $1 || '%' OR e.last_name ILIKE '%' || $1 || '%' OR e.email ILIKE '%' || $1 || '%')
		ORDER BY e.last_name, e.first_name
		LIMIT $2 OFFSET $3`, q, limit+1, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	var out []employeeRow
	for rows.Next() {
		var e employeeRow
		if err := rows.Scan(&e.ID, &e.FirstName, &e.LastName, &e.Email, &e.ProfileName); err != nil {
			continue
		}
		out = append(out, e)
	}
	if out == nil {
		out = []employeeRow{}
	}
	hasMore := len(out) > limit
	if hasMore {
		out = out[:limit]
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":      out,
		"employees": out,
		"meta":      map[string]bool{"hasMore": hasMore},
	})
}

func parseIntParam(s string, out *int) (int, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	*out = v
	return v, nil
}

type impersonateReq struct {
	EmployeeID string `json:"employee_id"`
}

// Impersonate creates a real session for the given employee and sets session_token cookie.
func (h *Handler) Impersonate(w http.ResponseWriter, r *http.Request) {
	if !isDevEnabled() {
		http.NotFound(w, r)
		return
	}
	var req impersonateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.EmployeeID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "employee_id required"})
		return
	}
	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid employee_id"})
		return
	}

	var profileName sql.NullString
	var isActive bool
	err = h.db.QueryRowContext(r.Context(),
		`SELECT COALESCE(ep.name,''), e.is_active FROM employees e LEFT JOIN evaluation_profiles ep ON e.profile_id = ep.id WHERE e.id = $1`, empID,
	).Scan(&profileName, &isActive)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "employee not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "lookup failed"})
		return
	}
	if !isActive {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "employee not active"})
		return
	}

	ip := r.RemoteAddr
	ua := r.UserAgent()
	session, token, err := h.sessionStore.Create(r.Context(), empID, ip, ua, "", "", "", "", false, time.Time{}, time.Time{})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "session create failed"})
		return
	}

	// Set cookie matching auth_handler behavior
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	})

	role := auth.ProfileNameToRole(profileName.String)
	writeJSON(w, http.StatusOK, map[string]string{
		"employee_id":  empID.String(),
		"profile_name": profileName.String,
		"role":         string(role),
	})
}
