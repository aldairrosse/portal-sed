// Package cycle provides the repository layer for Cycle entities.
// It uses Ent-generated queries for standard CRUD and raw SQL for operations
// involving the version field (optimistic locking).
package cycle

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/cycle"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// contextKey for db role routing.
type ctxKeyDBRole struct{}

const (
	// DBRolePrimary routes queries to the primary database.
	DBRolePrimary = "primary"
	// DBRoleReplica routes queries to a read replica.
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

// CycleRow is a full representation of a cycle including the version field.
type CycleRow struct {
	ID             uuid.UUID          `json:"id"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	Year           int                `json:"year"`
	CurrentPhase   cycle.CurrentPhase `json:"current_phase"`
	StartedAt      *time.Time         `json:"started_at,omitempty"`
	FinishedAt     *time.Time         `json:"finished_at,omitempty"`
	OrganizationID uuid.UUID          `json:"organization_id"`
	Version        int                `json:"version"`
}

// ToCycle converts a CycleRow to the generated Cycle model.
func (r *CycleRow) ToCycle() *internal.Cycle {
	return &internal.Cycle{
		ID:             r.ID,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
		Year:           r.Year,
		CurrentPhase:   r.CurrentPhase,
		StartedAt:      r.StartedAt,
		FinishedAt:     r.FinishedAt,
		OrganizationID: r.OrganizationID,
	}
}

// CycleRepo provides Ent-backed CRUD operations for cycles with raw SQL
// fallback for version-based optimistic locking.
type CycleRepo struct {
	client *internal.Client
	db     *sql.DB
}

// NewCycleRepo creates a new CycleRepo.
func NewCycleRepo(client *internal.Client, db *sql.DB) *CycleRepo {
	return &CycleRepo{client: client, db: db}
}

// clientFor returns the appropriate client based on context db role hint.
func (r *CycleRepo) clientFor(ctx context.Context) *internal.Client {
	return r.client
}

// CreateCycle inserts a new cycle with version=1 and current_phase='asignacion'.
// Uses raw SQL to include the version field.
func (r *CycleRepo) CreateCycle(ctx context.Context, tx *sql.Tx, year int, orgID uuid.UUID) (*CycleRow, error) {
	now := time.Now()
	id := uuid.New()

	_, err := tx.ExecContext(ctx,
		`INSERT INTO cycles (id, created_at, updated_at, year, current_phase, organization_id, version)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, now, now, year, "asignacion", orgID, 1,
	)
	if err != nil {
		return nil, err
	}

	return &CycleRow{
		ID:             id,
		CreatedAt:      now,
		UpdatedAt:      now,
		Year:           year,
		CurrentPhase:   cycle.CurrentPhaseAsignacion,
		OrganizationID: orgID,
		Version:        1,
	}, nil
}

// GetCycle retrieves a cycle by ID using raw SQL for full control including
// the version field.
func (r *CycleRepo) GetCycle(ctx context.Context, id uuid.UUID) (*CycleRow, error) {
	row := &CycleRow{}
	var currentPhase string
	var startedAt, finishedAt sql.NullTime

	err := r.db.QueryRowContext(ctx,
		`SELECT id, created_at, updated_at, year, current_phase, started_at, finished_at, organization_id, COALESCE(version, 1)
		 FROM cycles WHERE id = $1`, id,
	).Scan(&row.ID, &row.CreatedAt, &row.UpdatedAt, &row.Year,
		&currentPhase, &startedAt, &finishedAt, &row.OrganizationID, &row.Version)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrCycleNotFound
		}
		return nil, err
	}

	row.CurrentPhase = cycle.CurrentPhase(currentPhase)
	if startedAt.Valid {
		row.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		row.FinishedAt = &finishedAt.Time
	}

	return row, nil
}

// ListCycles returns cycles for an org, ordered by updated_at DESC, id DESC,
// with cursor-based pagination. Uses raw SQL for full ordering control.
func (r *CycleRepo) ListCycles(ctx context.Context, orgID uuid.UUID, year *int, phase *cycle.CurrentPhase, cursorID *uuid.UUID, cursorUpdatedAt *time.Time, limit int) ([]*CycleRow, error) {
	query := `SELECT id, created_at, updated_at, year, current_phase, started_at, finished_at, organization_id, COALESCE(version, 1) as version
	           FROM cycles WHERE organization_id = $1`
	args := []interface{}{orgID}
	idx := 2

	if year != nil {
		query += ` AND year = $` + strconv.Itoa(idx)
		args = append(args, *year)
		idx++
	}
	if phase != nil {
		query += ` AND current_phase = $` + strconv.Itoa(idx)
		args = append(args, string(*phase))
		idx++
	}
	if cursorID != nil && cursorUpdatedAt != nil {
		query += ` AND (updated_at, id) < ($` + strconv.Itoa(idx) + `, $` + strconv.Itoa(idx+1) + `)`
		args = append(args, *cursorUpdatedAt, *cursorID)
		idx += 2
	}

	query += ` ORDER BY updated_at DESC, id DESC LIMIT $` + strconv.Itoa(idx)
	args = append(args, limit+1)

	return r.queryCycles(ctx, query, args...)
}

