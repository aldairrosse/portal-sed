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

// CreateCycle inserts a new cycle with version=1 and current_phase='asignacion',
// plus the default phase definitions and transitions for it.
func (r *CycleRepo) CreateCycle(ctx context.Context, tx *sql.Tx, year int, orgID uuid.UUID) (*CycleRow, error) {
	now := time.Now()
	cycleID := uuid.New()

	_, err := tx.ExecContext(ctx,
		`INSERT INTO cycles (id, created_at, updated_at, year, current_phase, organization_id, version)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		cycleID, now, now, year, "asignacion", orgID, 1,
	)
	if err != nil {
		return nil, err
	}

	// Create default phase definitions & transitions for this cycle.
	// Canonical 3 phases: asignacion(1) → avance(2) → cierre(3).
	pdAsignacion := uuid.New()
	pdAvance := uuid.New()
	pdCierre := uuid.New()

	for _, def := range []struct {
		id    uuid.UUID
		phase string
		label string
		order int
	}{
		{pdAsignacion, "asignacion", "Asignación", 1},
		{pdAvance, "avance", "Avance", 2},
		{pdCierre, "cierre", "Cierre", 3},
	} {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO phase_definitions (id, created_at, updated_at, phase, label, "order", cycle_id)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			def.id, now, now, def.phase, def.label, def.order, cycleID,
		)
		if err != nil {
			return nil, err
		}
	}

	for _, trans := range []struct {
		fromPhase string
		toPhase   string
		fromID    uuid.UUID
		toID      uuid.UUID
	}{
		{"asignacion", "avance", pdAsignacion, pdAvance},
		{"avance", "cierre", pdAvance, pdCierre},
		{"avance", "asignacion", pdAvance, pdAsignacion},
		{"cierre", "avance", pdCierre, pdAvance},
	} {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO phase_transitions (id, from_phase, to_phase, trigger, created_at, cycle_id, from_phase_id, to_phase_id)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			uuid.New(), trans.fromPhase, trans.toPhase, "manual_rh", now, cycleID, trans.fromID, trans.toID,
		)
		if err != nil {
			return nil, err
		}
	}

	return &CycleRow{
		ID:             cycleID,
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

// GetCurrentPhaseID resolves the phase definition ID for a cycle's current
// phase by joining cycles with phase_definitions on the phase enum.
// Returns sql.ErrNoRows if the cycle has no phase definition matching its
// current_phase.
func (r *CycleRepo) GetCurrentPhaseID(ctx context.Context, cycleID uuid.UUID) (uuid.UUID, error) {
	var phaseID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT pd.id
		 FROM phase_definitions pd
		 JOIN cycles c ON c.id = pd.cycle_id AND c.current_phase = pd.phase
		 WHERE c.id = $1
		 LIMIT 1`,
		cycleID,
	).Scan(&phaseID)
	if err != nil {
		return uuid.Nil, err
	}
	return phaseID, nil
}

