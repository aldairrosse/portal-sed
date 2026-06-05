package goal

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/goalassignment"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// AssignmentRow is the full representation of a GoalAssignment.
type AssignmentRow struct {
	ID         uuid.UUID `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	EmployeeID uuid.UUID `json:"employee_id"`
	CycleID    uuid.UUID `json:"cycle_id"`
}

// assignmentToRow converts an ent GoalAssignment to an AssignmentRow.
func assignmentToRow(a *internal.GoalAssignment) *AssignmentRow {
	if a == nil {
		return nil
	}
	return &AssignmentRow{
		ID:         a.ID,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
		EmployeeID: a.EmployeeID,
		CycleID:    a.CycleID,
	}
}

// AssignmentRepo provides Ent-backed operations for GoalAssignment.
type AssignmentRepo struct {
	client *internal.Client
	db     *sql.DB
}

// NewAssignmentRepo creates a new AssignmentRepo.
func NewAssignmentRepo(client *internal.Client, db *sql.DB) *AssignmentRepo {
	return &AssignmentRepo{client: client, db: db}
}

// GetAssignment retrieves the assignment for an employee.
func (r *AssignmentRepo) GetAssignment(ctx context.Context, empID uuid.UUID) (*AssignmentRow, error) {
	a, err := r.client.GoalAssignment.Query().
		Where(goalassignment.EmployeeID(empID)).
		Order(internal.Desc(goalassignment.FieldCreatedAt)).
		First(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, pkgerrors.ErrGoalNotFound // reuse as "no assignment"
		}
		return nil, err
	}
	return assignmentToRow(a), nil
}

// GetAssignmentByEmployeeAndCycle retrieves an assignment by employee and cycle.
func (r *AssignmentRepo) GetAssignmentByEmployeeAndCycle(ctx context.Context, empID, cycleID uuid.UUID) (*AssignmentRow, error) {
	a, err := r.client.GoalAssignment.Query().
		Where(goalassignment.EmployeeID(empID), goalassignment.CycleID(cycleID)).
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return assignmentToRow(a), nil
}

// CreateAssignment inserts a new assignment with advisory lock to prevent duplicates.
func (r *AssignmentRepo) CreateAssignment(ctx context.Context, empID, cycleID uuid.UUID) (*AssignmentRow, error) {
	// Acquire advisory lock
	lockKey := hashEmployeeCycle(empID, cycleID)
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, lockKey)
	if err != nil {
		return nil, fmt.Errorf("advisory lock acquire failed: %w", err)
	}
	defer func() {
		_, _ = conn.ExecContext(ctx, `SELECT pg_advisory_unlock($1)`, lockKey)
	}()

	// Check for existing assignment (idempotent)
	existing, err := r.GetAssignmentByEmployeeAndCycle(ctx, empID, cycleID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	// Create assignment via Ent
	a, err := r.client.GoalAssignment.Create().
		SetEmployeeID(empID).
		SetCycleID(cycleID).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return assignmentToRow(a), nil
}

// hashEmployeeCycle creates a deterministic int64 hash from employee_id and cycle_id.
func hashEmployeeCycle(empID, cycleID uuid.UUID) int64 {
	h := sha256.New()
	h.Write(empID[:])
	h.Write(cycleID[:])
	sum := h.Sum(nil)
	// Use first 8 bytes as int64
	val := int64(binary.BigEndian.Uint64(sum[:8]))
	if val < 0 {
		val = -val
	}
	return val
}
