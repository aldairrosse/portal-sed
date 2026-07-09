// Package commentchange provides lightweight HTTP handlers for goal comments
// and change requests. Uses raw SQL to avoid Ent codegen overhead.
package commentchange

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"
	authsvc "github.com/sed-evaluacion-desempeno/api/internal/service/auth"
)

// --- DTOs ---

type GoalComment struct {
	ID         string  `json:"id"`
	GoalID     *string `json:"goal_id,omitempty"`
	CategoryID *string `json:"category_id,omitempty"`
	AuthorID   string  `json:"author_id"`
	AuthorName string  `json:"author_name"`
	Content    string  `json:"content"`
	CreatedAt  string  `json:"created_at"`
}

type CreateCommentRequest struct {
	Content    string  `json:"content"`
	AuthorID   string  `json:"author_id"`
	AuthorName string  `json:"author_name"`
}

type ChangeRequest struct {
	ID          string  `json:"id"`
	EntityType  string  `json:"entity_type"`
	EntityID    string  `json:"entity_id"`
	RequestedBy string  `json:"requested_by"`
	Status      string  `json:"status"`
	ApprovedBy  *string `json:"approved_by,omitempty"`
	ApprovedAt  *string `json:"approved_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type CreateChangeRequestRequest struct {
	EntityType  string `json:"entity_type"`
	EntityID    string `json:"entity_id"`
	RequestedBy string `json:"requested_by"`
}

type UpdateChangeRequestRequest struct {
	Status    string `json:"status"`
	ApprovedBy string `json:"approved_by"`
}

// --- Handler ---

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("commentchange: encode error: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]interface{}{
		"error": map[string]string{"message": msg},
	})
}

// --- Goal Comments ---

func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	goalID := chi.URLParam(r, "goalId")
	if _, err := uuid.Parse(goalID); err != nil {
		writeError(w, 400, "invalid goal ID")
		return
	}

	rows, err := h.db.Query(`
		SELECT id, goal_id, category_id, author_id, author_name, content, created_at
		FROM goal_comments WHERE goal_id = $1 ORDER BY created_at ASC
	`, goalID)
	if err != nil {
		writeError(w, 500, "failed to list comments")
		return
	}
	defer rows.Close()

	var comments []GoalComment
	for rows.Next() {
		var c GoalComment
		if err := rows.Scan(&c.ID, &c.GoalID, &c.CategoryID, &c.AuthorID, &c.AuthorName, &c.Content, &c.CreatedAt); err != nil {
			writeError(w, 500, "failed to scan comment")
			return
		}
		comments = append(comments, c)
	}
	if comments == nil {
		comments = []GoalComment{}
	}
	writeJSON(w, 200, comments)
}

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	goalID := chi.URLParam(r, "goalId")
	if _, err := uuid.Parse(goalID); err != nil {
		writeError(w, 400, "invalid goal ID")
		return
	}

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid request body")
		return
	}
	if req.Content == "" {
		writeError(w, 400, "content is required")
		return
	}

	id := uuid.New().String()
	var comment GoalComment
	err := h.db.QueryRow(`
		INSERT INTO goal_comments (id, goal_id, author_id, author_name, content)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, goal_id, category_id, author_id, author_name, content, created_at
	`, id, goalID, req.AuthorID, req.AuthorName, req.Content,
	).Scan(&comment.ID, &comment.GoalID, &comment.CategoryID, &comment.AuthorID, &comment.AuthorName, &comment.Content, &comment.CreatedAt)
	if err != nil {
		writeError(w, 500, "failed to create comment")
		return
	}
	writeJSON(w, 201, comment)
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	goalID := chi.URLParam(r, "goalId")
	commentID := chi.URLParam(r, "commentId")
	if _, err := uuid.Parse(goalID); err != nil {
		writeError(w, 400, "invalid goal ID")
		return
	}
	if _, err := uuid.Parse(commentID); err != nil {
		writeError(w, 400, "invalid comment ID")
		return
	}

	result, err := h.db.Exec(`DELETE FROM goal_comments WHERE id = $1 AND goal_id = $2`, commentID, goalID)
	if err != nil {
		writeError(w, 500, "failed to delete comment")
		return
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		writeError(w, 404, "comment not found")
		return
	}
	w.WriteHeader(204)
}

// --- Category Comments ---

func (h *Handler) ListCategoryComments(w http.ResponseWriter, r *http.Request) {
	catID := chi.URLParam(r, "catId")
	if _, err := uuid.Parse(catID); err != nil {
		writeError(w, 400, "invalid category ID")
		return
	}

	rows, err := h.db.Query(`
		SELECT id, goal_id, category_id, author_id, author_name, content, created_at
		FROM goal_comments WHERE category_id = $1 ORDER BY created_at ASC
	`, catID)
	if err != nil {
		writeError(w, 500, "failed to list category comments")
		return
	}
	defer rows.Close()

	var comments []GoalComment
	for rows.Next() {
		var c GoalComment
		if err := rows.Scan(&c.ID, &c.GoalID, &c.CategoryID, &c.AuthorID, &c.AuthorName, &c.Content, &c.CreatedAt); err != nil {
			writeError(w, 500, "failed to scan comment")
			return
		}
		comments = append(comments, c)
	}
	if comments == nil {
		comments = []GoalComment{}
	}
	writeJSON(w, 200, comments)
}

func (h *Handler) CreateCategoryComment(w http.ResponseWriter, r *http.Request) {
	catID := chi.URLParam(r, "catId")
	if _, err := uuid.Parse(catID); err != nil {
		writeError(w, 400, "invalid category ID")
		return
	}

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid request body")
		return
	}
	if req.Content == "" {
		writeError(w, 400, "content is required")
		return
	}

	id := uuid.New().String()
	var comment GoalComment
	err := h.db.QueryRow(`
		INSERT INTO goal_comments (id, category_id, author_id, author_name, content)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, goal_id, category_id, author_id, author_name, content, created_at
	`, id, catID, req.AuthorID, req.AuthorName, req.Content,
	).Scan(&comment.ID, &comment.GoalID, &comment.CategoryID, &comment.AuthorID, &comment.AuthorName, &comment.Content, &comment.CreatedAt)
	if err != nil {
		writeError(w, 500, "failed to create category comment")
		return
	}
	writeJSON(w, 201, comment)
}

func (h *Handler) DeleteCategoryComment(w http.ResponseWriter, r *http.Request) {
	catID := chi.URLParam(r, "catId")
	commentID := chi.URLParam(r, "commentId")
	if _, err := uuid.Parse(catID); err != nil {
		writeError(w, 400, "invalid category ID")
		return
	}
	if _, err := uuid.Parse(commentID); err != nil {
		writeError(w, 400, "invalid comment ID")
		return
	}

	result, err := h.db.Exec(`DELETE FROM goal_comments WHERE id = $1 AND category_id = $2`, commentID, catID)
	if err != nil {
		writeError(w, 500, "failed to delete category comment")
		return
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		writeError(w, 404, "comment not found")
		return
	}
	w.WriteHeader(204)
}

// --- Change Requests ---

func (h *Handler) ListChangeRequests(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")

	query := `SELECT id, entity_type, entity_id, requested_by, status, approved_by, approved_at, created_at FROM change_requests WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if entityType != "" {
		query += ` AND entity_type = $` + itoa(argIdx)
		args = append(args, entityType)
		argIdx++
	}
	if entityID != "" {
		query += ` AND entity_id = $` + itoa(argIdx)
		args = append(args, entityID)
		argIdx++
	}
	query += ` ORDER BY created_at DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		writeError(w, 500, "failed to list change requests")
		return
	}
	defer rows.Close()

	var requests []ChangeRequest
	for rows.Next() {
		var cr ChangeRequest
		if err := rows.Scan(&cr.ID, &cr.EntityType, &cr.EntityID, &cr.RequestedBy, &cr.Status, &cr.ApprovedBy, &cr.ApprovedAt, &cr.CreatedAt); err != nil {
			writeError(w, 500, "failed to scan change request")
			return
		}
		requests = append(requests, cr)
	}
	if requests == nil {
		requests = []ChangeRequest{}
	}
	writeJSON(w, 200, requests)
}

func (h *Handler) CreateChangeRequest(w http.ResponseWriter, r *http.Request) {
	var req CreateChangeRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid request body")
		return
	}
	if req.EntityType == "" || req.EntityID == "" || req.RequestedBy == "" {
		writeError(w, 400, "entity_type, entity_id, and requested_by are required")
		return
	}

	id := uuid.New().String()
	var cr ChangeRequest
	err := h.db.QueryRow(`
		INSERT INTO change_requests (id, entity_type, entity_id, requested_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, entity_type, entity_id, requested_by, status, approved_by, approved_at, created_at
	`, id, req.EntityType, req.EntityID, req.RequestedBy,
	).Scan(&cr.ID, &cr.EntityType, &cr.EntityID, &cr.RequestedBy, &cr.Status, &cr.ApprovedBy, &cr.ApprovedAt, &cr.CreatedAt)
	if err != nil {
		writeError(w, 500, "failed to create change request")
		return
	}
	writeJSON(w, 201, cr)
}

func (h *Handler) UpdateChangeRequest(w http.ResponseWriter, r *http.Request) {
	crID := chi.URLParam(r, "crId")
	if _, err := uuid.Parse(crID); err != nil {
		writeError(w, 400, "invalid change request ID")
		return
	}

	var req UpdateChangeRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid request body")
		return
	}
	if req.Status != "approved" && req.Status != "rejected" {
		writeError(w, 400, "status must be 'approved' or 'rejected'")
		return
	}

	now := time.Now().UTC()
	var cr ChangeRequest
	err := h.db.QueryRow(`
		UPDATE change_requests SET status = $1, approved_by = $2, approved_at = $3
		WHERE id = $4
		RETURNING id, entity_type, entity_id, requested_by, status, approved_by, approved_at, created_at
	`, req.Status, req.ApprovedBy, now, crID,
	).Scan(&cr.ID, &cr.EntityType, &cr.EntityID, &cr.RequestedBy, &cr.Status, &cr.ApprovedBy, &cr.ApprovedAt, &cr.CreatedAt)
	if err != nil {
		writeError(w, 500, "failed to update change request")
		return
	}
	writeJSON(w, 200, cr)
}

// itoa converts a small int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// RegisterRoutes registers comment and change-request endpoints on the given router.
// Expected mount: /api/v1
func RegisterRoutes(r chi.Router, handler *Handler, authSvc *authsvc.AuthService) {
	readRateLimit := middleware.RateLimitConfig{
		Window:   time.Minute,
		MaxCount: 1000,
		Store:    middleware.NewInMemoryRateLimitStore(),
	}
	writeRateLimit := middleware.RateLimitConfig{
		Window:   time.Minute,
		MaxCount: 100,
		Store:    middleware.NewInMemoryRateLimitStore(),
	}

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(authSvc))

		// Goal comments
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalRead))
			r.Use(middleware.RateLimit(readRateLimit))
			r.Get("/goals/{goalId}/comments", handler.ListComments)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalRead))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Post("/goals/{goalId}/comments", handler.CreateComment)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalRead))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Delete("/goals/{goalId}/comments/{commentId}", handler.DeleteComment)
		})

		// Category comments
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalRead))
			r.Use(middleware.RateLimit(readRateLimit))
			r.Get("/categories/{catId}/comments", handler.ListCategoryComments)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalRead))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Post("/categories/{catId}/comments", handler.CreateCategoryComment)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalRead))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Delete("/categories/{catId}/comments/{commentId}", handler.DeleteCategoryComment)
		})

		// Change requests
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalRead))
			r.Use(middleware.RateLimit(readRateLimit))
			r.Get("/change-requests", handler.ListChangeRequests)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalRead))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Post("/change-requests", handler.CreateChangeRequest)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalRead))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Patch("/change-requests/{crId}", handler.UpdateChangeRequest)
		})
	})
}