// GetPhaseID resolves the phase definition ID for a cycle + phase name.
// "avance" and "medio-anio" fall back to each other when the requested value
// has no phase definition (legacy cycles only created "avance").
func (r *CycleRepo) GetPhaseID(ctx context.Context, cycleID uuid.UUID, phase string) (uuid.UUID, error) {
	var phaseID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT pd.id FROM phase_definitions pd WHERE pd.cycle_id = $1 AND pd.phase = $2 LIMIT 1`,
		cycleID, phase,
	).Scan(&phaseID)
	if err == nil {
		return phaseID, nil
	}
	if err != sql.ErrNoRows {
		return uuid.Nil, err
	}
	fallback := ""
	switch phase {
	case "avance":
		fallback = "medio-anio"
	case "medio-anio":
		fallback = "avance"
	default:
		return uuid.Nil, sql.ErrNoRows
	}
	err = r.db.QueryRowContext(ctx,
		`SELECT pd.id FROM phase_definitions pd WHERE pd.cycle_id = $1 AND pd.phase = $2 LIMIT 1`,
		cycleID, fallback,
	).Scan(&phaseID)
	if err != nil {
		return uuid.Nil, err
	}
	return phaseID, nil
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
// Does NOT touch finished_at: cierre stays active (finished_at IS NULL)
// until the new annual cycle is created in asignacion.
// Returns CONCURRENT_UPDATE error if RowsAffected == 0.
func (r *CycleRepo) UpdatePhase(ctx context.Context, tx *sql.Tx, cycleID uuid.UUID, nextPhase cycle.CurrentPhase, expectedVersion int) error {
	res, err := tx.ExecContext(ctx,
		`UPDATE cycles
		 SET current_phase = $1::phase, version = version + 1, updated_at = NOW()
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

// GetActive returns the most recent cycle ordered by year DESC, created_at DESC.
// Active = latest cycle; finished_at is never used to resolve the active cycle.
// NOTE: global, NOT tenant-scoped (no organization_id filter). Legacy /
// admin-only helper. Evaluations resolve the active cycle per-tenant via
// GetActiveCycleID(ctx, orgID) — use that in multi-tenant paths.
// Frontend divergence: frontend badge "Cerrado" reads finished_at, but backend
// active resolution ignores finished_at; finished_at stays write-dead (no writer
// yet — cierre finaliza solo al crear nuevo ciclo o update manual pendiente).
func (r *CycleRepo) GetActive(ctx context.Context) (*CycleRow, error) {
	row := &CycleRow{}
	var currentPhase string
	var startedAt, finishedAt sql.NullTime
	err := r.db.QueryRowContext(ctx,
		`SELECT id, created_at, updated_at, year, current_phase, started_at, finished_at, organization_id, COALESCE(version, 1)
		 FROM cycles ORDER BY year DESC, created_at DESC LIMIT 1`,
	).Scan(&row.ID, &row.CreatedAt, &row.UpdatedAt, &row.Year, &currentPhase, &startedAt, &finishedAt, &row.OrganizationID, &row.Version)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
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

// GetCurrent returns the cycle for the given year when it exists, otherwise
// the most recent cycle ordered by year DESC, created_at DESC.
// Returns nil, nil when neither exists.
func (r *CycleRepo) GetCurrent(ctx context.Context, orgID uuid.UUID, year int) (*CycleRow, error) {
	rows, err := r.queryCycles(ctx,
		`SELECT id, created_at, updated_at, year, current_phase, started_at, finished_at, organization_id, COALESCE(version, 1) as version
		 FROM cycles WHERE organization_id = $1
		 ORDER BY CASE WHEN year = $2 THEN 0 ELSE 1 END, year DESC, created_at DESC LIMIT 1`,
		orgID, year,
	)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

// GetActiveCycleID finds the active (latest) cycle for an organization,
// ordered by year DESC, created_at DESC.
func (r *CycleRepo) GetActiveCycleID(ctx context.Context, orgID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT id FROM cycles WHERE organization_id = $1 ORDER BY year DESC, created_at DESC LIMIT 1`,
		orgID,
	).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, errors.ErrCycleNotFound
		}
		return uuid.Nil, err
	}
	return id, nil
}

// ReopenCompletedEvaluations reopens completed evaluations for a cycle.
// Sets state='pendiente_avance' only where phase IN ('avance','medio-anio')
// and state='completada', so cierre rows are never overwritten: a revert lands
// the cycle back in 'avance', reopened avance rows become editable while cierre
// data stays intact. 'medio-anio' is the legacy enum value conserved alongside
// 'avance' — both must match so revert reopens legacy evaluations too.
// Returns the number of rows reopened.
func (r *CycleRepo) ReopenCompletedEvaluations(ctx context.Context, tx *sql.Tx, cycleID uuid.UUID) (int64, error) {
	res, err := tx.ExecContext(ctx,
		`UPDATE evaluations SET state='pendiente_avance', updated_at = NOW() WHERE cycle_id=$1 AND state='completada' AND phase IN ('avance','medio-anio')`,
		cycleID,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// BeginTx starts a *sql.Tx for use in transactional operations.
func (r *CycleRepo) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, opts)
}
