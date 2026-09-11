// Package evaluation provides repository-level CRUD and transactional operations
// for Evaluation, EvaluationCompetency, EvaluationGoal, and the 9×9 matrix
// entities. It uses Ent-generated queries for reads and raw SQL for locking.
package evaluation

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
	ErrCodeEvaluationNotFound       pkgerrors.DomainCode = "EVALUATION_NOT_FOUND"
	ErrCodeMatrixNotFound           pkgerrors.DomainCode = "MATRIX_NOT_FOUND"
	ErrCodeEntryNotFound            pkgerrors.DomainCode = "ENTRY_NOT_FOUND"
	ErrCodeEvaluationFinalized      pkgerrors.DomainCode = "EVALUATION_ALREADY_FINALIZED"
	ErrCodeSelfEvalDeadlinePassed   pkgerrors.DomainCode = "SELF_EVAL_DEADLINE_PASSED"
	ErrCodeQuadrantOutOfRange       pkgerrors.DomainCode = "QUADRANT_OUT_OF_RANGE"
	ErrCodeUnauthorizedEvaluator    pkgerrors.DomainCode = "UNAUTHORIZED_EVALUATOR"
	ErrCodeEvaluationProfileMissing pkgerrors.DomainCode = "EVALUATION_PROFILE_NOT_FOUND"
	ErrCodeInvalidCompetencySource  pkgerrors.DomainCode = "INVALID_COMPETENCY_SOURCE"
)

// Sentinel errors.
var (
	ErrEvaluationNotFound       = pkgerrors.NewDomainError(ErrCodeEvaluationNotFound, "The requested evaluation was not found.", nil)
	ErrMatrixNotFound           = pkgerrors.NewDomainError(ErrCodeMatrixNotFound, "The requested 9×9 matrix was not found.", nil)
	ErrEntryNotFound            = pkgerrors.NewDomainError(ErrCodeEntryNotFound, "The requested matrix entry was not found.", nil)
	ErrEvaluationFinalized      = pkgerrors.NewDomainError(ErrCodeEvaluationFinalized, "The evaluation has already been finalized; no further changes allowed.", nil)
	ErrSelfEvalDeadlinePassed   = pkgerrors.NewDomainError(ErrCodeSelfEvalDeadlinePassed, "The self-evaluation deadline has passed for this cycle.", nil)
	ErrQuadrantOutOfRange       = pkgerrors.NewDomainError(ErrCodeQuadrantOutOfRange, "Performance and potential tiers must be between 1 and 3.", nil)
	ErrUnauthorizedEvaluator    = pkgerrors.NewDomainError(ErrCodeUnauthorizedEvaluator, "The authenticated user is not the evaluator for this matrix.", nil)
	ErrEvaluationProfileMissing = pkgerrors.NewDomainError(ErrCodeEvaluationProfileMissing, "The employee has no evaluation profile; cannot rate competencies.", nil)
	ErrInvalidCompetencySource  = pkgerrors.NewDomainError(ErrCodeInvalidCompetencySource, "Competency source must be self or rh.", nil)
)

// CompetencyUpsert is a repository-level DTO for upserting a competency rating.
// IsManager distinguishes the jefe write path on the shared rh row: when true,
// ManagerComment is written to manager_comment and comments is left untouched;
// otherwise Comments is written to comments. Rating is always shared.
type CompetencyUpsert struct {
	CompetencyID   uuid.UUID
	Rating         int
	Comments       string
	ManagerComment string
	IsManager      bool
}

// EmployeeCompetencyRatingRow is a repository-level DTO for competency ratings
// returned from the employee + cycle query. Rows in evaluation_competencies are
// per (evaluation_id, competency_id, source), so self/rh rating+comments are
// aggregated by source; Comments is legacy compat (rh preferred, else self).
type EmployeeCompetencyRatingRow struct {
	CompetencyID    uuid.UUID
	SelfRating      *int
	RhRating        *int
	Comments        *string
	SelfComment     *string
	RhComment       *string
	ManagerComment  *string
	AcceptanceLevel *int
}

// GoalCommentUpsert is a repository-level DTO for updating goal comments.
type GoalCommentUpsert struct {
	GoalID  uuid.UUID
	Comment string
}

