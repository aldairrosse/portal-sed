// Package cycle provides business logic for evaluation cycle management,
// including creation with advisory locking and phase transitions with
// optimistic concurrency control.
package cycle

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/cycle"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/cursor"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/cycle"
)

// CyclePhaseOrder defines the linear phase progression.
var CyclePhaseOrder = []cycle.CurrentPhase{
	cycle.CurrentPhaseAsignacion,
	cycle.CurrentPhaseAvance,
	cycle.CurrentPhaseCierre,
}

// phaseIndex returns the index of a phase in CyclePhaseOrder, or -1 if not found.
func phaseIndex(ph cycle.CurrentPhase) int {
	for i, p := range CyclePhaseOrder {
		if p == ph {
			return i
		}
	}
	return -1
}

// resolveNextPhase returns the next phase in the linear order.
// Returns empty string and false if the current phase is the last one.
func resolveNextPhase(current cycle.CurrentPhase) (cycle.CurrentPhase, bool) {
	idx := phaseIndex(current)
	if idx < 0 || idx >= len(CyclePhaseOrder)-1 {
		return "", false
	}
	return CyclePhaseOrder[idx+1], true
}

// CreateCycleRequest is the DTO for creating a new cycle.
type CreateCycleRequest struct {
	Year           int    `json:"year"`
	OrganizationID string `json:"organization_id"`
	IdempotencyKey string `json:"-"`
}

// TransitionPhaseRequest is the DTO for transitioning a cycle's phase.
type TransitionPhaseRequest struct {
	CycleID         string `json:"-"`
	ExpectedVersion int    `json:"-"`
	Trigger         string `json:"trigger"`
	Reason          string `json:"reason"`
	IdempotencyKey  string `json:"-"`
}

// ListCyclesRequest is the DTO for listing cycles.
type ListCyclesRequest struct {
	OrganizationID string `json:"organization_id"`
	Year           *int   `json:"year,omitempty"`
	CurrentPhase   *string `json:"current_phase,omitempty"`
	Cursor         string `json:"cursor,omitempty"`
	Limit          int    `json:"limit"`
}

