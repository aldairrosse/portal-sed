// Package commentchange provides lightweight HTTP handlers for goal comments
// and change requests. Uses raw SQL to avoid Ent codegen overhead.
package commentchange

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"
	authsvc "github.com/sed-evaluacion-desempeno/api/internal/service/auth"
	notifypkg "github.com/sed-evaluacion-desempeno/api/internal/service/notify"
)

// appBaseURL is the public base URL used to build goal links in notification emails.
// Override in tests; main.go can also set it from env on startup.
var appBaseURL = func() string {
	if v := os.Getenv("APP_BASE_URL"); v != "" {
		return v
	}
	return "http://localhost:5173"
}()

// --- DTOs ---

type GoalComment struct {
	ID           string  `json:"id"`
	GoalID       *string `json:"goal_id,omitempty"`
	CategoryID   *string `json:"category_id,omitempty"`
	AssignmentID *string `json:"assignment_id,omitempty"`
	AuthorID     string  `json:"author_id"`
	AuthorName   string  `json:"author_name"`
	Content      string  `json:"content"`
	CreatedAt    string  `json:"created_at"`
}

type CreateCommentRequest struct {
	Content    string `json:"content"`
	AuthorID   string `json:"author_id"`
	AuthorName string `json:"author_name"`
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
	Status     string `json:"status"`
	ApprovedBy string `json:"approved_by"`
}

// --- Handler ---

type Handler struct {
	db       *sql.DB
	notifier notifypkg.Sender
}

