package goal

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/cycle"
	"github.com/sed-evaluacion-desempeno/api/internal/goalassignment"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// AssignmentRow is the full representation of a GoalAssignment.
type AssignmentRow struct {
	ID          uuid.UUID  `json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	EmployeeID  uuid.UUID  `json:"employee_id"`
	CycleID     uuid.UUID  `json:"cycle_id"`
	Status      string     `json:"status"`
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
}

// assignmentToRow converts an ent GoalAssignment to an AssignmentRow.
func assignmentToRow(a *internal.GoalAssignment) *AssignmentRow {
	if a == nil {
		return nil
	}
	return &AssignmentRow{
		ID:          a.ID,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
		EmployeeID:  a.EmployeeID,
		CycleID:     a.CycleID,
		Status:      string(a.Status),
		SubmittedAt: a.SubmittedAt,
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

func (r *AssignmentRepo) ListGlobalGoalsByEmployee(ctx context.Context, empID uuid.UUID) ([]*GlobalGoalRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT g.id, g.name, g.description, g.unit, g.direction, g.goal_kind, g.state,
		       g.weight, g.target_value, g.current_value, a.weight, a.target_value, a.baseline_value
		FROM goals g JOIN global_goal_assignments a ON a.goal_id = g.id
		JOIN employees e ON e.id = a.employee_id AND e.is_active = true
		WHERE g.type = 'global' AND a.employee_id = $1 ORDER BY g.created_at DESC`, empID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*GlobalGoalRow, 0)
	for rows.Next() {
		var g GlobalGoalRow
		var a GlobalAssignmentRow
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.Unit, &g.Direction, &g.GoalKind, &g.State, &g.Weight, &g.TargetValue, &g.CurrentValue, &a.Weight, &a.TargetValue, &a.BaselineValue); err != nil {
			return nil, err
		}
		a.GoalID, a.EmployeeID = g.ID, empID
		g.Assignments = []*GlobalAssignmentRow{&a}
		result = append(result, &g)
	}
	return result, rows.Err()
}

func (r *AssignmentRepo) ListSharedGoalsAsMember(ctx context.Context, empID uuid.UUID) ([]*SharedGoalRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT g.id, g.name, g.description, g.unit, g.direction, g.goal_kind, g.state,
		       g.weight, g.target_value, g.current_value, m.weight, m.target_value, m.baseline_value
		FROM goals g JOIN shared_goal_groups sg ON sg.goal_id = g.id
		JOIN shared_goal_members m ON m.group_id = sg.id
		JOIN employees e ON e.id = m.employee_id AND e.is_active = true
		WHERE g.type = 'shared' AND m.employee_id = $1 ORDER BY g.created_at DESC`, empID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*SharedGoalRow, 0)
	for rows.Next() {
		var g SharedGoalRow
		var m SharedMemberRow
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.Unit, &g.Direction, &g.GoalKind, &g.State, &g.Weight, &g.TargetValue, &g.CurrentValue, &m.Weight, &m.TargetValue, &m.BaselineValue); err != nil {
			return nil, err
		}
		m.EmployeeID = empID
		g.Members = []*SharedMemberRow{&m}
		result = append(result, &g)
	}
	return result, rows.Err()
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
	exists, err := r.client.Cycle.Query().Where(cycle.ID(cycleID)).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, pkgerrors.ErrCycleNotFound
	}

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

// UpdateAssignmentStatus transitions an assignment to the given status.
func (r *AssignmentRepo) UpdateAssignmentStatus(ctx context.Context, empID, cycleID uuid.UUID, status string, submittedAt *time.Time) (*AssignmentRow, error) {
	a, err := r.client.GoalAssignment.Query().
		Where(goalassignment.EmployeeID(empID), goalassignment.CycleID(cycleID)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	upd := r.client.GoalAssignment.UpdateOneID(a.ID).SetStatus(goalassignment.Status(status))
	if submittedAt != nil {
		upd = upd.SetSubmittedAt(*submittedAt)
	} else {
		upd = upd.ClearSubmittedAt()
	}
	saved, err := upd.Save(ctx)
	if err != nil {
		return nil, err
	}
	return assignmentToRow(saved), nil
}

// BatchGetByEmployeeIDs returns a map of employee_id -> status for a given cycle.
// Normalizes empty status to "borrador". Missing rows are not in the map (caller maps to no_iniciado).
func (r *AssignmentRepo) BatchGetByEmployeeIDs(ctx context.Context, empIDs []uuid.UUID, cycleID uuid.UUID) (map[uuid.UUID]string, error) {
	if len(empIDs) == 0 || cycleID == uuid.Nil {
		return map[uuid.UUID]string{}, nil
	}
	placeholders := make([]string, len(empIDs))
	args := make([]interface{}, len(empIDs)+1)
	args[0] = cycleID
	for i, id := range empIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}
	query := fmt.Sprintf(`SELECT employee_id, status FROM goal_assignments WHERE cycle_id=$1 AND employee_id IN (%s)`, strings.Join(placeholders, ","))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[uuid.UUID]string, len(empIDs))
	for rows.Next() {
		var eid uuid.UUID
		var status sql.NullString
		if err := rows.Scan(&eid, &status); err != nil {
			return nil, err
		}
		s := ""
		if status.Valid {
			s = strings.TrimSpace(status.String)
		}
		if s == "" {
			s = "borrador"
		}
		m[eid] = s
	}
	return m, rows.Err()
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
