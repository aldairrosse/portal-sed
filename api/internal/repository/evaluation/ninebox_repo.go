package evaluation

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/nineboxentry"
	"github.com/sed-evaluacion-desempeno/api/internal/nineboxmatrix"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// EmployeeInfo holds the minimal employee fields needed for DTO enrichment.
type EmployeeInfo struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	ProfileID uuid.UUID
}

// NineBoxRepo provides CRUD operations for NineBoxMatrix and NineBoxEntry.
type NineBoxRepo struct {
	client *internal.Client
	db     *sql.DB
}

// NewNineBoxRepo creates a new NineBoxRepo.
func NewNineBoxRepo(client *internal.Client, db *sql.DB) *NineBoxRepo {
	return &NineBoxRepo{client: client, db: db}
}

// CreateMatrix creates a new 9×9 matrix for an evaluator in a cycle.
func (r *NineBoxRepo) CreateMatrix(ctx context.Context, cycleID, evaluatorID uuid.UUID) (*internal.NineBoxMatrix, error) {
	return r.client.NineBoxMatrix.Create().
		SetCycleID(cycleID).
		SetEvaluatorID(evaluatorID).
		Save(ctx)
}

// CreateMatrixWithPhase creates a new 9×9 matrix for an evaluator in a cycle and phase.
func (r *NineBoxRepo) CreateMatrixWithPhase(ctx context.Context, cycleID, evaluatorID, phaseID uuid.UUID) (*internal.NineBoxMatrix, error) {
	return r.client.NineBoxMatrix.Create().
		SetCycleID(cycleID).
		SetEvaluatorID(evaluatorID).
		SetPhaseID(phaseID).
		Save(ctx)
}

// GetMatrixByID retrieves a matrix by ID with entries preloaded.
func (r *NineBoxRepo) GetMatrixByID(ctx context.Context, id uuid.UUID) (*internal.NineBoxMatrix, error) {
	m, err := r.client.NineBoxMatrix.Query().
		Where(nineboxmatrix.ID(id)).
		WithEntries().
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, ErrMatrixNotFound
		}
		return nil, err
	}
	return m, nil
}

// GetMatrixByPhase retrieves a matrix by cycle + evaluator + phase, with entries preloaded.
func (r *NineBoxRepo) GetMatrixByPhase(ctx context.Context, cycleID, evaluatorID, phaseID uuid.UUID) (*internal.NineBoxMatrix, error) {
	m, err := r.client.NineBoxMatrix.Query().
		Where(
			nineboxmatrix.CycleID(cycleID),
			nineboxmatrix.EvaluatorID(evaluatorID),
			nineboxmatrix.PhaseID(phaseID),
		).
		WithEntries().
		WithPhase().
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, ErrMatrixNotFound
		}
		return nil, err
	}
	return m, nil
}

// ListMatrices returns matrices filtered by cycle, evaluator, and/or phase.
func (r *NineBoxRepo) ListMatrices(ctx context.Context, cycleID, evaluatorID, phaseID uuid.UUID) ([]*internal.NineBoxMatrix, error) {
	q := r.client.NineBoxMatrix.Query()
	if cycleID != uuid.Nil {
		q = q.Where(nineboxmatrix.CycleID(cycleID))
	}
	if evaluatorID != uuid.Nil {
		q = q.Where(nineboxmatrix.EvaluatorID(evaluatorID))
	}
	if phaseID != uuid.Nil {
		q = q.Where(nineboxmatrix.PhaseID(phaseID))
	}
	results, err := q.All(ctx)
	if err != nil {
		return nil, err
	}
	if results == nil {
		return []*internal.NineBoxMatrix{}, nil
	}
	return results, nil
}