func NewHandler(db *sql.DB, notifier notifypkg.Sender) *Handler {
	if notifier == nil {
		notifier = notifypkg.NoopSender{}
	}
	return &Handler{db: db, notifier: notifier}
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
		SELECT id, goal_id, category_id, assignment_id, author_id, author_name, content, created_at
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
		if err := rows.Scan(&c.ID, &c.GoalID, &c.CategoryID, &c.AssignmentID, &c.AuthorID, &c.AuthorName, &c.Content, &c.CreatedAt); err != nil {
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

	authorID, ok := auth.GetEmployeeID(r.Context())
	if !ok {
		writeError(w, 401, "unauthenticated")
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

	// Derive author name from the DB; never trust the body's author_name.
	var authorName string
	if err := h.db.QueryRow(
		`SELECT first_name || ' ' || last_name FROM employees WHERE id = $1`,
		authorID,
	).Scan(&authorName); err != nil {
		log.Printf("commentchange: author lookup failed for %s: %v", authorID, err)
		writeError(w, 500, "failed to resolve author")
		return
	}

	id := uuid.New().String()
	var comment GoalComment
	err := h.db.QueryRow(`
		INSERT INTO goal_comments (id, goal_id, author_id, author_name, content)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, goal_id, category_id, assignment_id, author_id, author_name, content, created_at
	`, id, goalID, authorID, authorName, req.Content,
	).Scan(&comment.ID, &comment.GoalID, &comment.CategoryID, &comment.AssignmentID, &comment.AuthorID, &comment.AuthorName, &comment.Content, &comment.CreatedAt)
	if err != nil {
		writeError(w, 500, "failed to create comment")
		return
	}

	// Best-effort notification. Use a detached ctx for the lookups so a
	// client disconnect after the INSERT cannot blank out the email body.
	// ponytail: org scope is implicit via the auth middleware; explicit
	// employees.org_node_id filter would be needed if the middleware stops
	// scoping the request.
	titleCtx, titleCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer titleCancel()
	h.notifyOwner(titleCtx, notifypkg.Notification{
		Subject:  "Nuevo comentario en tu meta",
		Template: notifypkg.TemplateCommentCreated,
		Data: map[string]string{
			"AuthorName": authorName,
			"GoalTitle":  commentGoalTitle(titleCtx, h.db, goalID),
			"GoalURL":    appBaseURL + "/objetivos/asignacion",
		},
	}, goalID, "", authorID.String())

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
		SELECT id, goal_id, category_id, assignment_id, author_id, author_name, content, created_at
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
		if err := rows.Scan(&c.ID, &c.GoalID, &c.CategoryID, &c.AssignmentID, &c.AuthorID, &c.AuthorName, &c.Content, &c.CreatedAt); err != nil {
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

	authorID, ok := auth.GetEmployeeID(r.Context())
	if !ok {
		writeError(w, 401, "unauthenticated")
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

	var authorName string
	if err := h.db.QueryRow(
		`SELECT first_name || ' ' || last_name FROM employees WHERE id = $1`,
		authorID,
	).Scan(&authorName); err != nil {
		log.Printf("commentchange: author lookup failed for %s: %v", authorID, err)
		writeError(w, 500, "failed to resolve author")
		return
	}

	id := uuid.New().String()
	var comment GoalComment
	err := h.db.QueryRow(`
		INSERT INTO goal_comments (id, category_id, author_id, author_name, content)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, goal_id, category_id, assignment_id, author_id, author_name, content, created_at
	`, id, catID, authorID, authorName, req.Content,
	).Scan(&comment.ID, &comment.GoalID, &comment.CategoryID, &comment.AssignmentID, &comment.AuthorID, &comment.AuthorName, &comment.Content, &comment.CreatedAt)
	if err != nil {
		writeError(w, 500, "failed to create category comment")
		return
	}

	var catName string
	catCtx, catCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer catCancel()
	_ = h.db.QueryRowContext(catCtx, `SELECT name FROM goal_categories WHERE id = $1`, catID).Scan(&catName)

	h.notifyOwner(r.Context(), notifypkg.Notification{
		Subject:  "Nuevo comentario en tu categoría",
		Template: notifypkg.TemplateCommentCreated,
		Data: map[string]string{
			"AuthorName": authorName,
			"GoalTitle":  catName, // reuse the same template var; "categoría" reads fine
			"GoalURL":    appBaseURL + "/objetivos/asignacion",
		},
	}, "", catID, authorID.String())

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

// commentGoalTitle returns the goal's name for use in notification templates.
// Returns "" on error; the template should still render gracefully.
func commentGoalTitle(ctx context.Context, db *sql.DB, goalID string) string {
	var title string
	if err := db.QueryRowContext(ctx, `SELECT name FROM goals WHERE id = $1`, goalID).Scan(&title); err != nil {
		log.Printf("commentchange: goal title lookup failed for %s: %v", goalID, err)
		return ""
	}
	return title
}

// notifyOwner looks up the goal/category owner and fires a notification.
// Pass categoryID != "" for category comments (which use a different lookup path).
// authorID is the comment's author; we skip the send if author == owner (no self-email).
// ponytail: in-process goroutine, lost on restart; upgrade to a worker + notifications
// table when SLA demands delivery guarantees.
func (h *Handler) notifyOwner(parent context.Context, n notifypkg.Notification, goalID, categoryID, authorID string) {
	type lookup struct {
		ownerID string
		email   string
	}
	var l lookup
	var err error
	if categoryID != "" {
		err = h.db.QueryRow(`
			SELECT gc.employee_id, e.email
			FROM goal_categories gc
			JOIN employees e ON e.id = gc.employee_id
			WHERE gc.id = $1
		`, categoryID).Scan(&l.ownerID, &l.email)
	} else {
		err = h.db.QueryRow(`
			SELECT gc.employee_id, e.email
			FROM goals g
			JOIN goal_categories gc ON gc.id = g.category_id
			JOIN employees e ON e.id = gc.employee_id
			WHERE g.id = $1
		`, goalID).Scan(&l.ownerID, &l.email)
	}
	if err != nil {
		// Goal/category not found, no owner, or owner has no email — log and move on.
		log.Printf("commentchange: owner lookup skipped (goalID=%s categoryID=%s): %v", goalID, categoryID, err)
		return
	}
	if l.ownerID == authorID || l.email == "" {
		return
	}
	n.To = l.email
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.notifier.Send(ctx, n); err != nil {
			log.Printf("commentchange: notify failed (goalID=%s to=%s): %v", goalID, l.email, err)
		}
	}()
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

// --- Assignment Comments ---

func (h *Handler) ListAssignmentComments(w http.ResponseWriter, r *http.Request) {
	assignID := chi.URLParam(r, "assignId")
	if _, err := uuid.Parse(assignID); err != nil {
		writeError(w, 400, "invalid assignment ID")
		return
	}

	rows, err := h.db.Query(`
		SELECT id, goal_id, category_id, assignment_id, author_id, author_name, content, created_at
		FROM goal_comments WHERE assignment_id = $1 ORDER BY created_at ASC
	`, assignID)
	if err != nil {
		writeError(w, 500, "failed to list assignment comments")
		return
	}
	defer rows.Close()

	var comments []GoalComment
	for rows.Next() {
		var c GoalComment
		if err := rows.Scan(&c.ID, &c.GoalID, &c.CategoryID, &c.AssignmentID, &c.AuthorID, &c.AuthorName, &c.Content, &c.CreatedAt); err != nil {
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

func (h *Handler) CreateAssignmentComment(w http.ResponseWriter, r *http.Request) {
	assignID := chi.URLParam(r, "assignId")
	if _, err := uuid.Parse(assignID); err != nil {
		writeError(w, 400, "invalid assignment ID")
		return
	}

	authorID, ok := auth.GetEmployeeID(r.Context())
	if !ok {
		writeError(w, 401, "unauthenticated")
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

	var authorName string
	if err := h.db.QueryRow(
		`SELECT first_name || ' ' || last_name FROM employees WHERE id = $1`,
		authorID,
	).Scan(&authorName); err != nil {
		log.Printf("commentchange: author lookup failed for %s: %v", authorID, err)
		writeError(w, 500, "failed to resolve author")
		return
	}

	id := uuid.New().String()
	var comment GoalComment
	err := h.db.QueryRow(`
		INSERT INTO goal_comments (id, assignment_id, author_id, author_name, content)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, goal_id, category_id, assignment_id, author_id, author_name, content, created_at
	`, id, assignID, authorID, authorName, req.Content,
	).Scan(&comment.ID, &comment.GoalID, &comment.CategoryID, &comment.AssignmentID, &comment.AuthorID, &comment.AuthorName, &comment.Content, &comment.CreatedAt)
	if err != nil {
		writeError(w, 500, "failed to create assignment comment")
		return
	}

	writeJSON(w, 201, comment)
}

func (h *Handler) DeleteAssignmentComment(w http.ResponseWriter, r *http.Request) {
	assignID := chi.URLParam(r, "assignId")
	commentID := chi.URLParam(r, "commentId")
	if _, err := uuid.Parse(assignID); err != nil {
		writeError(w, 400, "invalid assignment ID")
		return
	}
	if _, err := uuid.Parse(commentID); err != nil {
		writeError(w, 400, "invalid comment ID")
		return
	}

	result, err := h.db.Exec(`DELETE FROM goal_comments WHERE id = $1 AND assignment_id = $2`, commentID, assignID)
	if err != nil {
		writeError(w, 500, "failed to delete assignment comment")
		return
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		writeError(w, 404, "comment not found")
		return
	}
	w.WriteHeader(204)
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
		r.Use(middleware.RequireLoA2())

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

		// Assignment comments
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalRead))
			r.Use(middleware.RateLimit(readRateLimit))
			r.Get("/assignments/{assignId}/comments", handler.ListAssignmentComments)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalRead))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Post("/assignments/{assignId}/comments", handler.CreateAssignmentComment)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalRead))
			r.Use(middleware.RateLimit(writeRateLimit))
			r.Delete("/assignments/{assignId}/comments/{commentId}", handler.DeleteAssignmentComment)
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