// GoalStateUpsert is a repository-level DTO for per-goal state updates.
// Nil pointer fields are left unchanged; non-nil values overwrite.
// FinalProgress is the DIRECT goal value in its own unit (never int(FP*5)).
// Phase selects the snapshot column (avance/medio-anio -> avance_progress,
// cierre/fin-anio -> cierre_progress, unknown defaults to avance_progress). FinalRating is only written when explicit.
type GoalStateUpsert struct {
	GoalID         uuid.UUID
	FinalProgress  *float64
	FinalRating    *int
	Phase          string
	SelfAssessment *string
	RhAssessment   *string
}

// GoalProgressSnapshot carries the per-phase direct-value snapshots.
type GoalProgressSnapshot struct {
	GoalID         uuid.UUID
	AvanceProgress *float64
	CierreProgress *float64
}

// EntryUpsert is a repository-level DTO for upserting a nine-box entry.
type EntryUpsert struct {
	EvaluateeID         uuid.UUID
	PerformanceTier     int // 1–3 (was PerformanceScore 1–9)
	PotentialTier       int // 1–3 (was PotentialScore 1–9)
	Quadrant            int
	Comments            string
	GoalProgressPercent *float64
	SelfRating          *float64
	HrRating            *float64
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
// Optional filters: state (exact match) and phase. "avance" and "medio-anio"
// are interchangeable and match both values (same mid-year phase).
func (r *EvaluationRepo) ListByCycle(ctx context.Context, cycleID uuid.UUID, state string, phase string, cursorStr string, limit int) ([]*EvaluationRow, string, error) {
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

	if phase == "avance" || phase == "medio-anio" {
		query += ` AND e.phase IN ($` + strconv.Itoa(idx) + `, $` + strconv.Itoa(idx+1) + `)`
		args = append(args, "avance", "medio-anio")
		idx += 2
	} else if phase != "" {
		query += ` AND e.phase = $` + strconv.Itoa(idx)
		args = append(args, phase)
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
// Dual-write: when setSelfCompleted, rating is also written to self_rating;
// when setRHCompleted, rating is also written to rh_rating.
func (r *EvaluationRepo) SubmitEval(ctx context.Context, tx *sql.Tx, evalID uuid.UUID, profileID uuid.UUID, comps []CompetencyUpsert, goals []GoalCommentUpsert, newState string, setSelfCompleted, setRHCompleted bool) error {
	// 1. Lock row and validate state
	row, err := r.LockEvalForUpdate(ctx, tx, evalID)
	if err != nil {
		return err
	}
	if row.State == "completada" {
		return ErrEvaluationFinalized
	}
	if profileID == uuid.Nil {
		return ErrEvaluationProfileMissing
	}

	now := time.Now()

	// 2. Bulk upsert competencies with dual-write for self_rating/rh_rating.
	// Rows are per (evaluation_id, competency_id, source) — constraint
	// idx_eval_comp_eval_comp_source — so source is set explicitly from the
	// submit path (self vs rh) and is part of the conflict target.
	source := "rh"
	if setSelfCompleted {
		source = "self"
	}
	if source != "self" && source != "rh" {
		return ErrInvalidCompetencySource
	}
	for _, c := range comps {
		// Jefe path on the shared rh row: rating shared, manager_comment only.
		if c.IsManager {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO evaluation_competencies (id, created_at, updated_at, evaluation_id, competency_id, rating, comments, profile_id, source, manager_comment)
				 VALUES ($1, $2, $3, $4, $5, $6, '', $7, $8, $9)
				 ON CONFLICT (evaluation_id, competency_id, source) DO UPDATE
				 SET rating = EXCLUDED.rating, manager_comment = EXCLUDED.manager_comment, updated_at = EXCLUDED.updated_at`,
				uuid.New(), now, now, evalID, c.CompetencyID, c.Rating, profileID, source, c.ManagerComment,
			)
			if err != nil {
				return err
			}
			continue
		}
		query := `INSERT INTO evaluation_competencies (id, created_at, updated_at, evaluation_id, competency_id, rating, comments, profile_id, source`
		args := []interface{}{uuid.New(), now, now, evalID, c.CompetencyID, c.Rating, c.Comments, profileID, source}
		setClauses := `rating = EXCLUDED.rating, comments = EXCLUDED.comments, updated_at = EXCLUDED.updated_at`
		argIdx := 10

		if setSelfCompleted {
			query += `, self_rating`
			setClauses += `, self_rating = EXCLUDED.self_rating`
		}
		if setRHCompleted {
			query += `, rh_rating`
			setClauses += `, rh_rating = EXCLUDED.rh_rating`
		}

		query += `) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9`
		// Build dynamic placeholders for self_rating/rh_rating
		if setSelfCompleted {
			query += fmt.Sprintf(`, $%d`, argIdx)
			args = append(args, c.Rating)
			argIdx++
		}
		if setRHCompleted {
			query += fmt.Sprintf(`, $%d`, argIdx)
			args = append(args, c.Rating)
			argIdx++
		}
		query += `) ON CONFLICT (evaluation_id, competency_id, source) DO UPDATE SET ` + setClauses

		_, err = tx.ExecContext(ctx, query, args...)
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

// GetCompetencyRatingsByEmployee fetches ALL competencies for an employee's profile in a cycle,
// LEFT JOINed with evaluation_competencies aggregated by source to get self/rh
// rating+comments (or null if not yet evaluated). GROUP BY collapses the two
// source rows per competency into one; manager_comment is source-independent.
func (r *EvaluationRepo) GetCompetencyRatingsByEmployee(ctx context.Context, employeeID, cycleID, profileID uuid.UUID) ([]EmployeeCompetencyRatingRow, error) {
	query := `SELECT c.id AS competency_id,
	       MAX(CASE WHEN ec.source = 'self' THEN ec.rating END) AS self_rating,
	       MAX(CASE WHEN ec.source = 'rh' THEN ec.rating END) AS rh_rating,
	       COALESCE(MAX(CASE WHEN ec.source = 'rh' THEN ec.comments END), MAX(CASE WHEN ec.source = 'self' THEN ec.comments END)) AS comments,
	       MAX(CASE WHEN ec.source = 'self' THEN ec.comments END) AS self_comment,
	       MAX(CASE WHEN ec.source = 'rh' THEN ec.comments END) AS rh_comment,
	       MAX(ec.manager_comment) AS manager_comment,
	       cal.level AS acceptance_level
		FROM competencies c
		LEFT JOIN competency_acceptance_levels cal ON cal.competency_id = c.id AND cal.profile_id = $3
		LEFT JOIN evaluations ev ON ev.employee_id = $1 AND ev.cycle_id = $2
		LEFT JOIN evaluation_competencies ec ON ec.evaluation_id = ev.id AND ec.competency_id = c.id
		GROUP BY c.id, cal.level
		ORDER BY c.id`

	rows, err := r.db.QueryContext(ctx, query, employeeID, cycleID, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []EmployeeCompetencyRatingRow
	for rows.Next() {
		var row EmployeeCompetencyRatingRow
		if err := rows.Scan(&row.CompetencyID, &row.SelfRating, &row.RhRating, &row.Comments, &row.SelfComment, &row.RhComment, &row.ManagerComment, &row.AcceptanceLevel); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if results == nil {
		results = []EmployeeCompetencyRatingRow{}
	}
	return results, nil
}

// GetCompetencyManagerComments returns manager_comment per competency for an
// evaluation (keyed by competency_id string). Missing column (drift without
// regen/migration) yields an empty map and nil error so reads stay safe.
func (r *EvaluationRepo) GetCompetencyManagerComments(ctx context.Context, evalID uuid.UUID) (map[string]string, error) {
	out := map[string]string{}
	if r.db == nil {
		return out, nil
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT competency_id, manager_comment FROM evaluation_competencies WHERE evaluation_id = $1`, evalID)
	if err != nil {
		return out, nil
	}
	defer rows.Close()
	for rows.Next() {
		var compID uuid.UUID
		var mc sql.NullString
		if err := rows.Scan(&compID, &mc); err != nil {
			return out, nil
		}
		if mc.Valid && mc.String != "" {
			out[compID.String()] = mc.String
		}
	}
	return out, nil
}

// CompetencyResultRow is a repository-level DTO for paginated competency results.
type CompetencyResultRow struct {
	ID            uuid.UUID
	Name          string
	ProfileName   string
	SelfRatingAvg *float64
	RHRatingAvg   *float64
	Status        string
}

// ListCompetencyResults returns paginated competency averages grouped by employee.
// Starts from employees table with LEFT JOIN to evaluations so employees without
// evaluation data still appear (status='sin-datos'). Filters: cycleID (required),
// phase (optional; "avance"/"medio-anio" match both — applied in the JOIN ON
// clause so employees without phase data keep status='sin-datos'),
// query (ILIKE on name), managerID (for team scope — filters direct reports),
// offset/limit for pagination. Ordered by last_name, first_name, e.id.
func (r *EvaluationRepo) ListCompetencyResults(ctx context.Context, cycleID uuid.UUID, phase string, query string, managerID *uuid.UUID, offset, limit int) ([]*CompetencyResultRow, error) {
	args := []interface{}{cycleID}
	idx := 2

	phaseJoin := ""
	if phase == "avance" || phase == "medio-anio" {
		phaseJoin = ` AND ev.phase IN ($` + strconv.Itoa(idx) + `, $` + strconv.Itoa(idx+1) + `)`
		args = append(args, "avance", "medio-anio")
		idx += 2
	} else if phase != "" {
		phaseJoin = ` AND ev.phase = $` + strconv.Itoa(idx)
		args = append(args, phase)
		idx++
	}

	baseQuery := `SELECT e.id,
		e.first_name || ' ' || e.last_name AS name,
		COALESCE(ep.name, '') AS profile_name,
		AVG(ec.self_rating) AS self_rating_avg,
		AVG(ec.rh_rating) AS rh_rating_avg,
		CASE
			WHEN AVG(ec.self_rating) IS NOT NULL AND AVG(ec.rh_rating) IS NOT NULL THEN 'completada'
			WHEN AVG(ec.self_rating) IS NOT NULL THEN 'autoevaluacion'
			WHEN AVG(ec.rh_rating) IS NOT NULL THEN 'pendiente'
			ELSE 'sin-datos'
		END AS status
	FROM employees e
	LEFT JOIN evaluation_profiles ep ON ep.id = e.profile_id
	LEFT JOIN evaluations ev ON ev.employee_id = e.id AND ev.cycle_id = $1` + phaseJoin + `
	LEFT JOIN evaluation_competencies ec ON ec.evaluation_id = ev.id
	WHERE e.is_active = true`

	// scope=team: filter by manager_id (direct reports only)
	if managerID != nil {
		baseQuery += ` AND e.manager_id = $` + strconv.Itoa(idx) + ` AND e.id != $` + strconv.Itoa(idx)
		args = append(args, *managerID)
		idx++
	}

	if query != "" {
		baseQuery += ` AND (e.first_name ILIKE $` + strconv.Itoa(idx) + ` OR e.last_name ILIKE $` + strconv.Itoa(idx) + `)`
		args = append(args, "%"+query+"%")
		idx++
	}

	baseQuery += ` GROUP BY e.id, ep.name ORDER BY e.last_name, e.first_name, e.id LIMIT $` + strconv.Itoa(idx) + ` OFFSET $` + strconv.Itoa(idx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*CompetencyResultRow
	for rows.Next() {
		row := &CompetencyResultRow{}
		var selfAvg, rhAvg sql.NullFloat64
		if err := rows.Scan(&row.ID, &row.Name, &row.ProfileName, &selfAvg, &rhAvg, &row.Status); err != nil {
			return nil, err
		}
		if selfAvg.Valid {
			row.SelfRatingAvg = &selfAvg.Float64
		}
		if rhAvg.Valid {
			row.RHRatingAvg = &rhAvg.Float64
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if results == nil {
		results = []*CompetencyResultRow{}
	}
	return results, nil
}

// CountCompetencyResults returns total count of distinct employees matching
// the same filters as ListCompetencyResults (no GROUP BY, no AVG columns).
// The phase filter lives in the JOIN ON clause, mirroring ListCompetencyResults.
func (r *EvaluationRepo) CountCompetencyResults(ctx context.Context, cycleID uuid.UUID, phase string, query string, managerID *uuid.UUID) (int, error) {
	args := []interface{}{cycleID}
	idx := 2

	phaseJoin := ""
	if phase == "avance" || phase == "medio-anio" {
		phaseJoin = ` AND ev.phase IN ($` + strconv.Itoa(idx) + `, $` + strconv.Itoa(idx+1) + `)`
		args = append(args, "avance", "medio-anio")
		idx += 2
	} else if phase != "" {
		phaseJoin = ` AND ev.phase = $` + strconv.Itoa(idx)
		args = append(args, phase)
		idx++
	}

	baseQuery := `SELECT COUNT(DISTINCT e.id)
	FROM employees e
	LEFT JOIN evaluations ev ON ev.employee_id = e.id AND ev.cycle_id = $1` + phaseJoin + `
	LEFT JOIN evaluation_competencies ec ON ec.evaluation_id = ev.id
	WHERE e.is_active = true`

	if managerID != nil {
		baseQuery += ` AND e.manager_id = $` + strconv.Itoa(idx) + ` AND e.id != $` + strconv.Itoa(idx)
		args = append(args, *managerID)
		idx++
	}

	if query != "" {
		baseQuery += ` AND (e.first_name ILIKE $` + strconv.Itoa(idx) + ` OR e.last_name ILIKE $` + strconv.Itoa(idx) + `)`
		args = append(args, "%"+query+"%")
		idx++
	}

	var total int
	err := r.db.QueryRowContext(ctx, baseQuery, args...).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// FindByEmployeeCycle returns the evaluation for an employee+cycle,
// preferring the avance-phase row when several phases exist.
func (r *EvaluationRepo) FindByEmployeeCycle(ctx context.Context, employeeID, cycleID uuid.UUID) (*EvaluationRow, error) {
	ev, err := r.client.Evaluation.Query().
		Where(evaluation.And(
			evaluation.EmployeeID(employeeID),
			evaluation.CycleID(cycleID),
			evaluation.PhaseEQ(evaluation.PhaseAvance),
		)).
		Order(evaluation.ByCreatedAt()).
		First(ctx)
	if err == nil {
		log.Printf("[evalId] found employee=%s cycle=%s eval=%s", employeeID, cycleID, ev.ID)
		return rowFromEnt(ev, ev.Version), nil
	}
	if !internal.IsNotFound(err) {
		return nil, err
	}
	ev, err = r.client.Evaluation.Query().
		Where(evaluation.And(
			evaluation.EmployeeID(employeeID),
			evaluation.CycleID(cycleID),
		)).
		Order(evaluation.ByCreatedAt()).
		First(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, ErrEvaluationNotFound
		}
		return nil, err
	}
	log.Printf("[evalId] found employee=%s cycle=%s eval=%s", employeeID, cycleID, ev.ID)
	return rowFromEnt(ev, ev.Version), nil
}

// EnsureEvaluation finds the evaluation for employee+cycle or creates it.
// phase is the cycle's current phase ("avance"/"medio-anio" or "cierre");
// actorID is used for created_by/updated_by (falls back to employeeID).
func (r *EvaluationRepo) EnsureEvaluation(ctx context.Context, employeeID, cycleID uuid.UUID, phase string, actorID uuid.UUID) (*EvaluationRow, error) {
	if row, err := r.FindByEmployeeCycle(ctx, employeeID, cycleID); err == nil {
		return row, nil
	} else if err != ErrEvaluationNotFound {
		return nil, err
	}
	entPhase := evaluation.PhaseAvance
	entState := evaluation.StatePendienteAvance
	if phase == "cierre" || phase == "fin-anio" {
		entPhase = evaluation.PhaseCierre
		entState = evaluation.StatePendienteEvaluacionFinal
	}
	if actorID == uuid.Nil {
		actorID = employeeID
	}
	ev, err := r.client.Evaluation.Create().
		SetPhase(entPhase).
		SetState(entState).
		SetEmployeeID(employeeID).
		SetCycleID(cycleID).
		SetCreatedBy(actorID).
		SetUpdatedBy(actorID).
		Save(ctx)
	if err != nil {
		// Concurrent creation won the race: re-read the winner.
		if row, ferr := r.FindByEmployeeCycle(ctx, employeeID, cycleID); ferr == nil {
			return row, nil
		}
		return nil, err
	}
	log.Printf("[evalId] upserted employee=%s cycle=%s eval=%s phase=%s", employeeID, cycleID, ev.ID, entPhase)
	return rowFromEnt(ev, ev.Version), nil
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