// ListMatricesByCycle returns all matrices for a given cycle.
func (r *NineBoxRepo) ListMatricesByCycle(ctx context.Context, cycleID uuid.UUID) ([]*internal.NineBoxMatrix, error) {
	results, err := r.client.NineBoxMatrix.Query().
		Where(nineboxmatrix.CycleID(cycleID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if results == nil {
		return []*internal.NineBoxMatrix{}, nil
	}
	return results, nil
}

// GetMatrixEntries returns all entries for a matrix.
func (r *NineBoxRepo) GetMatrixEntries(ctx context.Context, matrixID uuid.UUID) ([]*internal.NineBoxEntry, error) {
	results, err := r.client.NineBoxEntry.Query().
		Where(nineboxentry.MatrixID(matrixID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if results == nil {
		return []*internal.NineBoxEntry{}, nil
	}
	return results, nil
}

// GetMatrixEntriesByQuadrant returns entries for a matrix filtered by quadrant.
func (r *NineBoxRepo) GetMatrixEntriesByQuadrant(ctx context.Context, matrixID uuid.UUID, quadrant int) ([]*internal.NineBoxEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, created_at, updated_at, matrix_id, evaluatee_id, performance_tier, potential_tier, quadrant, comments
		 FROM nine_box_entries WHERE matrix_id = $1 AND quadrant = $2`,
		matrixID, quadrant,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*internal.NineBoxEntry
	for rows.Next() {
		var e internal.NineBoxEntry
		if err := rows.Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt, &e.MatrixID, &e.EvaluateeID, &e.PerformanceTier, &e.PotentialTier, &e.Quadrant, &e.Comments); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if results == nil {
		return []*internal.NineBoxEntry{}, nil
	}
	return results, nil
}

// GetEmployeesByIDs returns a map of employee ID to EmployeeInfo for the given IDs.
func (r *NineBoxRepo) GetEmployeesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*EmployeeInfo, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]*EmployeeInfo{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}
	query := `SELECT id, first_name, last_name, profile_id FROM employees WHERE id IN (` + strings.Join(placeholders, ",") + `)`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	infoMap := make(map[uuid.UUID]*EmployeeInfo)
	for rows.Next() {
		var id uuid.UUID
		var info EmployeeInfo
		if err := rows.Scan(&id, &info.FirstName, &info.LastName, &info.ProfileID); err != nil {
			return nil, err
		}
		info.ID = id
		infoMap[id] = &info
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return infoMap, nil
}

// GetManagerMapping returns a map of employee ID to manager ID.
// Employees with a NULL manager_id are omitted from the result.
func (r *NineBoxRepo) GetManagerMapping(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]uuid.UUID{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}
	query := `SELECT id, manager_id FROM employees WHERE id IN (` + strings.Join(placeholders, ",") + `) AND manager_id IS NOT NULL`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mapping := make(map[uuid.UUID]uuid.UUID)
	for rows.Next() {
		var employeeID, managerID uuid.UUID
		if err := rows.Scan(&employeeID, &managerID); err != nil {
			return nil, err
		}
		mapping[employeeID] = managerID
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return mapping, nil
}

// UpsertEntry creates or updates a single entry within a transaction.
// Deprecated: Use UpsertEntryByTiers instead.
func (r *NineBoxRepo) UpsertEntry(ctx context.Context, tx *sql.Tx, matrixID uuid.UUID, evaluateeID uuid.UUID, perf, pot int, quadrant int, comments string) (*internal.NineBoxEntry, error) {
	now := time.Now()
	entryID := uuid.New()

	// Lock existing entry if present
	_, err := tx.ExecContext(ctx,
		`SELECT id FROM nine_box_entries WHERE matrix_id = $1 AND evaluatee_id = $2 FOR UPDATE`,
		matrixID, evaluateeID,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Upsert with raw SQL (tier-based columns)
	err = tx.QueryRowContext(ctx,
		`INSERT INTO nine_box_entries (id, created_at, updated_at, matrix_id, evaluatee_id, performance_tier, potential_tier, quadrant, comments)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (matrix_id, evaluatee_id) DO UPDATE
		 SET performance_tier = EXCLUDED.performance_tier,
		     potential_tier = EXCLUDED.potential_tier,
		     quadrant = EXCLUDED.quadrant,
		     comments = EXCLUDED.comments,
		     updated_at = EXCLUDED.updated_at
		 RETURNING id`,
		entryID, now, now, matrixID, evaluateeID, perf, pot, quadrant, comments,
	).Scan(&entryID)
	if err != nil {
		return nil, err
	}

	// Ensure version tracking
	_, err = tx.ExecContext(ctx,
		`INSERT INTO ninebox_entry_versions (entry_id, version, updated_at)
		 VALUES ($1, 0, $2)
		 ON CONFLICT (entry_id) DO UPDATE SET version = ninebox_entry_versions.version + 1, updated_at = $2`,
		entryID, now,
	)
	if err != nil {
		return nil, err
	}

	// Fetch the persisted entry
	return r.getEntryByID(ctx, entryID)
}

// UpsertEntryByTiers creates or updates a single entry using tier values (1–3).
func (r *NineBoxRepo) UpsertEntryByTiers(ctx context.Context, tx *sql.Tx, matrixID uuid.UUID, evaluateeID uuid.UUID, perfTier, potTier, quadrant int, comments string) (*internal.NineBoxEntry, error) {
	return r.UpsertEntry(ctx, tx, matrixID, evaluateeID, perfTier, potTier, quadrant, comments)
}

// UpdateEntry updates an existing entry with optimistic lock.
// Deprecated: Prefer UpsertEntryByTiers via RecomputeMatrix.
func (r *NineBoxRepo) UpdateEntry(ctx context.Context, tx *sql.Tx, entryID uuid.UUID, perf, pot int, quadrant int, comments string, version int) (*internal.NineBoxEntry, error) {
	now := time.Now()

	// Lock and check version
	var currentVersion int
	err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(version, 0) FROM ninebox_entry_versions WHERE entry_id = $1 FOR UPDATE`,
		entryID,
	).Scan(&currentVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrEntryNotFound
		}
		return nil, err
	}

	if currentVersion != version {
		return nil, pkgerrors.ErrConcurrentUpdate
	}

	// Update entry (tier-based columns)
	_, err = tx.ExecContext(ctx,
		`UPDATE nine_box_entries
		 SET performance_tier = $1, potential_tier = $2, quadrant = $3, comments = $4, updated_at = $5
		 WHERE id = $6`,
		perf, pot, quadrant, comments, now, entryID,
	)
	if err != nil {
		return nil, err
	}

	// Increment version
	_, err = tx.ExecContext(ctx,
		`UPDATE ninebox_entry_versions SET version = version + 1, updated_at = $1 WHERE entry_id = $2`,
		now, entryID,
	)
	if err != nil {
		return nil, err
	}

	return r.getEntryByID(ctx, entryID)
}

// BatchUpsertEntries atomically upserts multiple entries within a transaction.
func (r *NineBoxRepo) BatchUpsertEntries(ctx context.Context, tx *sql.Tx, matrixID uuid.UUID, items []EntryUpsert) ([]*internal.NineBoxEntry, error) {
	if len(items) == 0 {
		return []*internal.NineBoxEntry{}, nil
	}

	// Lock all existing entries for this matrix
	_, err := tx.ExecContext(ctx,
		`SELECT id FROM nine_box_entries WHERE matrix_id = $1 FOR UPDATE`,
		matrixID,
	)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	for _, it := range items {
		entryID := uuid.New()
		err := tx.QueryRowContext(ctx,
			`INSERT INTO nine_box_entries (id, created_at, updated_at, matrix_id, evaluatee_id, performance_tier, potential_tier, quadrant, comments)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			 ON CONFLICT (matrix_id, evaluatee_id) DO UPDATE
			 SET performance_tier = EXCLUDED.performance_tier,
			     potential_tier = EXCLUDED.potential_tier,
			     quadrant = EXCLUDED.quadrant,
			     comments = EXCLUDED.comments,
			     updated_at = EXCLUDED.updated_at
			 RETURNING id`,
			entryID, now, now, matrixID, it.EvaluateeID, it.PerformanceTier, it.PotentialTier, it.Quadrant, it.Comments,
		).Scan(&entryID)
		if err != nil {
			return nil, err
		}

		_, err = tx.ExecContext(ctx,
			`INSERT INTO ninebox_entry_versions (entry_id, version, updated_at)
			 VALUES ($1, 0, $2)
			 ON CONFLICT (entry_id) DO UPDATE SET version = ninebox_entry_versions.version + 1, updated_at = $2`,
			entryID, now,
		)
		if err != nil {
			return nil, err
		}
	}

	// Re-fetch all entries for this matrix
	return r.GetMatrixEntries(ctx, matrixID)
}

// LockEntryForSelect locks an existing entry for update.
func (r *NineBoxRepo) LockEntryForSelect(ctx context.Context, tx *sql.Tx, matrixID, evaluateeID uuid.UUID) error {
	_, err := tx.ExecContext(ctx,
		`SELECT id FROM nine_box_entries WHERE matrix_id = $1 AND evaluatee_id = $2 FOR UPDATE`,
		matrixID, evaluateeID,
	)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	return nil
}

// FetchEntryVersion retrieves the version for an entry.
func (r *NineBoxRepo) FetchEntryVersion(ctx context.Context, entryID uuid.UUID) (int, error) {
	var version int
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(version, 0) FROM ninebox_entry_versions WHERE entry_id = $1`, entryID,
	).Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return version, nil
}

// ---- queries for RecomputeMatrix ----

// GetGoalAssigneesByCycle returns all employee IDs that have goals assigned in a given cycle.
// This is used to identify evaluatees for nine-box computation.
func (r *NineBoxRepo) GetGoalAssigneesByCycle(ctx context.Context, cycleID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT DISTINCT employee_id FROM goal_assignments WHERE cycle_id = $1`,
		cycleID,
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
	if ids == nil {
		return []uuid.UUID{}, nil
	}
	return ids, rows.Err()
}

// GetGoalProgressByEmployee returns the average goal progress (0–100) for an employee.
// Progress is calculated as: CASE WHEN direction='descendente' THEN (baseline-current)/baseline*100 ELSE (current/target)*100 END
// Only goals with current_value > 0 are considered (goals with measurable progress).
func (r *NineBoxRepo) GetGoalProgressByEmployee(ctx context.Context, employeeID, cycleID uuid.UUID) (float64, error) {
	var avgProgress sql.NullFloat64
	err := r.db.QueryRowContext(ctx, `
		SELECT AVG(
			CASE
				WHEN g.direction = 'descendente' AND g.baseline_value IS NOT NULL AND g.baseline_value > 0
					THEN GREATEST(0, (g.baseline_value - g.current_value) / g.baseline_value * 100)
				ELSE
					CASE WHEN g.target_value > 0
						THEN LEAST(100, g.current_value / g.target_value * 100)
						ELSE 0
					END
			END
		)
		FROM goals g
		JOIN goal_categories gc ON g.category_id = gc.id
		JOIN goal_assignments ga ON ga.employee_id = $1 AND ga.cycle_id = $2
		WHERE gc.employee_id = ga.employee_id
		  AND g.current_value > 0
	`, employeeID, cycleID).Scan(&avgProgress)
	if err != nil {
		return 0, err
	}
	if !avgProgress.Valid {
		return 0, nil
	}
	return avgProgress.Float64, nil
}

// GetCompetencyRatingsByEmployee returns the average competency rating for an employee's evaluation in a cycle.
// Returns self and HR ratings separately (or nil if not available).
func (r *NineBoxRepo) GetCompetencyRatingsByEmployee(ctx context.Context, employeeID, cycleID uuid.UUID) (selfRating, hrRating *float64, err error) {
	// Get the evaluation for this employee in this cycle
	var evalID uuid.UUID
	var selfCompleted, rhCompleted *time.Time
	err = r.db.QueryRowContext(ctx,
		`SELECT id, self_evaluation_completed_at, rh_evaluation_completed_at FROM evaluations WHERE employee_id = $1 AND cycle_id = $2`,
		employeeID, cycleID,
	).Scan(&evalID, &selfCompleted, &rhCompleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil // no evaluation → no ratings
		}
		return nil, nil, err
	}

	// Get all competency ratings for this evaluation
	rows, err := r.db.QueryContext(ctx,
		`SELECT rating, created_at FROM evaluation_competencies WHERE evaluation_id = $1 ORDER BY created_at`,
		evalID,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var selfSum, hrSum float64
	var selfCount, hrCount int

	for rows.Next() {
		var rating int
		var createdAt time.Time
		if err := rows.Scan(&rating, &createdAt); err != nil {
			return nil, nil, err
		}

		// Heuristic: ratings created after self_evaluation_completed_at are self evaluations
		// ratings created after rh_evaluation_completed_at are RH evaluations
		// If only one is completed, all ratings go to that bucket
		// If neither is completed, all ratings are treated as a single pool (returned as hrRating)
		if selfCompleted != nil && createdAt.After(*selfCompleted) && (rhCompleted == nil || createdAt.Before(*rhCompleted)) {
			selfSum += float64(rating)
			selfCount++
		} else if rhCompleted != nil && createdAt.After(*rhCompleted) {
			hrSum += float64(rating)
			hrCount++
		} else if rhCompleted != nil && selfCompleted != nil && createdAt.After(*selfCompleted) {
			// After self but before RH → treat as self
			selfSum += float64(rating)
			selfCount++
		} else if rhCompleted != nil {
			hrSum += float64(rating)
			hrCount++
		} else {
			// Default: treat as HR ratings
			hrSum += float64(rating)
			hrCount++
		}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	if selfCount > 0 {
		v := selfSum / float64(selfCount)
		selfRating = &v
	}
	if hrCount > 0 {
		v := hrSum / float64(hrCount)
		hrRating = &v
	}

	return selfRating, hrRating, nil
}

// getEntryByID fetches a single entry by ID.
func (r *NineBoxRepo) getEntryByID(ctx context.Context, entryID uuid.UUID) (*internal.NineBoxEntry, error) {
	entry, err := r.client.NineBoxEntry.Query().
		Where(nineboxentry.ID(entryID)).
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, ErrEntryNotFound
		}
		return nil, err
	}
	return entry, nil
}
