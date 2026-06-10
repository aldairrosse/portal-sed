// Package evaluation provides repository-level CRUD and transactional operations
// for Evaluation, EvaluationCompetency, EvaluationGoal, and the 9×9 matrix
// entities. It uses Ent-generated queries for reads and raw SQL for locking.
package evaluation

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/evaluation"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/cursor"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// contextKey for db role routing.
type ctxKeyDBRole struct{}

const (
	DBRolePrimary = "primary"
	DBRoleReplica = "replica"
)

// WithDBRole embeds a db role hint into the context.
func WithDBRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, ctxKeyDBRole{}, role)
}

// DBRoleFromContext extracts the db role from context; returns "primary" if not set.
func DBRoleFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyDBRole{}).(string)
	if v == "" {
		return DBRolePrimary
	}
	return v
}

// Domain error codes.
const (
	ErrCodeEvaluationNotFound     pkgerrors.DomainCode = "EVALUATION_NOT_FOUND"
	ErrCodeMatrixNotFound         pkgerrors.DomainCode = "MATRIX_NOT_FOUND"
	ErrCodeEntryNotFound          pkgerrors.DomainCode = "ENTRY_NOT_FOUND"
	ErrCodeEvaluationFinalized    pkgerrors.DomainCode = "EVALUATION_ALREADY_FINALIZED"
	ErrCodeSelfEvalDeadlinePassed pkgerrors.DomainCode = "SELF_EVAL_DEADLINE_PASSED"
	ErrCodeQuadrantOutOfRange     pkgerrors.DomainCode = "QUADRANT_OUT_OF_RANGE"
	ErrCodeUnauthorizedEvaluator  pkgerrors.DomainCode = "UNAUTHORIZED_EVALUATOR"
)

// Sentinel errors.
var (
	ErrEvaluationNotFound     = pkgerrors.NewDomainError(ErrCodeEvaluationNotFound, "The requested evaluation was not found.", nil)
	ErrMatrixNotFound         = pkgerrors.NewDomainError(ErrCodeMatrixNotFound, "The requested 9×9 matrix was not found.", nil)
	ErrEntryNotFound          = pkgerrors.NewDomainError(ErrCodeEntryNotFound, "The requested matrix entry was not found.", nil)
	ErrEvaluationFinalized    = pkgerrors.NewDomainError(ErrCodeEvaluationFinalized, "The evaluation has already been finalized; no further changes allowed.", nil)
	ErrSelfEvalDeadlinePassed = pkgerrors.NewDomainError(ErrCodeSelfEvalDeadlinePassed, "The self-evaluation deadline has passed for this cycle.", nil)
	ErrQuadrantOutOfRange     = pkgerrors.NewDomainError(ErrCodeQuadrantOutOfRange, "Performance and potential scores must be between 1 and 9.", nil)
	ErrUnauthorizedEvaluator  = pkgerrors.NewDomainError(ErrCodeUnauthorizedEvaluator, "The authenticated user is not the evaluator for this matrix.", nil)
)

// CompetencyUpsert is a repository-level DTO for upserting a competency rating.
type CompetencyUpsert struct {
	CompetencyID uuid.UUID
	Rating       int
	Comments     string
}

// GoalCommentUpsert is a repository-level DTO for updating goal comments.
type GoalCommentUpsert struct {
	GoalID  uuid.UUID
	Comment string
}

// EntryUpsert is a repository-level DTO for upserting a nine-box entry.
type EntryUpsert struct {
	EvaluateeID      uuid.UUID
	PerformanceScore int
	PotentialScore   int
	Quadrant         int
	Comments         string
}

// EvaluationRow is a full representation of an evaluation including version.
type EvaluationRow struct {
	ID                        uuid.UUID
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	Phase                     string
	State                     string
	SelfEvaluationCompletedAt *time.Time
	RhEvaluationCompletedAt   *time.Time
	EmployeeID                uuid.UUID
	CycleID                   uuid.UUID
	Version                   int
}

// EvaluationRepo provides repository operations for evaluations.
type EvaluationRepo struct {
	client *internal.Client
	db     *sql.DB
}

// NewEvaluationRepo creates a new EvaluationRepo.
func NewEvaluationRepo(client *internal.Client, db *sql.DB) *EvaluationRepo {
	return &EvaluationRepo{client: client, db: db}
}

// GetByID retrieves an evaluation by ID with version.
func (r *EvaluationRepo) GetByID(ctx context.Context, id uuid.UUID) (*EvaluationRow, error) {
	ev, err := r.client.Evaluation.Query().
		Where(evaluation.ID(id)).
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, ErrEvaluationNotFound
		}
		return nil, err
	}

	return rowFromEnt(ev, ev.Version), nil
}

