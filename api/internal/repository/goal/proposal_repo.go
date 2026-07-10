package goal

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// GoalProposalRow is the full representation of a goal proposal.
type GoalProposalRow struct {
	ID            uuid.UUID  `json:"id"`
	GoalID        uuid.UUID  `json:"goal_id"`
	RequestedBy   uuid.UUID  `json:"requested_by"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Unit          string     `json:"unit"`
	Direction     string     `json:"direction"`
	Weight        float64    `json:"weight"`
	TargetValue   float64    `json:"target_value"`
	BaselineValue *float64   `json:"baseline_value,omitempty"`
	Status        string     `json:"status"`
	ReviewedBy    *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// GoalProposalRepo provides raw SQL operations for goal proposals.
type GoalProposalRepo struct {
	db *sql.DB
}

// NewGoalProposalRepo creates a new GoalProposalRepo.
func NewGoalProposalRepo(db *sql.DB) *GoalProposalRepo {
	return &GoalProposalRepo{db: db}
}

// scanProposalRow scans a proposal row from the database.
func scanProposalRow(scanner interface {
	Scan(dest ...interface{}) error
}) (*GoalProposalRow, error) {
	var p GoalProposalRow
	var createdAt, updatedAt sql.NullTime
	var description sql.NullString
	var baselineValue sql.NullFloat64
	var reviewedBy sql.NullString
	var reviewedAt sql.NullTime

	err := scanner.Scan(
		&p.ID, &createdAt, &updatedAt,
		&p.GoalID, &p.RequestedBy,
		&p.Name, &description,
		&p.Unit, &p.Direction,
		&p.Weight, &p.TargetValue, &baselineValue,
		&p.Status, &reviewedBy, &reviewedAt,
	)
	if err != nil {
		return nil, err
	}
	if createdAt.Valid {
		p.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		p.UpdatedAt = updatedAt.Time
	}
	if description.Valid {
		p.Description = description.String
	}
	if baselineValue.Valid {
		p.BaselineValue = &baselineValue.Float64
	}
	if reviewedBy.Valid {
		id, parseErr := uuid.Parse(reviewedBy.String)
		if parseErr == nil {
			p.ReviewedBy = &id
		}
	}
	if reviewedAt.Valid {
		p.ReviewedAt = &reviewedAt.Time
	}
	return &p, nil
}

// buildProposalColumns returns the SELECT column list and the scan targets order used by scanProposalRow.
const proposalColumns = `id, created_at, updated_at, goal_id, requested_by,
	name, COALESCE(description, ''),
	unit, direction, weight, target_value, baseline_value,
	status, reviewed_by, reviewed_at`

// Create inserts a new goal proposal and its KPI links.
func (r *GoalProposalRepo) Create(ctx context.Context, goalID, requestedBy uuid.UUID, name, description, unit, direction string, weight, targetValue float64, baselineValue *float64) (*GoalProposalRow, error) {
	now := time.Now()
	id := uuid.New()

	var desc interface{} = nil
	if description != "" {
		desc = description
	}

	row := r.db.QueryRowContext(ctx,
		`INSERT INTO goal_proposals (id, created_at, updated_at, goal_id, requested_by,
		 name, description, unit, direction, weight, target_value, baseline_value)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 RETURNING `+proposalColumns,
		id, now, now, goalID, requestedBy,
		name, desc,
		unit, direction, weight, targetValue, baselineValue,
	)
	return scanProposalRow(row)
}

// GetByID retrieves a single proposal by ID.
func (r *GoalProposalRepo) GetByID(ctx context.Context, proposalID uuid.UUID) (*GoalProposalRow, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+proposalColumns+` FROM goal_proposals WHERE id = $1`,
		proposalID,
	)
	p, err := scanProposalRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrGoalNotFound
		}
		return nil, err
	}
	return p, nil
}

// ListByGoal retrieves all proposals for a goal, newest first.
func (r *GoalProposalRepo) ListByGoal(ctx context.Context, goalID uuid.UUID) ([]*GoalProposalRow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+proposalColumns+` FROM goal_proposals WHERE goal_id = $1 ORDER BY created_at DESC`,
		goalID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*GoalProposalRow
	for rows.Next() {
		p, err := scanProposalRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

// ListPendingByGoalIDs retrieves all pending proposals for a set of goal IDs.
// Returns proposals with status='pending', newest per goal first.
func (r *GoalProposalRepo) ListPendingByGoalIDs(ctx context.Context, goalIDs []uuid.UUID) ([]*GoalProposalRow, error) {
	if len(goalIDs) == 0 {
		return nil, nil
	}

	args := make([]interface{}, len(goalIDs))
	placeholders := make([]string, len(goalIDs))
	for i, id := range goalIDs {
		args[i] = id
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(
		`SELECT `+proposalColumns+` FROM goal_proposals
		 WHERE goal_id IN (%s) AND status = 'pending'
		 ORDER BY created_at DESC`,
		strings.Join(placeholders, ", "),
	)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*GoalProposalRow
	for rows.Next() {
		p, err := scanProposalRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

// UpdateStatus updates the proposal status and reviewer.
func (r *GoalProposalRepo) UpdateStatus(ctx context.Context, proposalID uuid.UUID, status string, reviewedBy uuid.UUID) (*GoalProposalRow, error) {
	now := time.Now()
	reviewedByStr := reviewedBy.String()
	row := r.db.QueryRowContext(ctx,
		`UPDATE goal_proposals
		 SET status = $1, reviewed_by = $2, reviewed_at = $3, updated_at = $4
		 WHERE id = $5
		 RETURNING `+proposalColumns,
		status, reviewedByStr, now, now, proposalID,
	)
	p, err := scanProposalRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrGoalNotFound
		}
		return nil, err
	}
	return p, nil
}

// RejectPendingByGoalExcept rejects all pending proposals for a goal except the given one.
func (r *GoalProposalRepo) RejectPendingByGoalExcept(ctx context.Context, goalID, exceptProposalID, reviewedBy uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE goal_proposals
		 SET status = 'rejected', reviewed_by = $1, reviewed_at = now(), updated_at = now()
		 WHERE goal_id = $2 AND status = 'pending' AND id != $3`,
		reviewedBy.String(), goalID, exceptProposalID,
	)
	return err
}

// GetKpiIDs returns the KPI IDs linked to a proposal.
func (r *GoalProposalRepo) GetKpiIDs(ctx context.Context, proposalID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT kpi_id FROM goal_proposal_kpi_links WHERE proposal_id = $1`,
		proposalID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// SetKpis replaces all KPI links for a proposal.
func (r *GoalProposalRepo) SetKpis(ctx context.Context, proposalID uuid.UUID, kpiIDs []uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM goal_proposal_kpi_links WHERE proposal_id = $1`,
		proposalID,
	)
	if err != nil {
		return err
	}

	for _, kpiID := range kpiIDs {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO goal_proposal_kpi_links (proposal_id, kpi_id)
			 VALUES ($1, $2)
			 ON CONFLICT (proposal_id, kpi_id) DO NOTHING`,
			proposalID, kpiID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
