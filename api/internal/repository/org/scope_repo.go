package org

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
)

// EvaluatorScopeRow is a read model for evaluator scope.
type EvaluatorScopeRow struct {
	ID          uuid.UUID              `json:"id"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	EvaluatorID uuid.UUID              `json:"evaluator_id"`
	CycleID     *uuid.UUID             `json:"cycle_id,omitempty"`
	ScopeType   string                 `json:"scope_type"`
	ScopeData   map[string]interface{} `json:"scope_data,omitempty"`
}

// EvaluatorScopeRepo provides database queries for evaluator scopes.
type EvaluatorScopeRepo struct {
	client *internal.Client
	db     *sql.DB
}

// NewEvaluatorScopeRepo creates a new EvaluatorScopeRepo.
func NewEvaluatorScopeRepo(client *internal.Client, db *sql.DB) *EvaluatorScopeRepo {
	return &EvaluatorScopeRepo{client: client, db: db}
}

// GetByEvaluatorAndCycle returns the scope for a given evaluator and cycle.
func (r *EvaluatorScopeRepo) GetByEvaluatorAndCycle(ctx context.Context, evaluatorID, cycleID uuid.UUID) (*EvaluatorScopeRow, error) {
	return scanScopeRow(r.db.QueryRowContext(ctx,
		`SELECT id, created_at, updated_at, evaluator_id, cycle_id, scope_type, scope_data
		 FROM evaluator_scopes WHERE evaluator_id = $1 AND cycle_id = $2`,
		evaluatorID, cycleID,
	))
}

// GetByEvaluator returns all scopes for a given evaluator.
func (r *EvaluatorScopeRepo) GetByEvaluator(ctx context.Context, evaluatorID uuid.UUID) ([]*EvaluatorScopeRow, error) {
	return queryScopeRows(r.db, ctx,
		`SELECT id, created_at, updated_at, evaluator_id, cycle_id, scope_type, scope_data
		 FROM evaluator_scopes WHERE evaluator_id = $1`, evaluatorID)
}

// GetByID returns a single scope by ID.
func (r *EvaluatorScopeRepo) GetByID(ctx context.Context, scopeID uuid.UUID) (*EvaluatorScopeRow, error) {
	return scanScopeRow(r.db.QueryRowContext(ctx,
		`SELECT id, created_at, updated_at, evaluator_id, cycle_id, scope_type, scope_data
		 FROM evaluator_scopes WHERE id = $1`, scopeID))
}

// ---------- helpers ----------

func scanScopeRow(row *sql.Row) (*EvaluatorScopeRow, error) {
	s := &EvaluatorScopeRow{}
	var cycleID sql.NullString
	var scopeData sql.NullString

	err := row.Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt, &s.EvaluatorID, &cycleID, &s.ScopeType, &scopeData)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrScopeNotFound
		}
		return nil, err
	}
	if cycleID.Valid {
		cid, _ := uuid.Parse(cycleID.String)
		s.CycleID = &cid
	}
	if scopeData.Valid && scopeData.String != "" {
		_ = json.Unmarshal([]byte(scopeData.String), &s.ScopeData)
	}
	return s, nil
}

func queryScopeRows(db *sql.DB, ctx context.Context, query string, args ...interface{}) ([]*EvaluatorScopeRow, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*EvaluatorScopeRow
	for rows.Next() {
		s := &EvaluatorScopeRow{}
		var cycleID sql.NullString
		var scopeData sql.NullString
		err := rows.Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt, &s.EvaluatorID, &cycleID, &s.ScopeType, &scopeData)
		if err != nil {
			return nil, err
		}
		if cycleID.Valid {
			cid, _ := uuid.Parse(cycleID.String)
			s.CycleID = &cid
		}
		if scopeData.Valid && scopeData.String != "" {
			_ = json.Unmarshal([]byte(scopeData.String), &s.ScopeData)
		}
		results = append(results, s)
	}
	return results, rows.Err()
}