// GetDetail retrieves an evaluation with preloaded competency and goal ratings.
func (r *EvaluationRepo) GetDetail(ctx context.Context, id uuid.UUID) (*EvaluationRow, []*internal.EvaluationCompetency, []*internal.EvaluationGoal, error) {
	ev, err := r.client.Evaluation.Query().
		Where(evaluation.ID(id)).
		WithCompetencyRatings().
		WithGoalRatings().
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, nil, nil, ErrEvaluationNotFound
		}
		return nil, nil, nil, err
	}

	row := rowFromEnt(ev, ev.Version)

	comps := ev.Edges.CompetencyRatings
	if comps == nil {
		comps = []*internal.EvaluationCompetency{}
	}
	goals := ev.Edges.GoalRatings
	if goals == nil {
		goals = []*internal.EvaluationGoal{}
	}

	return row, comps, goals, nil
}

// ListByCycle returns cursor-paginated evaluations for a cycle.
func (r *EvaluationRepo) ListByCycle(ctx context.Context, cycleID uuid.UUID, state string, cursorStr string, limit int) ([]*EvaluationRow, string, error) {
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}

	var cID *uuid.UUID
	var cUpdatedAt *time.Time
	if cursorStr != "" {
		c, err := cursor.DecodeCursor(cursorStr)
		if err != nil {
			return nil, "", err
		}
		cID = &c.ID
		cUpdatedAt = &c.UpdatedAt
	}

	query := `SELECT e.id, e.created_at, e.updated_at, e.phase, e.state,
		e.self_evaluation_completed_at, e.rh_evaluation_completed_at,
		e.employee_id, e.cycle_id, e.version
		FROM evaluations e
		WHERE e.cycle_id = $1`
	args := []interface{}{cycleID}
	idx := 2

	if state != "" {
		query += ` AND e.state = $` + strconv.Itoa(idx)
		args = append(args, state)
		idx++
	}

	if cID != nil && cUpdatedAt != nil {
		query += ` AND (e.updated_at, e.id) < ($` + strconv.Itoa(idx) + `, $` + strconv.Itoa(idx+1) + `)`
		args = append(args, *cUpdatedAt, *cID)
		idx += 2
	}

	query += ` ORDER BY e.updated_at DESC, e.id DESC LIMIT $` + strconv.Itoa(idx)
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var results []*EvaluationRow
	for rows.Next() {
		row := &EvaluationRow{}
		var stateStr, phaseStr string
		var selfComp, rhComp sql.NullTime
		err := rows.Scan(&row.ID, &row.CreatedAt, &row.UpdatedAt,
			&phaseStr, &stateStr, &selfComp, &rhComp,
			&row.EmployeeID, &row.CycleID, &row.Version)
		if err != nil {
			return nil, "", err
		}
		row.Phase = phaseStr
		row.State = stateStr
		if selfComp.Valid {
			row.SelfEvaluationCompletedAt = &selfComp.Time
		}
		if rhComp.Valid {
			row.RhEvaluationCompletedAt = &rhComp.Time
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	hasMore := len(results) > limit
	if hasMore {
		results = results[:limit]
	}

	nextCursor := ""
	if hasMore && len(results) > 0 {
		last := results[len(results)-1]
		c := &cursor.Cursor{ID: last.ID, UpdatedAt: last.UpdatedAt}
		nextCursor, err = c.Encode()
		if err != nil {
			return nil, "", err
		}
	}

	return results, nextCursor, nil
}

// LockEvalForUpdate locks the evaluation row with SELECT FOR UPDATE inside a tx.
func (r *EvaluationRepo) LockEvalForUpdate(ctx context.Context, tx *sql.Tx, evalID uuid.UUID) (*EvaluationRow, error) {
	row := &EvaluationRow{}
	var stateStr, phaseStr string
	var selfComp, rhComp sql.NullTime

	err := tx.QueryRowContext(ctx,
		`SELECT e.id, e.created_at, e.updated_at, e.phase, e.state,
			e.self_evaluation_completed_at, e.rh_evaluation_completed_at,
			e.employee_id, e.cycle_id, e.version
		 FROM evaluations e
		 WHERE e.id = $1 FOR UPDATE`,
		evalID,
	).Scan(&row.ID, &row.CreatedAt, &row.UpdatedAt,
		&phaseStr, &stateStr, &selfComp, &rhComp,
		&row.EmployeeID, &row.CycleID, &row.Version)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrEvaluationNotFound
		}
		return nil, err
	}

	row.Phase = phaseStr
	row.State = stateStr
	if selfComp.Valid {
		row.SelfEvaluationCompletedAt = &selfComp.Time
	}
	if rhComp.Valid {
		row.RhEvaluationCompletedAt = &rhComp.Time
	}
	return row, nil
}

