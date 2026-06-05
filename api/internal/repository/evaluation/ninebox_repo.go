package evaluation

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/nineboxentry"
	"github.com/sed-evaluacion-desempeno/api/internal/nineboxmatrix"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

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

// ListMatrices returns matrices filtered by cycle and/or evaluator.
func (r *NineBoxRepo) ListMatrices(ctx context.Context, cycleID, evaluatorID uuid.UUID) ([]*internal.NineBoxMatrix, error) {
	q := r.client.NineBoxMatrix.Query()
	if cycleID != uuid.Nil {
		q = q.Where(nineboxmatrix.CycleID(cycleID))
	}
	if evaluatorID != uuid.Nil {
		q = q.Where(nineboxmatrix.EvaluatorID(evaluatorID))
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

// UpsertEntry creates or updates a single entry within a transaction.
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

	// Upsert with raw SQL
	err = tx.QueryRowContext(ctx,
		`INSERT INTO nine_box_entries (id, created_at, updated_at, matrix_id, evaluatee_id, performance_score, potential_score, quadrant, comments)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (matrix_id, evaluatee_id) DO UPDATE
		 SET performance_score = EXCLUDED.performance_score,
		     potential_score = EXCLUDED.potential_score,
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

// UpdateEntry updates an existing entry with optimistic lock check.
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

	// Update entry
	_, err = tx.ExecContext(ctx,
		`UPDATE nine_box_entries
		 SET performance_score = $1, potential_score = $2, quadrant = $3, comments = $4, updated_at = $5
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
			`INSERT INTO nine_box_entries (id, created_at, updated_at, matrix_id, evaluatee_id, performance_score, potential_score, quadrant, comments)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			 ON CONFLICT (matrix_id, evaluatee_id) DO UPDATE
			 SET performance_score = EXCLUDED.performance_score,
			     potential_score = EXCLUDED.potential_score,
			     quadrant = EXCLUDED.quadrant,
			     comments = EXCLUDED.comments,
			     updated_at = EXCLUDED.updated_at
			 RETURNING id`,
			entryID, now, now, matrixID, it.EvaluateeID, it.PerformanceScore, it.PotentialScore, it.Quadrant, it.Comments,
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