// UpdatePhase applies the phase transition using an optimistic-lock UPDATE.
// Uses raw SQL for atomic version check. Expects a *sql.Tx.
// Returns CONCURRENT_UPDATE error if RowsAffected == 0.
func (r *CycleRepo) UpdatePhase(ctx context.Context, tx *sql.Tx, cycleID uuid.UUID, nextPhase cycle.CurrentPhase, expectedVersion int) error {
	res, err := tx.ExecContext(ctx,
		`UPDATE cycles
		 SET current_phase = $1, version = version + 1, updated_at = NOW()
		 WHERE id = $2 AND version = $3`,
		string(nextPhase), cycleID, expectedVersion,
	)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return errors.ErrConcurrentUpdate
	}
	return nil
}

// InsertPhaseHistory inserts an audit log entry for a phase transition.
func (r *CycleRepo) InsertPhaseHistory(ctx context.Context, tx *sql.Tx, cycleID uuid.UUID, fromPhase, toPhase string, triggeredBy uuid.UUID, reason string) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO cycle_phase_history (id, cycle_id, from_phase, to_phase, triggered_by, reason, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		uuid.New(), cycleID, fromPhase, toPhase, triggeredBy, reason, time.Now(),
	)
	return err
}

// CheckExistingCycle checks if a cycle exists for the given org and year.
func (r *CycleRepo) CheckExistingCycle(ctx context.Context, tx *sql.Tx, orgID uuid.UUID, year int) (bool, error) {
	var count int
	err := tx.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM cycles WHERE organization_id = $1 AND year = $2`,
		orgID, year,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// LockCycleForUpdate locks the cycle row with SELECT FOR UPDATE.
// Returns the full CycleRow including version.
func (r *CycleRepo) LockCycleForUpdate(ctx context.Context, tx *sql.Tx, cycleID uuid.UUID) (*CycleRow, error) {
	row := &CycleRow{}
	var currentPhase string
	var startedAt, finishedAt sql.NullTime

	err := tx.QueryRowContext(ctx,
		`SELECT id, created_at, updated_at, year, current_phase, started_at, finished_at, organization_id, COALESCE(version, 1)
		 FROM cycles WHERE id = $1 FOR UPDATE`,
		cycleID,
	).Scan(&row.ID, &row.CreatedAt, &row.UpdatedAt, &row.Year,
		&currentPhase, &startedAt, &finishedAt, &row.OrganizationID, &row.Version)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrCycleNotFound
		}
		return nil, err
	}

	row.CurrentPhase = cycle.CurrentPhase(currentPhase)
	if startedAt.Valid {
		row.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		row.FinishedAt = &finishedAt.Time
	}

	return row, nil
}

// fetchVersion retrieves the version field for a cycle.
func (r *CycleRepo) fetchVersion(ctx context.Context, id uuid.UUID) (int, error) {
	var version int
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(version, 1) FROM cycles WHERE id = $1`, id,
	).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// queryCycles runs a raw SQL query and scans results into CycleRow values.
func (r *CycleRepo) queryCycles(ctx context.Context, query string, args ...interface{}) ([]*CycleRow, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*CycleRow
	for rows.Next() {
		row := &CycleRow{}
		var currentPhase string
		var startedAt, finishedAt sql.NullTime

		err := rows.Scan(&row.ID, &row.CreatedAt, &row.UpdatedAt, &row.Year,
			&currentPhase, &startedAt, &finishedAt, &row.OrganizationID, &row.Version)
		if err != nil {
			return nil, err
		}
		row.CurrentPhase = cycle.CurrentPhase(currentPhase)
		if startedAt.Valid {
			row.StartedAt = &startedAt.Time
		}
		if finishedAt.Valid {
			row.FinishedAt = &finishedAt.Time
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// ExecuteRawAdvisoryLock acquires a PostgreSQL advisory lock and returns a
// function to release it. Uses hashtext('cycle:create:<orgID>:<year>') as key.
func (r *CycleRepo) ExecuteRawAdvisoryLock(ctx context.Context, orgID uuid.UUID, year int) (func() error, error) {
	lockKey := "cycle:create:" + orgID.String() + ":" + strconv.Itoa(year)
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}

	_, err = conn.ExecContext(ctx, `SELECT pg_advisory_lock(hashtext($1))`, lockKey)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return func() error {
		defer conn.Close()
		_, err := conn.ExecContext(ctx, `SELECT pg_advisory_unlock(hashtext($1))`, lockKey)
		return err
	}, nil
}

// BeginTx starts a *sql.Tx for use in transactional operations.
func (r *CycleRepo) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, opts)
}