// CycleResponse is the API response for a cycle.
type CycleResponse struct {
	ID             string  `json:"id"`
	Year           int     `json:"year"`
	OrganizationID string  `json:"organization_id"`
	CurrentPhase   string  `json:"current_phase"`
	Version        int     `json:"version"`
	StartedAt      *string `json:"started_at,omitempty"`
	FinishedAt     *string `json:"finished_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// rowToResponse converts a CycleRow to the API response format.
func rowToResponse(r *repo.CycleRow) *CycleResponse {
	resp := &CycleResponse{
		ID:             r.ID.String(),
		Year:           r.Year,
		OrganizationID: r.OrganizationID.String(),
		CurrentPhase:   string(r.CurrentPhase),
		Version:        r.Version,
		CreatedAt:      r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      r.UpdatedAt.Format(time.RFC3339),
	}
	if r.StartedAt != nil {
		s := r.StartedAt.Format(time.RFC3339)
		resp.StartedAt = &s
	}
	if r.FinishedAt != nil {
		s := r.FinishedAt.Format(time.RFC3339)
		resp.FinishedAt = &s
	}
	return resp
}

// Service defines the interface for cycle business operations.
type Service interface {
	CreateCycle(ctx context.Context, req CreateCycleRequest) (*CycleResponse, error)
	TransitionPhase(ctx context.Context, req TransitionPhaseRequest) (*CycleResponse, error)
	GetCycle(ctx context.Context, cycleID string) (*CycleResponse, error)
	ListCycles(ctx context.Context, req ListCyclesRequest) (*cursor.PaginatedList[*CycleResponse], error)
}

// CycleRepository defines the methods used by service from CycleRepo.
type CycleRepository interface {
	ExecuteRawAdvisoryLock(ctx context.Context, orgID uuid.UUID, year int) (func() error, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	CheckExistingCycle(ctx context.Context, tx *sql.Tx, orgID uuid.UUID, year int) (bool, error)
	CreateCycle(ctx context.Context, tx *sql.Tx, year int, orgID uuid.UUID) (*repo.CycleRow, error)
	LockCycleForUpdate(ctx context.Context, tx *sql.Tx, cycleID uuid.UUID) (*repo.CycleRow, error)
	UpdatePhase(ctx context.Context, tx *sql.Tx, cycleID uuid.UUID, nextPhase cycle.CurrentPhase, expectedVersion int) error
	GetCycle(ctx context.Context, id uuid.UUID) (*repo.CycleRow, error)
	InsertPhaseHistory(ctx context.Context, tx *sql.Tx, cycleID uuid.UUID, fromPhase, toPhase string, triggeredBy uuid.UUID, reason string) error
	ListCycles(ctx context.Context, orgID uuid.UUID, year *int, phase *cycle.CurrentPhase, cursorID *uuid.UUID, cursorUpdatedAt *time.Time, limit int) ([]*repo.CycleRow, error)
}

// PhaseRepository defines the methods used by service from PhaseRepo.
type PhaseRepository interface {
	ValidateTransition(ctx context.Context, fromPhase, toPhase, trigger string) error
}

// service implements Service.
type service struct {
	cycleRepo CycleRepository
	phaseRepo PhaseRepository
	entClient *internal.Client
}

// NewService creates a new cycle service.
func NewService(cycleRepo CycleRepository, phaseRepo PhaseRepository, entClient *internal.Client) Service {
	return &service{
		cycleRepo: cycleRepo,
		phaseRepo: phaseRepo,
		entClient: entClient,
	}
}

// CreateCycle validates the request, acquires an advisory lock, checks for
// duplicates inside a transaction, and creates the cycle.
func (s *service) CreateCycle(ctx context.Context, req CreateCycleRequest) (*CycleResponse, error) {
	// Validate request
	if req.Year < 2000 || req.Year > 2100 {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"year must be between 2000 and 2100", nil)
	}
	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"organization_id must be a valid UUID v4", err)
	}

	// Acquire PostgreSQL advisory lock
	unlock, err := s.cycleRepo.ExecuteRawAdvisoryLock(ctx, orgID, req.Year)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = unlock()
	}()

	// Start transaction
	tx, err := s.cycleRepo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	// Check for existing cycle
	exists, err := s.cycleRepo.CheckExistingCycle(ctx, tx, orgID, req.Year)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, pkgerrors.ErrCycleAlreadyActive
	}

	// Create cycle
	row, err := s.cycleRepo.CreateCycle(ctx, tx, req.Year, orgID)
	if err != nil {
		return nil, err
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil // prevent deferred rollback

	return rowToResponse(row), nil
}

// TransitionPhase transitions a cycle's phase within a transaction.
// It locks the row, validates the optimistic lock, resolves the next phase,
// validates the transition rule, updates, and writes the audit log.
func (s *service) TransitionPhase(ctx context.Context, req TransitionPhaseRequest) (*CycleResponse, error) {
	cycleID, err := uuid.Parse(req.CycleID)
	if err != nil {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"cycle_id must be a valid UUID v4", err)
	}

	trigger := req.Trigger
	if trigger == "" {
		trigger = "manual_rh"
	}

	// Start transaction
	tx, err := s.cycleRepo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	// Step 1: Lock cycle row with SELECT FOR UPDATE
	row, err := s.cycleRepo.LockCycleForUpdate(ctx, tx, cycleID)
	if err != nil {
		return nil, err
	}

	// Step 2: Validate optimistic lock
	if row.Version != req.ExpectedVersion {
		return nil, pkgerrors.ErrConcurrentUpdate.WithDetails(
			"expected_version: " + itoa(req.ExpectedVersion),
			"actual_version: " + itoa(row.Version),
		)
	}

	// Step 3: Resolve next phase (linear)
	nextPhase, ok := resolveNextPhase(row.CurrentPhase)
	if !ok {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidTransition,
			"the current phase '"+string(row.CurrentPhase)+"' has no next phase; the cycle is at its final phase", nil,
		).WithDetails("current_phase: " + string(row.CurrentPhase))
	}

	// Step 4: Validate transition exists
	err = s.phaseRepo.ValidateTransition(ctx, string(row.CurrentPhase), string(nextPhase), trigger)
	if err != nil {
		return nil, err
	}

	// Step 5: Update cycle with optimistic lock
	err = s.cycleRepo.UpdatePhase(ctx, tx, cycleID, nextPhase, req.ExpectedVersion)
	if err != nil {
		return nil, err
	}

	// Step 6: Insert audit log
	triggeredBy, _ := uuid.Parse("00000000-0000-0000-0000-000000000000") // TODO(auth:C7): use real employee ID
	err = s.cycleRepo.InsertPhaseHistory(ctx, tx, cycleID, string(row.CurrentPhase), string(nextPhase), triggeredBy, req.Reason)
	if err != nil {
		return nil, err
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	// Fetch updated cycle
	updatedRow, err := s.cycleRepo.GetCycle(ctx, cycleID)
	if err != nil {
		return nil, err
	}

	return rowToResponse(updatedRow), nil
}

// GetCycle retrieves a single cycle by ID.
func (s *service) GetCycle(ctx context.Context, cycleID string) (*CycleResponse, error) {
	id, err := uuid.Parse(cycleID)
	if err != nil {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"cycle_id must be a valid UUID v4", err)
	}

	row, err := s.cycleRepo.GetCycle(ctx, id)
	if err != nil {
		return nil, err
	}

	return rowToResponse(row), nil
}

// ListCycles decodes the cursor, delegates to the repository, and builds the
// paginated response.
func (s *service) ListCycles(ctx context.Context, req ListCyclesRequest) (*cursor.PaginatedList[*CycleResponse], error) {
	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"organization_id must be a valid UUID v4", err)
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}

	// Decode cursor
	var cursorID *uuid.UUID
	var cursorUpdatedAt *time.Time
	if req.Cursor != "" {
		c, err := cursor.DecodeCursor(req.Cursor)
		if err != nil {
			return nil, err
		}
		cursorID = &c.ID
		cursorUpdatedAt = &c.UpdatedAt
	}

	// Determine phase filter
	var phaseFilter *cycle.CurrentPhase
	if req.CurrentPhase != nil {
		p := cycle.CurrentPhase(*req.CurrentPhase)
		phaseFilter = &p
	}

	// Query
	rows, err := s.cycleRepo.ListCycles(ctx, orgID, req.Year, phaseFilter, cursorID, cursorUpdatedAt, limit)
	if err != nil {
		return nil, err
	}

	// Convert to response
	responses := make([]*CycleResponse, len(rows))
	for i, r := range rows {
		responses[i] = rowToResponse(r)
	}

	// Build paginated list
	return cursor.NewPaginatedList(responses, limit, func(cr *CycleResponse) (uuid.UUID, time.Time) {
		uid, _ := uuid.Parse(cr.ID)
		t, _ := time.Parse(time.RFC3339, cr.UpdatedAt)
		return uid, t
	})
}

// itoa is a small helper for integer to string conversion.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := make([]byte, 0, 10)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