// SubmitEval performs the atomic evaluation submission inside a *sql.Tx.
// It upserts competencies, updates goal comments, sets state and timestamps.
func (r *EvaluationRepo) SubmitEval(ctx context.Context, tx *sql.Tx, evalID uuid.UUID, comps []CompetencyUpsert, goals []GoalCommentUpsert, newState string, setSelfCompleted, setRHCompleted bool) error {
	// 1. Lock row and validate state
	row, err := r.LockEvalForUpdate(ctx, tx, evalID)
	if err != nil {
		return err
	}
	if row.State == "completada" {
		return ErrEvaluationFinalized
	}

	now := time.Now()

	// 2. Bulk upsert competencies
	for _, c := range comps {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO evaluation_competencies (id, created_at, updated_at, evaluation_id, competency_id, rating, comments, profile_id)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT (evaluation_id, competency_id) DO UPDATE
			 SET rating = EXCLUDED.rating, comments = EXCLUDED.comments, updated_at = EXCLUDED.updated_at`,
			uuid.New(), now, now, evalID, c.CompetencyID, c.Rating, c.Comments, uuid.Nil,
		)
		if err != nil {
			return err
		}
	}

	// 3. Update goal comments
	for _, g := range goals {
		_, err = tx.ExecContext(ctx,
			`UPDATE evaluation_goals SET final_comments = $1, updated_at = $2
			 WHERE evaluation_id = $3 AND goal_id = $4`,
			g.Comment, now, evalID, g.GoalID,
		)
		if err != nil {
			return err
		}
	}

	// 4. Update evaluation state
	setClauses := `state = $1, updated_at = $2`
	args := []interface{}{newState, now}
	argIdx := 3

	if setSelfCompleted {
		setClauses += fmt.Sprintf(`, self_evaluation_completed_at = $%d`, argIdx)
		args = append(args, now)
		argIdx++
	}
	if setRHCompleted {
		setClauses += fmt.Sprintf(`, rh_evaluation_completed_at = $%d`, argIdx)
		args = append(args, now)
		argIdx++
	}

	query := fmt.Sprintf(`UPDATE evaluations SET %s WHERE id = $%d`, setClauses, argIdx)
	args = append(args, evalID)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	// 5. Increment version
	return r.upsertVersion(ctx, tx, evalID)
}

// FinalizeEval sets the evaluation state to completada.
func (r *EvaluationRepo) FinalizeEval(ctx context.Context, tx *sql.Tx, evalID uuid.UUID) error {
	now := time.Now()
	_, err := tx.ExecContext(ctx,
		`UPDATE evaluations SET state = $1, rh_evaluation_completed_at = $2, updated_at = $2 WHERE id = $3`,
		"completada", now, evalID,
	)
	if err != nil {
		return err
	}
	return r.upsertVersion(ctx, tx, evalID)
}

// GetSummaryByCycle queries the evaluation_summary materialized view.
func (r *EvaluationRepo) GetSummaryByCycle(ctx context.Context, cycleID uuid.UUID) (map[string]int64, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT state, count FROM evaluation_summary WHERE cycle_id = $1`, cycleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int64)
	for rows.Next() {
		var state string
		var count int64
		if err := rows.Scan(&state, &count); err != nil {
			return nil, err
		}
		counts[state] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, s := range []string{"pendiente_asignacion", "pendiente_avance", "pendiente_evaluacion_final", "en_progreso", "completada"} {
		if _, ok := counts[s]; !ok {
			counts[s] = 0
		}
	}

	return counts, nil
}

// RefreshSummaryView refreshes the materialized view concurrently.
func (r *EvaluationRepo) RefreshSummaryView(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `REFRESH MATERIALIZED VIEW CONCURRENTLY evaluation_summary`)
	return err
}

// BeginTx starts a database transaction.
func (r *EvaluationRepo) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, opts)
}

// getVersion retrieves the current version for an evaluation.
func (r *EvaluationRepo) getVersion(ctx context.Context, evalID uuid.UUID) (int, error) {
	var version int
	err := r.db.QueryRowContext(ctx,
		`SELECT version FROM evaluations WHERE id = $1`, evalID,
	).Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return version, nil
}

// upsertVersion increments the version on the evaluations row.
func (r *EvaluationRepo) upsertVersion(ctx context.Context, tx *sql.Tx, evalID uuid.UUID) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE evaluations SET version = version + 1, updated_at = NOW() WHERE id = $1`,
		evalID,
	)
	return err
}

// rowFromEnt converts an Ent Evaluation to an EvaluationRow.
func rowFromEnt(ev *internal.Evaluation, version int) *EvaluationRow {
	return &EvaluationRow{
		ID:                        ev.ID,
		CreatedAt:                 ev.CreatedAt,
		UpdatedAt:                 ev.UpdatedAt,
		Phase:                     string(ev.Phase),
		State:                     string(ev.State),
		SelfEvaluationCompletedAt: ev.SelfEvaluationCompletedAt,
		RhEvaluationCompletedAt:   ev.RhEvaluationCompletedAt,
		EmployeeID:                ev.EmployeeID,
		CycleID:                   ev.CycleID,
		Version:                   version,
	}
}
